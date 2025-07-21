package chainio

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/imua-xyz/imua-avs-sdk/logging"

	avs "github.com/imua-xyz/imua-avs/contracts/bindings/avs"
	"github.com/imua-xyz/imua-avs/core/chainio/eth"
)

type ContractBindings struct {
	AvsAddr    gethcommon.Address
	AVSManager *avs.ContracthelloWorld
}

func NewContractBindings(
	avsAddr gethcommon.Address,
	ethclient eth.EthClient,
	logger logging.Logger,
) (*ContractBindings, error) {
	contractAvsManager, err := avs.NewContracthelloWorld(avsAddr, ethclient)
	if err != nil {
		logger.Error("Failed to fetch Avs contract", "err", err)
		return nil, err
	}

	return &ContractBindings{
		AvsAddr:    avsAddr,
		AVSManager: contractAvsManager,
	}, nil
}
