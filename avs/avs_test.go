package avs_test

import (
	"github.com/imua-xyz/imua-avs/core"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateYAML(t *testing.T) {
	// Example usage: update 'avs_address' field
	expectedAddr := "0xce5b680d1fd259ada4820e9314bcf0723bdb0000"
	filePath := "../config.yaml"
	key := "avs_address"
	err := core.UpdateYAMLWithComments(filePath, key, expectedAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Read the original YAML file content
	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		t.Fatalf("Error Read the original YAML file content: %v", err)
	}
	// Parse YAML using yaml.v3 node parser to preserve comments
	var doc yaml.Node
	err = yaml.Unmarshal(data, &doc)
	if err != nil {
		t.Fatalf("Error Parse YAML using yaml.v3 node parser to preserve comments: %v", err)
	}
	// Iterate through YAML content to find and update the specified key
	for i := 0; i < len(doc.Content[0].Content); i += 2 {
		if doc.Content[0].Content[i].Value == key {
			// Update the value while preserving original comments
			if expectedAddr != doc.Content[0].Content[i+1].Value {
				t.Fatalf("Expected expectedAddr: %s, but got: %s", expectedAddr, doc.Content[0].Content[i+1].Value)
			}
			break
		}
	}
}
