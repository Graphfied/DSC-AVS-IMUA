package operator

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	"github.com/imua-xyz/imua-avs-sdk/crypto/bls"
	sdklogging "github.com/imua-xyz/imua-avs-sdk/logging"
	"github.com/imua-xyz/imua-avs-sdk/nodeapi"
	"github.com/imua-xyz/imua-avs-sdk/signer"
	avs "github.com/imua-xyz/imua-avs/contracts/bindings/avs"
	"github.com/imua-xyz/imua-avs/core"
	chain "github.com/imua-xyz/imua-avs/core/chainio"
	"github.com/imua-xyz/imua-avs/core/chainio/eth"
	"github.com/imua-xyz/imua-avs/types"
	blscommon "github.com/prysmaticlabs/prysm/v5/crypto/bls/common"
)

const (
	AvsName    = "hello-world-avs-demo"
	SemVer     = "0.0.1"
	maxRetries = 80
	retryDelay = 1 * time.Second
)

type Operator struct {
	config      types.NodeConfig
	logger      sdklogging.Logger
	ethClient   eth.EthClient
	nodeApi     *nodeapi.NodeApi
	avsWriter   chain.AvsWriter
	avsReader   chain.ChainReader
	ethWsClient *ethclient.Client

	blsKeypair   blscommon.SecretKey
	operatorAddr common.Address
	// receive new tasks in this chan (typically from listening to onchain event)
	newTaskCreatedChan chan *avs.ContracthelloWorldTaskCreated
	// needed when opting in to avs (allow this service manager contract to slash operator)
	avsAddr         common.Address
	epochIdentifier string
	contractABI     abi.ABI
}

func NewOperatorFromConfig(c types.NodeConfig) (*Operator, error) {
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

	blsKeyPassword, ok := os.LookupEnv("OPERATOR_BLS_KEY_PASSWORD")
	if !ok {
		logger.Info("OPERATOR_BLS_KEY_PASSWORD env var not set. using empty string")
	}
	blsKeyPair, err := bls.ReadPrivateKeyFromFile(c.BlsPrivateKeyStorePath, blsKeyPassword)
	if err != nil {
		logger.Error("Cannot parse bls private key", "err", err)
		return nil, err
	}

	chainId, err := ethRpcClient.ChainID(context.Background())
	if err != nil {
		logger.Error("Cannot get chainId", "err", err)
		return nil, err
	}

	ecdsaKeyPassword, ok := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
	if !ok {
		logger.Info("OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string")
	}

	signer, operatorSender, err := signer.SignerFromConfig(signer.Config{
		KeystorePath: c.OperatorEcdsaPrivateKeyStorePath,
		Password:     ecdsaKeyPassword,
	}, chainId)
	if err != nil {
		panic(err)
	}
	logger.Info("operatorSender:", "operatorSender", operatorSender.String())

	balance, err := ethRpcClient.BalanceAt(context.Background(), operatorSender, nil)
	if err != nil {
		logger.Error("Cannot get Balance", "err", err)
	}
	if balance.Cmp(big.NewInt(0)) != 1 {
		logger.Error("operatorSender has not enough Balance")
	}
	if c.OperatorAddress != operatorSender.String() {
		logger.Error("operatorSender is not equal OperatorAddress")
	}
	txMgr := txmgr.NewSimpleTxManager(ethRpcClient, logger, signer, common.HexToAddress(c.OperatorAddress))

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

	operator := &Operator{
		config:             c,
		logger:             logger,
		nodeApi:            nodeApi,
		ethClient:          ethRpcClient,
		avsWriter:          avsWriter,
		avsReader:          *avsReader,
		ethWsClient:        ethWsClient,
		blsKeypair:         blsKeyPair,
		operatorAddr:       common.HexToAddress(c.OperatorAddress),
		newTaskCreatedChan: make(chan *avs.ContracthelloWorldTaskCreated),
		avsAddr:            common.HexToAddress(c.AVSAddress),
		epochIdentifier:    epochIdentifier,
		contractABI:        *contractABI,
	}

	if c.RegisterOperatorOnStartup {
		operator.registerOperatorOnStartup()
	}
	// Wait for transaction which operator optin avs to be mined
	time.Sleep(5 * retryDelay)
	logger.Info("Operator info",
		"operatorAddr", c.OperatorAddress,
		"operatorKey", operator.blsKeypair.PublicKey().Marshal(),
	)

	return operator, nil
}

