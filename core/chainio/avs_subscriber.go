package chainio

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"

	"github.com/imua-xyz/imua-avs-sdk/logging"
	avssub "github.com/imua-xyz/imua-avs/contracts/bindings/avs"
	"github.com/imua-xyz/imua-avs/core/chainio/eth"
)

type AvsRegistrySubscriber interface {
	SubscribeToNewTasks(newTaskCreatedChan chan *avssub.ContracthelloWorldTaskCreated) event.Subscription
}

type AvsRegistryChainSubscriber struct {
	logger logging.Logger
	avssub avssub.ContracthelloWorld
}

// forces EthSubscriber to implement the chainio.Subscriber interface
var _ AvsRegistrySubscriber = (*AvsRegistryChainSubscriber)(nil)

func NewAvsRegistryChainSubscriber(
	avssub avssub.ContracthelloWorld,
	logger logging.Logger,
) (*AvsRegistryChainSubscriber, error) {
	return &AvsRegistryChainSubscriber{
		logger: logger,
		avssub: avssub,
	}, nil
}

func BuildAvsRegistryChainSubscriber(
	avssubAddr common.Address,
	ethWsClient eth.EthClient,
	logger logging.Logger,
) (*AvsRegistryChainSubscriber, error) {
	avssub, err := avssub.NewContracthelloWorld(avssubAddr, ethWsClient)
	if err != nil {
		logger.Error("Failed to create BLSApkRegistry contract", "err", err)
		return nil, err
	}
	return NewAvsRegistryChainSubscriber(*avssub, logger)
}

func (s *AvsRegistryChainSubscriber) SubscribeToNewTasks(newTaskCreatedChan chan *avssub.ContracthelloWorldTaskCreated) event.Subscription {
	sub, err := s.avssub.WatchTaskCreated(
		&bind.WatchOpts{}, newTaskCreatedChan,
	)
	if err != nil {
		s.logger.Error("Failed to subscribe to new  tasks", "err", err)
	}
	s.logger.Infof("Subscribed to new TaskManager tasks")
	return sub
}
