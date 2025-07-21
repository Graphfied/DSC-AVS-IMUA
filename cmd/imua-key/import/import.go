package _import

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	utils "github.com/imua-xyz/imua-avs-sdk/crypto/bls"
	"github.com/imua-xyz/imua-avs-sdk/crypto/ecdsa"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls/blst"
	blscommon "github.com/prysmaticlabs/prysm/v5/crypto/bls/common"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

var (
	KeyTypeFlag = &cli.StringFlag{
		Name:     "key-type",
		Usage:    "key type to import (ecdsa or bls)",
		Required: true,
	}
	PrivateKeyFlag = &cli.StringFlag{
		Name:     "private-key",
		Usage:    "hex of private key",
		Required: true,
	}
	PasswordFlag = &cli.StringFlag{
		Name:        "password",
		Usage:       "encrypt the key with a password and write it to the specified JSON file",
		Required:    false,
		DefaultText: "",
	}
	OutputDirFlag = &cli.StringFlag{
		Name:     "output-dir",
		Usage:    "folder to store key",
		Required: true,
	}
)

var Command = &cli.Command{
	Name:    "importKey",
	Aliases: []string{"i"},
	Description: `Import keys for testing purpose.
This command encrypt the key with a password and write it to the specified JSON file.
`,
	Action: importKey,
	Flags: []cli.Flag{
		KeyTypeFlag,
		PrivateKeyFlag,
		PasswordFlag,
		OutputDirFlag,
	},
}

func importKey(c *cli.Context) error {
	keyType := c.String(KeyTypeFlag.Name)
	privateKey := c.String(PrivateKeyFlag.Name)
	password := c.String(PasswordFlag.Name)
	outputDir := c.String(OutputDirFlag.Name)

	switch keyType {
	case "ecdsa":
		ImportEcdsaKeyToFile(privateKey, password, outputDir)
	case "bls":
		ImportBlsKeyToFile(privateKey, password, outputDir)
	default:
		return nil
	}
	return nil
}

func SaveToFile(key blscommon.SecretKey, path string, password string) error {
	cryptoStruct, err := keystore.EncryptDataV3(
		key.Marshal(),
		[]byte(password),
		keystore.StandardScryptN,
		keystore.StandardScryptP,
	)
	if err != nil {
		return err
	}

	encryptedBLSStruct := utils.EncryptedBLSKeyJSONV3{
		PubKey: hex.EncodeToString(key.PublicKey().Marshal()),
		Crypto: cryptoStruct,
	}
	data, err := json.Marshal(encryptedBLSStruct)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println("Error creating directories:", err)
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ImportEcdsaKeyToFile(privateKeyHex, pw, keyPath string) error {
	key, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to convert hex to ECDSA: %v", err)
	}

	// Check if the length of privateKeyHex is 32 bytes (64 characters)
	lenPrivateKey := len(privateKeyHex)
	if lenPrivateKey != 64 {
		return fmt.Errorf("the private key is not 32 bytes: %s %d", privateKeyHex, lenPrivateKey)
	}
	currentDir, _ := os.Getwd()
	KeystorePath := filepath.Join(currentDir, keyPath)

	// Extract directory and filename from the given path
	dir := filepath.Dir(KeystorePath)

	// Ensure the directory exists
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write the key to the file
	err = ecdsa.WriteKey(KeystorePath, key, pw)
	if err != nil {
		return fmt.Errorf("failed to write key: %v", err)
	}

	fmt.Printf("Key successfully written to %s", KeystorePath)
	return nil
}

func ImportBlsKeyToFile(privateKeyHex, pw, keyPath string) error {
	decodeString, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return err
	}
	privateKey, err := blst.SecretKeyFromBytes(decodeString)

	if err != nil {
		return fmt.Errorf("failed to convert hex to ECDSA: %v", err)
	}

	currentDir, err := os.Getwd()
	KeystorePath := filepath.Join(currentDir, keyPath)

	// Extract directory and filename from the given path
	dir := filepath.Dir(KeystorePath)

	// Ensure the directory exists
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	err = SaveToFile(privateKey, KeystorePath, pw)
	if err != nil {
		return fmt.Errorf("failed to write key: %v", err)
	}

	return nil
}