func (o *Operator) Start(ctx context.Context) error {
	// 1.operator register chain
	// 2.operator opt-in avs
	// 3.operator accept staker delegation so that avs voting power is not 0, otherwise the task cannot be created
	// 4.operator register BLSPublicKey
	// 5.operator submit task

	operatorAddress, err := core.SwitchEthAddressToImAddress(o.operatorAddr.String())
	if err != nil {
		o.logger.Error("Cannot switch eth address to im address", "err", err)
		panic(err)
	}

	flag, err := o.avsReader.IsOperator(&bind.CallOpts{}, o.operatorAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec IsOperator", "err", err)
		return err
	}
	if !flag {
		o.logger.Error("Operator is not registered.", "err", err)
		panic(fmt.Sprintf("Operator is not registered: %s", operatorAddress))
	}

	pubKey, err := o.avsReader.GetRegisteredPubkey(&bind.CallOpts{}, o.operatorAddr.String(), o.avsAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec GetRegisteredPubKey", "err", err)
		return err
	}

	if len(pubKey) == 0 {
		// operator register BLSPublicKey  via evm tx
		msg := fmt.Sprintf(core.BLSMessageToSign,
			core.ChainIDWithoutRevision("imuachainlocalnet_232"), operatorAddress)
		hashedMsg := crypto.Keccak256Hash([]byte(msg))
		sig := o.blsKeypair.Sign(hashedMsg.Bytes())

		_, err = o.avsWriter.RegisterBLSPublicKey(
			context.Background(),
			o.avsAddr.String(),
			o.blsKeypair.PublicKey().Marshal(),
			sig.Marshal())

		if err != nil {
			o.logger.Error("operator failed to registerBLSPublicKey", "err", err)
			return err
		}
	}

	pubKey, err = o.avsReader.GetRegisteredPubkey(&bind.CallOpts{}, o.operatorAddr.String(), o.avsAddr.String())
	if err != nil {
		o.logger.Error("Cannot exec GetRegisteredPubKey", "err", err)
		return err
	}
	if len(pubKey) == 0 {
		o.logger.Error("Cannot exec GetRegisteredPubKey", "err", err)
	}
	// Make sure the amount can be queried
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 1.check operator delegation usd amount
		amount, err := o.avsReader.GetOperatorOptedUSDValue(&bind.CallOpts{}, o.avsAddr.String(), o.operatorAddr.String())
		if err != nil {
			o.logger.Error("Cannot exec GetOperatorOptedUSDValue", "err", err)
			return err
		}
		if err == nil && !amount.IsZero() && !amount.IsNegative() {
			break
		}
		// 2.Perform Deposit and so on
		if amount.IsZero() {
			//deposit and delegate
			err := o.Deposit()
			if err != nil {
				panic(fmt.Sprintf("Can not Deposit: %s", err))
			}
			err = o.Delegate()
			if err != nil {
				panic(fmt.Sprintf("Can not Delegate: %s", err))

			}
			err = o.SelfDelegate()
			if err != nil {
				panic(fmt.Sprintf("Can not SelfDelegate: %s", err))

			}
		}
		// USD value voting power is updated by epoch for cycle
		// So we need to wait for an epoch to work and try again,
		// but we need to make our tests more efficient, so we'll wait for an applicable time
		for attempt := 1; attempt <= maxRetries; attempt++ {
			amount, lastErr := o.avsReader.GetOperatorOptedUSDValue(&bind.CallOpts{}, o.avsAddr.String(), o.operatorAddr.String())

			if lastErr == nil && !amount.IsZero() && !amount.IsNegative() {
				break
			}
			if lastErr != nil {
				o.logger.Error("Cannot GetOperatorOptedUSDValue",
					"err", lastErr,
					"attempt", attempt,
					"max_attempts", maxRetries)
			} else {
				o.logger.Info("OperatorOptedUSDValue is zero or negative",
					"operator usd value", amount,
					"attempt", attempt,
					"max_attempts", maxRetries)
			}
			time.Sleep(retryDelay)
		}
		// 3.After checking the above maxRetries times ,check operator delegation usd amount again
		amount, err = o.avsReader.GetOperatorOptedUSDValue(&bind.CallOpts{}, o.avsAddr.String(), o.operatorAddr.String())
		if err != nil {
			o.logger.Error("Cannot exec GetOperatorOptedUSDValue", "err", err)
			return err
		}
		if amount.IsZero() {
			//deposit and delegate
			err := o.Deposit()
			if err != nil {
				panic(fmt.Sprintf("Can not Deposit: %s", err))
			}
			err = o.Delegate()
			if err != nil {
				panic(fmt.Sprintf("Can not Delegate: %s", err))
			}
			err = o.SelfDelegate()
			if err != nil {
				panic(fmt.Sprintf("Can not SelfDelegate: %s", err))
			}
		}
	}
	o.logger.Infof("Starting operator.")

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
				taskResponse := o.ProcessNewTaskCreatedLog(e)
				sig, resBytes, err := o.SignTaskResponse(taskResponse)
				if err != nil {
					o.logger.Error("Failed to sign task response", "err", err)
					continue
				}
				taskInfo, _ := o.avsReader.GetTaskInfo(&bind.CallOpts{}, o.avsAddr.String(), taskResponse.TaskID)
				go func() {
					_, err := o.SendSignedTaskResponseToChain(context.Background(), taskResponse.TaskID, resBytes, sig, taskInfo)
					if err != nil {

					}
				}()
			}
		}
	}
}
func (o *Operator) parseEvent(vLog ethtypes.Log) (interface{}, error) {

	vLog.Topics[0] = o.contractABI.Events["TaskCreated"].ID

	var event avs.ContracthelloWorldTaskCreated

	err := o.contractABI.UnpackIntoInterface(&event, "TaskCreated", vLog.Data)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ProcessNewTaskCreatedLog TaskResponse is the struct that is signed and sent to the chain as a task response.
func (o *Operator) ProcessNewTaskCreatedLog(e *avs.ContracthelloWorldTaskCreated) *core.TaskResponse {
	o.logger.Info("New Task Created", "TaskID", e.TaskId.Uint64(),
		"Issuer", e.Issuer.String(), "Name", e.Name, "NumberToBeSquared", e.NumberToBeSquared)
	// Perform multiplication to complete the task requirements
	taskResponse := &core.TaskResponse{
		TaskID:        e.TaskId.Uint64(),
		NumberSquared: e.NumberToBeSquared * e.NumberToBeSquared,
	}
	return taskResponse
}

func (o *Operator) SignTaskResponse(taskResponse *core.TaskResponse) ([]byte, []byte, error) {

	taskResponseHash, data, err := core.GetTaskResponseDigestEncodeByAbi(*taskResponse)
	if err != nil {
		o.logger.Error("Error SignTaskResponse with getting task response header hash. skipping task (this is not expected and should be investigated)", "err", err)
		return nil, nil, err
	}
	msgBytes := taskResponseHash[:]

	sig := o.blsKeypair.Sign(msgBytes)

	return sig.Marshal(), data, nil
}

func (o *Operator) SendSignedTaskResponseToChain(
	ctx context.Context,
	taskId uint64,
	taskResponse []byte,
	blsSignature []byte,
	taskInfo avs.TaskInfo) (string, error) {

	startingEpoch := taskInfo.StartingEpoch
	taskResponsePeriod := taskInfo.TaskResponsePeriod
	taskStatisticalPeriod := taskInfo.TaskStatisticalPeriod

	// Track submission status for each phase
	phaseOneSubmitted := false
	phaseTwoSubmitted := false

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err() // Gracefully exit if context is canceled
		default:
			// Fetch the current epoch information
			epochIdentifier, err := o.avsReader.GetAVSEpochIdentifier(&bind.CallOpts{}, o.avsAddr.String())
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
				o.logger.Info("Exiting loop: Task period has passed",
					"Task", taskInfo.TaskContractAddress.String()+"--"+strconv.FormatUint(taskId, 10))
				return "The current task period has passed:", nil
			}

			switch {
			case currentEpoch <= startingEpoch:
				o.logger.Info("current epoch is less than or equal to the starting epoch", "currentEpoch", currentEpoch, "startingEpoch", startingEpoch, "taskId", taskId)
				time.Sleep(retryDelay)

			case currentEpoch <= startingEpoch+taskResponsePeriod:
				if !phaseOneSubmitted {
					o.logger.Info("Execute Phase One Submission Task", "currentEpoch", currentEpoch,
						"startingEpoch", startingEpoch, "taskResponsePeriod", taskResponsePeriod, "taskId", taskId)
					o.logger.Info("Submitting task response for task response period",
						"taskAddr", o.avsAddr.String(), "taskId", taskId, "operator-addr", o.operatorAddr)
					_, err := o.avsWriter.OperatorSubmitTask(
						ctx,
						taskId,
						nil,
						blsSignature,
						o.avsAddr.String(),
						1)
					if err != nil {
						o.logger.Error("Avs failed to OperatorSubmitTask", "err", err, "taskId", taskId)
						return "", fmt.Errorf("failed to submit task during taskResponsePeriod: %w", err)
					}
					phaseOneSubmitted = true
					o.logger.Info("Successfully submitted task response for phase one", "taskId", taskId)
				} else {
					o.logger.Info("Phase One already submitted", "taskId", taskId)
					time.Sleep(retryDelay)
				}

			case currentEpoch <= startingEpoch+taskResponsePeriod+taskStatisticalPeriod && currentEpoch > startingEpoch+taskResponsePeriod:
				if !phaseTwoSubmitted {
					o.logger.Info("Execute Phase Two Submission Task", "currentEpoch", currentEpoch,
						"startingEpoch", startingEpoch, "taskResponsePeriod", taskResponsePeriod, "taskStatisticalPeriod", taskStatisticalPeriod, "taskId", taskId)
					o.logger.Info("Submitting task response for statistical period",
						"taskAddr", o.avsAddr.String(), "taskId", taskId, "operator-addr", o.operatorAddr)
					_, err := o.avsWriter.OperatorSubmitTask(
						ctx,
						taskId,
						taskResponse,
						blsSignature,
						o.avsAddr.String(),
						2)
					if err != nil {
						o.logger.Error("Avs failed to OperatorSubmitTask", "err", err, "taskId", taskId)
						return "", fmt.Errorf("failed to submit task during statistical period: %w", err)
					}
					phaseTwoSubmitted = true
					o.logger.Info("Successfully submitted task response for phase two", "taskId", taskId)
				} else {
					o.logger.Info("Phase Two already submitted", "taskId", taskId)
					time.Sleep(retryDelay)
				}

			default:
				o.logger.Info("Current epoch is not within expected range", "currentEpoch", currentEpoch, "taskId", taskId)
				return "", fmt.Errorf("current epoch %d is not within expected range %d", currentEpoch, startingEpoch)
			}

			// If both phases are submitted, exit the loop
			if phaseOneSubmitted && phaseTwoSubmitted {
				o.logger.Info("Both phases completed successfully", "taskId", taskId)
				return "Both task response phases completed successfully", nil
			}

			// Add a small delay to prevent tight looping
			time.Sleep(retryDelay)
		}
	}
}
