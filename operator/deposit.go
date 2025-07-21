package operator

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/imua-xyz/imua-avs/core"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	DepositABI  = `[{"inputs":[{"internalType":"uint32","name":"clientChainID","type":"uint32"},{"internalType":"bytes","name":"assetsAddress","type":"bytes"},{"internalType":"bytes","name":"stakerAddress","type":"bytes"},{"internalType":"uint256","name":"opAmount","type":"uint256"}],"name":"depositLST","outputs":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"uint256","name":"latestAssetState","type":"uint256"}],"stateMutability":"nonpayable","type":"function"}]`
	DelegateABI = `[
	{
		"inputs": [
			{
				"internalType": "uint32",
				"name": "clientChainID",
				"type": "uint32"
			},
			{
				"internalType": "bytes",
				"name": "staker",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "operator",
				"type": "bytes"
			}
		],
		"name": "associateOperatorWithStaker",
		"outputs": [
			{
				"internalType": "bool",
				"name": "success",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint32",
				"name": "clientChainID",
				"type": "uint32"
			},
			{
				"internalType": "bytes",
				"name": "assetsAddress",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "stakerAddress",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "operatorAddr",
				"type": "bytes"
			},
			{
				"internalType": "uint256",
				"name": "opAmount",
				"type": "uint256"
			}
		],
		"name": "delegate",
		"outputs": [
			{
				"internalType": "bool",
				"name": "success",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint32",
				"name": "clientChainID",
				"type": "uint32"
			},
			{
				"internalType": "bytes",
				"name": "staker",
				"type": "bytes"
			}
		],
		"name": "dissociateOperatorFromStaker",
		"outputs": [
			{
				"internalType": "bool",
				"name": "success",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint32",
				"name": "clientChainID",
				"type": "uint32"
			},
			{
				"internalType": "bytes",
				"name": "assetsAddress",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "stakerAddress",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "operatorAddr",
				"type": "bytes"
			},
			{
				"internalType": "uint256",
				"name": "opAmount",
				"type": "uint256"
			}
		],
		"name": "undelegate",
		"outputs": [
			{
				"internalType": "bool",
				"name": "success",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
	//	DelegateABI               = `[{"inputs":[{"internalType":"uint32","name":"clientChainID","type":"uint32"},{"internalType":"bytes","name":"staker","type":"bytes"},{"internalType":"bytes","name":"operator","type":"bytes"}],"name":"associateOperatorWithStaker","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint32","name":"clientChainID","type":"uint32"},{"internalType":"uint64","name":"lzNonce","type":"uint64"},{"internalType":"bytes","name":"assetsAddress","type":"bytes"},{"internalType":"bytes","name":"stakerAddress","type":"bytes"},{"internalType":"bytes","name":"operatorAddr","type":"bytes"},{"internalType":"uint256","name":"opAmount","type":"uint256"}],"name":"delegate","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint32","name":"clientChainID","type":"uint32"},{"internalType":"bytes","name":"staker","type":"bytes"}],"name":"dissociateOperatorFromStaker","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint32","name":"clientChainID","type":"uint32"},{"internalType":"uint64","name":"lzNonce","type":"uint64"},{"internalType":"bytes","name":"assetsAddress","type":"bytes"},{"internalType":"bytes","name":"stakerAddress","type":"bytes"},{"internalType":"bytes","name":"operatorAddr","type":"bytes"},{"internalType":"uint256","name":"opAmount","type":"uint256"}],"name":"undelegate","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`
	depositPrecompileAddress  = "0x0000000000000000000000000000000000000804"
	delegatePrecompileAddress = "0x0000000000000000000000000000000000000805"
	defaultAssetID            = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	privateKey                = "D196DCA836F8AC2FFF45B3C9F0113825CCBB33FA1B39737B948503B263ED75AE"
	layerZeroID               = uint32(101)
)

func (operator *Operator) Deposit() error {
	return deposit(operator.config.EthRpcUrl, operator.config.Staker, big.NewInt(operator.config.DepositAmount))
}

func (operator *Operator) Delegate() error {
	operatorbench32Address, err := core.SwitchEthAddressToImAddress(operator.config.OperatorAddress)
	if err != nil {
		operator.logger.Error("Cannot switch eth address to bech32 address", "err", err)
		return err
	}
	return delegateTo(operator.config.EthRpcUrl, operator.config.Staker, operatorbench32Address, big.NewInt(operator.config.DelegateAmount))
}

func (operator *Operator) SelfDelegate() error {
	operatorbench32Address, err := core.SwitchEthAddressToImAddress(operator.config.OperatorAddress)
	if err != nil {
		operator.logger.Error("Cannot switch eth address to bech32 address", "err", err)
		return err
	}
	// trim the "0x" prefix
	staker := operator.config.Staker[2:]
	return selfDelegate(operator.config.EthRpcUrl, staker, operatorbench32Address)
}

