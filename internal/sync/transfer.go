package sync

import (
	"context"
	"path/filepath"
	"strings"
)

type LocalTransferer struct{}

func NewTransferer() *LocalTransferer { return &LocalTransferer{} }

func (t *LocalTransferer) Upload(ctx context.Context, source Filesystem, sourcePath string, target Filesystem, targetPath string, opts TransferOptions) (TransferResult, error) {
	return copyTree(ctx, source, sourcePath, target, targetPath, opts)
}

func (t *LocalTransferer) Download(ctx context.Context, source Filesystem, sourcePath string, target Filesystem, targetPath string, opts TransferOptions) (TransferResult, error) {
	return copyTree(ctx, source, sourcePath, target, targetPath, opts)
}

func copyTree(ctx context.Context, source Filesystem, sourcePath string, target Filesystem, targetPath string, opts TransferOptions) (TransferResult, error) {
	var result TransferResult
	if err := target.MkdirAll(targetPath, 0o755); err != nil && opts.CreateDirs {
		return result, err
	}

	info, err := source.Stat(sourcePath)
	if err != nil {
		return result, err
	}

	if !info.IsDir() {
		data, err := source.ReadFile(sourcePath)
		if err != nil {
			return result, err
		}
		if err := target.WriteFile(targetPath, data, 0o600); err != nil {
			return result, err
		}
		result.FilesTransferred = 1
		result.BytesTransferred = int64(len(data))
		return result, nil
	}

	base := filepath.Clean(sourcePath)
	err = walkDir(ctx, source, base, func(rel string, isDir bool) error {
		if ctx != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		dst := filepath.Join(targetPath, rel)
		if isDir {
			return target.MkdirAll(dst, 0o755)
		}
		data, err := source.ReadFile(filepath.Join(base, rel))
		if err != nil {
			return err
		}
		if shouldSkipBySize(int64(len(data)), opts.MaxFileSize) {
			result.FilesSkipped++
			return nil
		}
		if err := target.WriteFile(dst, data, 0o600); err != nil {
			return err
		}
		result.FilesTransferred++
		result.BytesTransferred += int64(len(data))
		return nil
	})
	return result, err
}

func walkDir(ctx context.Context, fsys Filesystem, root string, fn func(rel string, isDir bool) error) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	entries, err := fsys.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		rel := entry.Name()
		full := filepath.Join(root, rel)
		if err := fn(rel, entry.IsDir()); err != nil {
			return err
		}
		if entry.IsDir() {
			if err := walkSubDir(ctx, fsys, full, rel, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

func walkSubDir(ctx context.Context, fsys Filesystem, root, prefix string, fn func(rel string, isDir bool) error) error {
	entries, err := fsys.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		rel := filepath.Join(prefix, entry.Name())
		if err := fn(rel, entry.IsDir()); err != nil {
			return err
		}
		if entry.IsDir() {
			if err := walkSubDir(ctx, fsys, filepath.Join(root, entry.Name()), rel, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

func shouldSkipBySize(size int64, max int64) bool {
	return max > 0 && size > max
}

func matchesAny(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		if ok, _ := filepath.Match(pattern, path); ok {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}
