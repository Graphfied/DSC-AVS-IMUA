package avs

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	sdkEcdsa "github.com/imua-xyz/imua-avs-sdk/crypto/ecdsa"
	"github.com/imua-xyz/imua-avs-sdk/logging"
	sdklogging "github.com/imua-xyz/imua-avs-sdk/logging"
	"github.com/imua-xyz/imua-avs-sdk/signer"
	avs "github.com/imua-xyz/imua-avs/contracts/bindings/avs"
	"github.com/imua-xyz/imua-avs/core"
	chain "github.com/imua-xyz/imua-avs/core/chainio"
	"github.com/imua-xyz/imua-avs/core/chainio/eth"
	"github.com/imua-xyz/imua-avs/types"
	"math/big"
	"math/rand"
	"os"
	"time"
)

const (
	avsName    = "hello-avs-demo"
	maxRetries = 25
	retryDelay = 6 * time.Second
	// DayEpochID defines the identifier for a daily epoch.
	DayEpochID = "day"
	// HourEpochID defines the identifier for an hourly epoch.
	HourEpochID = "hour"
	// MinuteEpochID defines the identifier for an epoch that is a minute long.
	MinuteEpochID = "minute"
	// WeekEpochID defines the identifier for a weekly epoch.
	WeekEpochID = "week"
)

type Avs struct {
	logger                logging.Logger
	avsWriter             chain.AvsWriter
	avsReader             chain.AvsReader
	avsAddress            string
	createTaskInterval    int64
	taskResponsePeriod    uint64
	taskChallengePeriod   uint64
	thresholdPercentage   uint8
	taskStatisticalPeriod uint64
	avsEpochIdentifier    string
}

