package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CreateModuleArchive(sourceDir string) (string, error) {
	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("module-%d.zip", os.Getpid()))

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldIgnoreFile(path, info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		relativePath = filepath.ToSlash(relativePath)

		if info.IsDir() {
			_, err := zipWriter.Create(relativePath + "/")
			return err
		}

		zipFileWriter, err := zipWriter.Create(relativePath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFileWriter, file)
		return err
	})

	if err != nil {
		os.Remove(zipPath)
		return "", err
	}

	return zipPath, nil
}

func shouldIgnoreFile(path string, info os.FileInfo) bool {
	name := info.Name()

	if strings.HasPrefix(name, ".") && name != ".gitignore" {
		return true
	}

	ignoreDirs := []string{"node_modules", "dist", "build", ".git", ".vscode", ".idea"}
	for _, dir := range ignoreDirs {
		if name == dir {
			return true
		}
	}

	ignoreExtensions := []string{".log", ".tmp", ".temp", ".swp", ".swo"}
	for _, ext := range ignoreExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}

	return false
}
