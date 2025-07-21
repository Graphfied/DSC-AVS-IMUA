package challenge

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	sdklogging "github.com/imua-xyz/imua-avs-sdk/logging"
	"github.com/imua-xyz/imua-avs-sdk/nodeapi"
	"github.com/imua-xyz/imua-avs-sdk/signer"
	avs "github.com/imua-xyz/imua-avs/contracts/bindings/avs"
	chain "github.com/imua-xyz/imua-avs/core/chainio"
	"github.com/imua-xyz/imua-avs/core/chainio/eth"
	"github.com/imua-xyz/imua-avs/types"
)

const (
	AvsName = "hello-world-avs-demo"
	SemVer  = "0.0.1"
	// DayEpochID defines the identifier for a daily epoch.
)

type Challenger struct {
	config          types.NodeConfig
	logger          sdklogging.Logger
	ethClient       eth.EthClient
	ethWsClient     *ethclient.Client
	nodeApi         *nodeapi.NodeApi
	avsWriter       chain.AvsWriter
	avsReader       chain.ChainReader
	avsAddr         common.Address
	epochIdentifier string
	contractABI     abi.ABI
}

func NewChallengeFromConfig(c types.NodeConfig) (*Challenger, error) {
	var logLevel sdklogging.LogLevel
	if c.Production {
		logLevel = sdklogging.Production
	} else {
		logLevel = sdklogging.Development
	}
	logger, err := sdklogging.NewZapLogger(logLevel)
	if err != nil {
		return nil, err
	}

	// Setup Node Api
	nodeApi := nodeapi.NewNodeApi(AvsName, SemVer, c.NodeApiIpPortAddress, logger)

	var ethRpcClient eth.EthClient
	ethRpcClient, err = eth.NewClient(c.EthRpcUrl)
	if err != nil {
		logger.Error("can not create http eth client", "err", err)

		return nil, err
	}
	ethWsClient, err := ethclient.Dial(c.EthWsUrl)
	if err != nil {
		logger.Error("Cannot create ws eth client", "err", err)
		return nil, err
	}

	chainId, err := ethRpcClient.ChainID(context.Background())
	if err != nil {
		logger.Error("Cannot get chainId", "err", err)
		return nil, err
	}

	ecdsaKeyPassword, ok := os.LookupEnv("AVS_ECDSA_KEY_PASSWORD")
	if !ok {
		logger.Info("AVS_ECDSA_KEY_PASSWORD env var not set. using empty string")
	}

	signer, challengeSender, err := signer.SignerFromConfig(signer.Config{
		KeystorePath: c.AVSEcdsaPrivateKeyStorePath,
		Password:     ecdsaKeyPassword,
	}, chainId)
	if err != nil {
		panic(err)
	}
	logger.Info("challengeSender:", "challengeSender", challengeSender.String())

	balance, err := ethRpcClient.BalanceAt(context.Background(), challengeSender, nil)
	if err != nil {
		logger.Error("Cannot get Balance", "err", err)
	}
	if balance.Cmp(big.NewInt(0)) != 1 {
		logger.Error("challengeSender has not enough Balance")
	}
	if c.AVSOwnerAddress != challengeSender.String() {
		logger.Error("challengeSender is not equal AVSOwnerAddress")
	}
	txMgr := txmgr.NewSimpleTxManager(ethRpcClient, logger, signer, common.HexToAddress(c.AVSOwnerAddress))

	avsReader, _ := chain.BuildChainReader(
		common.HexToAddress(c.AVSAddress),
		ethRpcClient,
		logger)

	avsWriter, _ := chain.BuildChainWriter(
		common.HexToAddress(c.AVSAddress),
		ethRpcClient,
		logger,
		txMgr)

	if err != nil {
		logger.Error("Cannot create AvsSubscriber", "err", err)
		return nil, err
	}
	epochIdentifier, err := avsReader.GetAVSEpochIdentifier(&bind.CallOpts{}, c.AVSAddress)
	if err != nil {
		logger.Error("Cannot GetAVSEpochIdentifier", "err", err)
		return nil, err
	}
	contractABI, err := avs.ContracthelloWorldMetaData.GetAbi()
	if err != nil {
		logger.Error("Cannot GetAbi", "err", err)
		return nil, err
	}
	challenger := &Challenger{
		config:          c,
		logger:          logger,
		nodeApi:         nodeApi,
		ethClient:       ethRpcClient,
		ethWsClient:     ethWsClient,
		avsWriter:       avsWriter,
		avsReader:       *avsReader,
		avsAddr:         common.HexToAddress(c.AVSAddress),
		epochIdentifier: epochIdentifier,
		contractABI:     *contractABI,
	}
	logger.Info("challenger info", "challengeAddr", c.AVSOwnerAddress)

	return challenger, nil
}
func (o *Challenger) Exec(ctx context.Context, taskID, num uint64) error {
	taskInfo, _ := o.avsReader.GetTaskInfo(&bind.CallOpts{}, o.avsAddr.String(), taskID)
	infos, _ := o.avsReader.GetOperatorTaskResponseList(&bind.CallOpts{}, taskInfo.TaskContractAddress.String(), taskInfo.TaskID)
	task := &avs.AvsServiceContractChallengeReq{
		TaskId:            taskInfo.TaskID,
		TaskAddress:       taskInfo.TaskContractAddress,
		NumberToBeSquared: num,
		Infos:             infos,
		SignedOperators:   taskInfo.SignedOperators,
		NoSignedOperators: taskInfo.NoSignedOperators,
		TaskTotalPower:    taskInfo.TaskTotalPower,
	}
	o.logger.Info("challenger info", "challenge-TaskResponse", taskInfo)
	_, err := o.avsWriter.Challenge(
		ctx,
		*task)

	if err != nil {
		o.logger.Error("Challeger failed to raiseAndResolveChallenge", "err", err)
		return fmt.Errorf("failed to raiseAndResolveChallenge: %w", err)
	}
	return nil
}
func (o *Challenger) Start(ctx context.Context) error {
	// 1. First, the task reaches the challenge period and the module is verified
	// 2. Is an effective task:
	// Has not been processed, query kv to make sure it has not been processed module validation
	// At the same time, optInOperators is not empty, which is ensured when creating, and avs usd will be required not to be 0 module verification when creating
	// 3. Get the array string optInOperators, iterate through it, query and verify that the submitted task respond of each Operator is equal to the formula,
	// If you validate the total passed power by putting it into the array string[] eligibleRewardOperators and getting the data from operatorActivePower, Do not verify by putting in the array string[] eligibleSlashOperators;
	// 4. Calculate the proportion of passed power in the total taskTotalPower
	// 5. Compare whether the above proportion is greater than or equal to thresholdPercentage. If it is greater than, isExpected is true; otherwise, it is false
	o.logger.Infof("Starting Challenge.")
	if o.config.EnableNodeApi {
		o.nodeApi.Start()
	}
	query := ethereum.FilterQuery{
		Addresses: []common.Address{o.avsAddr},
	}
	logs := make(chan ethtypes.Log)

	sub, err := o.ethWsClient.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		o.logger.Error("Subscribe failed", "err", err)

	}
	defer sub.Unsubscribe()

	o.logger.Infof("Starting event monitoring...")

	for {
		select {
		case err := <-sub.Err():
			o.logger.Error("Subscription error:", err)
		case vLog := <-logs:
			event, err := o.parseEvent(vLog)
			if err != nil {
				o.logger.Info("Not as expected TaskCreated log, parse err:", "err", err)
			}
			if event != nil {
				e := event.(*avs.ContracthelloWorldTaskCreated)
				task := o.ProcessNewTaskCreatedLog(e)
				taskInfo, err := o.avsReader.GetTaskInfo(&bind.CallOpts{}, o.avsAddr.String(), task.TaskId)
				if err != nil {
					o.logger.Error("Failed to GetTaskInfo", "err", err)
					return err
				}
				go func() {
					_, err := o.TriggerChallenge(context.Background(), *task, taskInfo)
					if err != nil {

					}
				}()
			}
		}
	}
}