// NewAvs creates a new Avs with the provided config.
func NewAvs(c *types.NodeConfig) (*Avs, error) {
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

	ethRpcClient, err := eth.NewClient(c.EthRpcUrl)
	if err != nil {
		logger.Error("Cannot create http ethclient", "err", err)
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

	signer, avsSender, err := signer.SignerFromConfig(signer.Config{
		KeystorePath: c.AVSEcdsaPrivateKeyStorePath,
		Password:     ecdsaKeyPassword,
	}, chainId)
	if err != nil {
		panic(err)
	}
	logger.Info("avsSender:", "avsSender", avsSender.String())
	logger.Info("AVSOwnerAddress:", "AVSOwnerAddress", c.AVSOwnerAddress)

	balance, err := ethRpcClient.BalanceAt(context.Background(), avsSender, nil)
	if err != nil {
		logger.Error("Cannot get Balance", "err", err)
	}
	if balance.Cmp(big.NewInt(0)) != 1 {
		logger.Error("avsSender has not enough Balance")
	}
	if c.AVSOwnerAddress != avsSender.String() {
		logger.Error("avsSender is not equal AVSOwnerAddress")
	}
	code, err := ethRpcClient.CodeAt(context.Background(), common.HexToAddress(c.AVSAddress), nil)
	if err != nil {
		logger.Error("Cannot get code", "err", err)
	}

	if c.AVSAddress == "" || len(code) < 3 {
		logger.Info("AVS_ADDRESS env var not set. will deploy avs contract")

		key, err := sdkEcdsa.ReadKey(c.AVSEcdsaPrivateKeyStorePath, ecdsaKeyPassword)

		avsAddr, _, err := chain.DeployAVS(
			ethRpcClient,
			logger,
			*key,
			chainId,
		)
		if err != nil {
			panic(err)
		}

		c.AVSAddress = avsAddr.String()
		c.TaskAddress = avsAddr.String()
		c.AVSRewardAddress = avsAddr.String()
		c.AVSSlashAddress = avsAddr.String()
		filePath, err := core.GetFileInCurrentDirectory("config.yaml")
		if err != nil {
			panic(err)
		}
		err = core.UpdateYAMLWithComments(filePath, "avs_address", avsAddr.String())
		err = core.UpdateYAMLWithComments(filePath, "avs_reward_address", avsAddr.String())
		err = core.UpdateYAMLWithComments(filePath, "avs_slash_address", avsAddr.String())
		err = core.UpdateYAMLWithComments(filePath, "task_address", avsAddr.String())

		if err != nil {
			logger.Error("AVS_ADDRESS env var not set. will deploy avs contract")

		}
	}

	txMgr := txmgr.NewSimpleTxManager(ethRpcClient, logger, signer, avsSender)
	avsWriter, err := chain.BuildChainWriter(
		common.HexToAddress(c.AVSAddress),
		ethRpcClient,
		logger,
		txMgr)
	if err != nil {
		logger.Error("Cannot create avsWriter", "err", err)
		return nil, err
	}

	avsReader, err := chain.BuildChainReader(
		common.HexToAddress(c.AVSAddress),
		ethRpcClient,
		logger)
	if err != nil {
		logger.Error("Cannot create chainReader", "err", err)
		return nil, err
	}
	// Wait for transaction which avs deployed to be mined
	time.Sleep(retryDelay)
	info, err := avsReader.GetAVSEpochIdentifier(&bind.CallOpts{}, c.AVSAddress)
	if err != nil {
		logger.Error("Cannot GetAVSEpochIdentifier", "err", err)
		return nil, err
	}
	if info == "" {
		params := avs.AVSParams{
			Sender:              common.HexToAddress(c.AVSOwnerAddress),
			AvsName:             avsName,
			MinStakeAmount:      c.MinStakeAmount,
			TaskAddress:         common.HexToAddress(c.TaskAddress),
			SlashAddress:        common.HexToAddress(c.AVSRewardAddress),
			RewardAddress:       common.HexToAddress(c.AVSSlashAddress),
			AvsOwnerAddresses:   core.ConvertToEthAddresses(c.AvsOwnerAddresses),
			WhitelistAddresses:  core.ConvertToEthAddresses(c.WhitelistAddresses),
			AssetIDs:            c.AssetIDs,
			AvsUnbondingPeriod:  c.AvsUnbondingPeriod,
			MinSelfDelegation:   c.MinSelfDelegation,
			EpochIdentifier:     c.EpochIdentifier,
			MiniOptInOperators:  1,
			MinTotalStakeAmount: 1,
			AvsRewardProportion: 5,
			AvsSlashProportion:  5,
		}
		_, err = avsWriter.RegisterAVSToChain(context.Background(),
			params,
		)
		if err != nil {
			logger.Error("register Avs failed ", "err", err)
			return &Avs{}, err
		}
	}
	info, _ = avsReader.GetAVSEpochIdentifier(&bind.CallOpts{}, c.AVSAddress)

	return &Avs{
		logger:                logger,
		avsWriter:             avsWriter,
		avsReader:             avsReader,
		avsAddress:            c.AVSAddress,
		createTaskInterval:    c.CreateTaskInterval,
		taskResponsePeriod:    c.TaskResponsePeriod,
		taskChallengePeriod:   c.TaskChallengePeriod,
		thresholdPercentage:   c.ThresholdPercentage,
		taskStatisticalPeriod: c.TaskStatisticalPeriod,
		avsEpochIdentifier:    info,
	}, nil
}

func (avs *Avs) Start(ctx context.Context) error {
	avs.logger.Infof("Starting avs.")
	ticker := time.NewTicker(time.Duration(avs.createTaskInterval) * time.Second)
	avs.logger.Infof("Avs owner set to send new task every %d seconds", avs.createTaskInterval)
	defer ticker.Stop()
	taskNum := int64(1)
	// Wait for the operator process to prepare work, such as deposit delegation, before sending the task
	time.Sleep(20 * time.Second)
	err := avs.sendNewTask()
	if err != nil {
		// we log the errors inside sendNewTask() so here we just continue to do the next task
		avs.logger.Info("sendNewTask encountered an error: %v; continuing to do the next task.", err)
	}
	taskNum++
	for {
		select {
		case <-ctx.Done():
			avs.logger.Info("Context canceled; stopping AVS.")
			return nil
		case <-ticker.C:
			avs.logger.Info("sendNewTask-num:", "taskNum", taskNum)
			err := avs.sendNewTask()
			if err != nil {
				// we log the errors inside sendNewTask() so here we just continue to do the next task
				avs.logger.Info("sendNewTask encountered an error: %v; continuing to do the next task.", err)
				continue
			}
			taskNum++
		}
	}
}

// sendNewTask sends a new task to the task manager contract.
func (avs *Avs) sendNewTask() error {
	avs.logger.Info("Avs sending new task")
	var taskPowerTotal sdkmath.LegacyDec
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		taskPowerTotal, lastErr = avs.avsReader.GtAVSUSDValue(&bind.CallOpts{}, avs.avsAddress)

		if lastErr == nil && !taskPowerTotal.IsZero() && !taskPowerTotal.IsNegative() {
			break
		}

		if lastErr != nil {
			avs.logger.Error("Cannot get AVSUSDValue",
				"err", lastErr,
				"attempt", attempt,
				"max_attempts", maxRetries)
		} else {
			avs.logger.Info("AVS USD value is zero or negative",
				"avs usd value", taskPowerTotal,
				"attempt", attempt,
				"max_attempts", maxRetries)
		}

		if attempt == maxRetries {
			panic("the voting power of AVS is zero or negative")
		}
		// USD value voting power is updated by epoch for cycle
		// So we need to wait for an epoch to work and try again,
		// but we need to make our tests more efficient, so we'll leave it out for now
		/*var sleepDuration time.Duration
		switch avs.avsEpochIdentifier {
		case DayEpochID:
			sleepDuration = 24 * time.Hour
		case HourEpochID:
			sleepDuration = time.Hour
		case MinuteEpochID:
			sleepDuration = time.Minute
		case WeekEpochID:
			sleepDuration = 7 * 24 * time.Hour
		}
				time.Sleep(1 * sleepDuration)
		*/
		time.Sleep(retryDelay)
	}

	if taskPowerTotal.IsZero() || taskPowerTotal.IsNegative() {
		// panic("the voting power of AVS is zero or negative")
	}
	_, err := avs.avsWriter.CreateNewTask(
		context.Background(),
		GenerateRandomName(5),
		uint64(rand.Intn(500)),
		avs.taskResponsePeriod,
		avs.taskChallengePeriod,
		avs.thresholdPercentage,
		avs.taskStatisticalPeriod)

	if err != nil {
		avs.logger.Error("Avs failed to sendNewTask", "err", err)
		return err
	}
	return nil
}
func GenerateRandomName(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
