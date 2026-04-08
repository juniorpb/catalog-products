package catalog

import (
	"encoding/json"
	"fmt"
	"os"
)

// ProductEntry represents a raw product record received from a seller's catalog file.
type ProductEntry struct {
	Id         string  `json:"Id"`
	SellerName string  `json:"SellerName"`
	Name       string  `json:"Name"`
	Brand      *string `json:"Brand"`
	Category   string  `json:"Category"`
}

// ParseJSONFile reads and decodes a JSON file into a slice of ProductEntry.
func ParseJSONFile(path string) ([]ProductEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var entries []ProductEntry
	if err = json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return entries, nil
}
