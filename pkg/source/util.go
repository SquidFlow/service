package source

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/squidflow/service/pkg/fs"
)

// Helper function to copy files from repofs to memory filesystem
func copyToMemFS(repofs fs.FS, srcPath, destPath string, memFS filesys.FileSystem) error {
	entries, err := repofs.ReadDir(srcPath)
	if err != nil {
		return err
	}

	// Create the destination directory in memFS
	if err := memFS.MkdirAll(destPath); err != nil {
		return fmt.Errorf("failed to create directory %s in memory fs: %w", destPath, err)
	}

	for _, entry := range entries {
		srcFilePath := repofs.Join(srcPath, entry.Name())
		destFilePath := filepath.Join(destPath, entry.Name())

		if entry.IsDir() {
			if err := copyToMemFS(repofs, srcFilePath, destFilePath, memFS); err != nil {
				return err
			}
			continue
		}

		// Read file content from repofs
		content, err := repofs.ReadFile(srcFilePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", srcFilePath, err)
		}

		// Write file to memFS
		err = memFS.WriteFile(destFilePath, content)
		if err != nil {
			return fmt.Errorf("failed to write file %s to memory fs: %w", destFilePath, err)
		}
	}

	return nil
}
