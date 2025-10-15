package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type Writer struct {
	outputDir string
}

func NewWriter(outputDir string) *Writer {
	return &Writer{
		outputDir: outputDir,
	}
}

func (w *Writer) WriteBundle(content string, checksum string) error {
	bundlePath := filepath.Join(w.outputDir, "bundle.js")
	if err := ioutil.WriteFile(bundlePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write bundle: %w", err)
	}
	return nil
}

func (w *Writer) WriteManifest(manifest map[string]interface{}) error {
	metadataPath := filepath.Join(w.outputDir, "dashspace.json")

	metadataJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := ioutil.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (w *Writer) WriteSourceMap(sourceMap string) error {
	if sourceMap == "" {
		return nil
	}

	sourceMapPath := filepath.Join(w.outputDir, "bundle.js.map")
	if err := ioutil.WriteFile(sourceMapPath, []byte(sourceMap), 0644); err != nil {
		return fmt.Errorf("failed to write source map: %w", err)
	}

	return nil
}