func deposit(rpcUrl, stakerAddress string, amount *big.Int) error {
	depositAddr := common.HexToAddress(depositPrecompileAddress)
	assetAddr := common.HexToAddress(defaultAssetID)
	stakerAddr := common.HexToAddress(stakerAddress)
	opAmount := amount

	_, ethClient, err := connectToEthereum(rpcUrl)
	if err != nil {
		return err
	}

	sk, callAddr, err := getPrivateKeyAndAddress(privateKey)
	if err != nil {
		return err
	}

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return err
	}

	depositAbi, err := abi.JSON(strings.NewReader(DepositABI))
	if err != nil {
		return err
	}

	data, err := depositAbi.Pack("depositLST", layerZeroID, paddingAddressTo32(assetAddr), paddingAddressTo32(stakerAddr), opAmount)
	if err != nil {
		return err
	}

	txID, err := sendTransaction(ethClient, chainID, callAddr, sk, depositAddr, data)
	if err != nil {
		return err
	}

	fmt.Println("Deposit Transaction ID:", txID)
	return waitForTransaction(ethClient, txID)
}

func delegateTo(rpcUrl, stakerAddress, operatorBench32Str string, amount *big.Int) error {
	delegateAddr := common.HexToAddress(delegatePrecompileAddress)
	assetAddr := common.HexToAddress(defaultAssetID)
	stakerAddr := common.HexToAddress(stakerAddress)
	operatorAddr := []byte(operatorBench32Str)
	opAmount := amount
	// lzNonce := uint64(0)
	_, ethClient, err := connectToEthereum(rpcUrl)
	if err != nil {
		return err
	}

	sk, callAddr, err := getPrivateKeyAndAddress(privateKey)
	if err != nil {
		return err
	}

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return err
	}

	delegateAbi, err := abi.JSON(strings.NewReader(DelegateABI))
	if err != nil {
		return err
	}

	data, err := delegateAbi.Pack("delegate", layerZeroID, paddingAddressTo32(assetAddr), paddingAddressTo32(stakerAddr), operatorAddr, opAmount)
	if err != nil {
		return err
	}

	txID, err := sendTransaction(ethClient, chainID, callAddr, sk, delegateAddr, data)
	if err != nil {
		return err
	}

	fmt.Println("Delegate To Transaction ID:", txID)
	return waitForTransaction(ethClient, txID)
}

// stakerAddr without "0x" string
func selfDelegate(rpcUrl, stakerAddr, operatorBench32Str string) error {
	delegateAddr := common.HexToAddress(delegatePrecompileAddress)
	operatorAddr := []byte(operatorBench32Str)

	_, ethClient, err := connectToEthereum(rpcUrl)
	if err != nil {
		return err
	}

	sk, callAddr, err := getPrivateKeyAndAddress(privateKey)
	if err != nil {
		return err
	}

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return err
	}

	delegateAbi, err := abi.JSON(strings.NewReader(DelegateABI))
	if err != nil {
		return err
	}

	staker, err := hex.DecodeString(stakerAddr)
	if err != nil {
		return err
	}

	data, err := delegateAbi.Pack("associateOperatorWithStaker", layerZeroID, staker, operatorAddr)
	if err != nil {
		return err
	}

	txID, err := sendTransaction(ethClient, chainID, callAddr, sk, delegateAddr, data)
	if err != nil {
		return err
	}

	fmt.Println("Self Delegate Transaction ID:", txID)
	return waitForTransaction(ethClient, txID)
}

func connectToEthereum(nodeURL string) (*rpc.Client, *ethclient.Client, error) {
	client, err := rpc.DialContext(context.Background(), nodeURL)
	if err != nil {
		return nil, nil, err
	}
	ethClient := ethclient.NewClient(client)
	return client, ethClient, nil
}

func getPrivateKeyAndAddress(privateKey string) (*ecdsa.PrivateKey, common.Address, error) {
	sk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, common.Address{}, err
	}
	callAddr := crypto.PubkeyToAddress(sk.PublicKey)
	return sk, callAddr, nil
}

func paddingAddressTo32(address common.Address) []byte {
	paddingLen := 32 - len(address)
	ret := make([]byte, len(address))
	copy(ret, address[:])
	for i := 0; i < paddingLen; i++ {
		ret = append(ret, 0)
	}
	fmt.Println("Padded address:", hexutil.Encode(ret))
	return ret
}

func sendTransaction(client *ethclient.Client, chainID *big.Int, from common.Address, sk *ecdsa.PrivateKey, to common.Address, data []byte) (string, error) {
	ctx := context.Background()
	nonce, err := client.NonceAt(ctx, from, nil)
	if err != nil {
		return "", err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	gasLimit := uint64(500000)
	tx := types.NewTransaction(nonce, to, big.NewInt(0), gasLimit, gasPrice, data)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, sk)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().String(), nil
}

func waitForTransaction(client *ethclient.Client, txID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	txHash := common.HexToHash(txID)
	tx, _, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %v", err)
	}

	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction to be mined: %v", err)
	}

	if receipt.Status != 1 {
		return fmt.Errorf("transaction failed with status: %v", receipt.Status)
	}

	fmt.Println("Transaction mined successfully with receipt:", receipt)
	return nil
}
