package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type Manifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Main        string   `json:"main"`
	Providers   []string `json:"providers"`
	Interfaces  []string `json:"interfaces"`
}

func ReadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le manifest: %v", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("manifest JSON invalide: %v", err)
	}

	if manifest.ID == "" {
		return nil, fmt.Errorf("ID du module requis")
	}
	if manifest.Name == "" {
		return nil, fmt.Errorf("nom du module requis")
	}
	if manifest.Version == "" {
		return nil, fmt.Errorf("version du module requise")
	}

	return &manifest, nil
}
