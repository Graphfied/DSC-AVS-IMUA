package _import_test

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	sdkBls "github.com/imua-xyz/imua-avs-sdk/crypto/bls"
	sdkEcdsa "github.com/imua-xyz/imua-avs-sdk/crypto/ecdsa"
	"strings"

	"github.com/prysmaticlabs/prysm/v5/crypto/bls/blst"
	"os"
	"path/filepath"
	"testing"

	key "github.com/imua-xyz/imua-avs/cmd/imua-key/import"
)

const (
	KeystorePath = "/keys/key.json"
)

func TestEcdsaKey(t *testing.T) {

	ecdsaKeyPassword, _ := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")

	expectedPrivateKey := "d196dca836f8ac2fff45b3c9f0113825ccbb33fa1b39737b948503b263ed75ae"

	err := key.ImportEcdsaKeyToFile(expectedPrivateKey, ecdsaKeyPassword, KeystorePath)
	if err != nil {
		t.Fatalf("Error importing key: %v", err)
	}

	currentDir, _ := os.Getwd()
	KeystorePath := filepath.Join(currentDir, KeystorePath)

	importedKey, err := sdkEcdsa.ReadKey(KeystorePath, ecdsaKeyPassword)
	if err != nil {
		t.Fatalf("Error reading imported key: %v", err)
	}

	actualPrivateKey := hex.EncodeToString(importedKey.D.Bytes())
	if actualPrivateKey != expectedPrivateKey {
		t.Fatalf("Expected private key: %s, but got: %s", expectedPrivateKey, actualPrivateKey)
	}

	fmt.Println("Test passed: Imported and read ecdsa key successfully")
}

func TestBlsKey(t *testing.T) {

	blsKeyPassword, _ := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
	privateKey, err := blst.RandKey()
	expectedPrivateKey := hex.EncodeToString(privateKey.Marshal())
	currentDir, _ := os.Getwd()

	KeystorePath := filepath.Join(currentDir, KeystorePath)

	err = key.SaveToFile(privateKey, KeystorePath, blsKeyPassword)
	if err != nil {
		t.Fatalf("Error importing key: %v", err)
	}

	importedKey, err := sdkBls.ReadPrivateKeyFromFile(KeystorePath, blsKeyPassword)
	if err != nil {
		t.Fatalf("Error reading imported key: %v", err)
	}

	actualPrivateKey := hex.EncodeToString(importedKey.Marshal())
	if actualPrivateKey != expectedPrivateKey {
		t.Fatalf("Expected private key: %s, but got: %s", expectedPrivateKey, actualPrivateKey)
	}

	fmt.Println("Test passed: Imported and read bls key successfully")
}
func TestHex(t *testing.T) {

	//bytes := common.HexToAddress("0x3e108c058e8066da635321dc3018294ca82ddedf").Bytes()
	//fmt.Println(bytes)

	iteratorKey := []byte{62, 16, 140, 5, 142, 128, 102, 218, 99, 83, 33, 220, 48, 24, 41, 76, 168, 45, 222, 223}
	avsAddr := common.BytesToAddress(iteratorKey).String()
	fmt.Println(avsAddr)

	a := "0x3e108c058e8066da635321dc3018294ca82ddedf"
	b := "0x3e108c058e8066DA635321Dc3018294cA82ddEdf"
	fmt.Println([]byte(a))
	fmt.Println([]byte(b))

	fmt.Println(strings.ToLower(strings.ToLower(strings.ToLower(strings.ToLower(b)))))

}

func TestBlsKeyCmd(t *testing.T) {

	pw := ""
	expectedPrivateKey := "61f470e2cc50746480a2af8530718109140d0f7fd7875b856f34a3e624c68f74"

	err := key.ImportBlsKeyToFile(expectedPrivateKey, pw, KeystorePath)
	if err != nil {
		t.Fatalf("Error importing key: %v", err)
	}

	currentDir, _ := os.Getwd()
	KeystorePath := filepath.Join(currentDir, KeystorePath)

	importedKey, err := sdkBls.ReadPrivateKeyFromFile(KeystorePath, pw)
	if err != nil {
		t.Fatalf("Error reading imported key: %v", err)
	}

	actualPrivateKey := hex.EncodeToString(importedKey.Marshal())
	if actualPrivateKey != expectedPrivateKey {
		t.Fatalf("Expected private key: %s, but got: %s", expectedPrivateKey, actualPrivateKey)
	}

	fmt.Println("Test passed: Imported and read bls key successfully")

}