// ProcessNewTaskCreatedLog TaskResponse is the struct that is signed and sent to the chain as a task response.
func (o *Challenger) ProcessNewTaskCreatedLog(e *avs.ContracthelloWorldTaskCreated) *avs.AvsServiceContractChallengeReq {
	o.logger.Info("New Task Created", "TaskID", e.TaskId.Uint64(),
		"Issuer", e.Issuer.String(), "Name", e.Name, "NumberToBeSquared", e.NumberToBeSquared)
	task := &avs.AvsServiceContractChallengeReq{
		TaskId:            e.TaskId.Uint64(),
		NumberToBeSquared: e.NumberToBeSquared,
	}
	return task
}

func (o *Challenger) TriggerChallenge(
	ctx context.Context,
	task avs.AvsServiceContractChallengeReq,
	taskInfo avs.TaskInfo) (string, error) {
	o.logger.Info("TriggerChallenge", "taskInfo", taskInfo)
	epochIdentifier, err := o.avsReader.GetAVSEpochIdentifier(&bind.CallOpts{}, o.avsAddr.String())
	startingEpoch := taskInfo.StartingEpoch
	taskResponsePeriod := taskInfo.TaskResponsePeriod
	taskStatisticalPeriod := taskInfo.TaskStatisticalPeriod
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err() // Gracefully exit if context is canceled
		default:
			// Fetch the current epoch information
			if err != nil {
				o.logger.Error("Cannot GetAVSEpochIdentifier", "err", err)
				return "", fmt.Errorf("failed to get AVS info: %w", err) // Stop on persistent error
			}

			num, err := o.avsReader.GetCurrentEpoch(&bind.CallOpts{}, epochIdentifier)
			if err != nil {
				o.logger.Error("Cannot exec GetCurrentEpoch", "err", err)
				return "", fmt.Errorf("failed to get current epoch: %w", err) // Stop on persistent error
			}

			currentEpoch := uint64(num)
			// o.logger.Info("current epoch  is :", "currentEpoch", currentEpoch)
			if currentEpoch > startingEpoch+taskResponsePeriod+taskStatisticalPeriod {
				// ReQuery the latest taskInfo
				taskInfo, err = o.avsReader.GetTaskInfo(&bind.CallOpts{}, taskInfo.TaskContractAddress.String(), taskInfo.TaskID)
				if err != nil {
					o.logger.Error("Failed to GetTaskInfo", "err", err)
					return "", nil
				}
				o.logger.Info("latest-taskInfo", "taskInfo", taskInfo)

				infos, _ := o.avsReader.GetOperatorTaskResponseList(&bind.CallOpts{}, taskInfo.TaskContractAddress.String(), taskInfo.TaskID)
				task.TaskAddress = taskInfo.TaskContractAddress
				task.Infos = infos
				task.SignedOperators = taskInfo.SignedOperators
				task.NoSignedOperators = taskInfo.NoSignedOperators
				task.TaskTotalPower = taskInfo.TaskTotalPower

				if taskInfo.IsExpected {
					o.logger.Infof("Task %d is expected. Skipping challenge", task.TaskId)
					return "", nil
				}
				if len(taskInfo.OptInOperators) < 1 {
					o.logger.Infof("Task %d does not have any optIn operators. Skipping challenge", task.TaskId)
					return "", nil
				}

				o.logger.Info("Execute raiseAndResolveChallenge", "currentEpoch", currentEpoch,
					"startingEpoch", startingEpoch, "taskResponsePeriod", taskResponsePeriod, "taskStatisticalPeriod", taskStatisticalPeriod)
				o.logger.Info("Challenge-task-req", "task", task)

				_, err := o.avsWriter.Challenge(
					ctx,
					task)
				if err != nil {
					o.logger.Error("Challenger failed to raiseAndResolveChallenge", "err", err)
					return "", fmt.Errorf("failed to raiseAndResolveChallenge: %w", err)
				}
				o.logger.Infof("The current task %s has been challenged:",
					taskInfo.TaskContractAddress.String()+"--"+strconv.FormatUint(taskInfo.TaskID, 10))
				return "The current task has been challenged .", nil
			}
		}
	}
}
func (o *Challenger) parseEvent(vLog ethtypes.Log) (interface{}, error) {

	vLog.Topics[0] = o.contractABI.Events["TaskCreated"].ID

	var event avs.ContracthelloWorldTaskCreated

	err := o.contractABI.UnpackIntoInterface(&event, "TaskCreated", vLog.Data)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
