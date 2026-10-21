package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	localRootDir = "./data"
	localURLPath = "/static"
)

type localClient struct {
	root string
}

func newLocal() (Client, error) {
	root, err := filepath.Abs(localRootDir)
	if err != nil {
		return nil, fmt.Errorf("解析本地存储目录失败: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("创建本地存储目录失败: %w", err)
	}
	return &localClient{root: root}, nil
}

func (c *localClient) Upload(objectKey string, reader io.Reader, _ string) (string, error) {
	objectKey = strings.TrimPrefix(filepath.ToSlash(objectKey), "/")
	if objectKey == "" || strings.Contains(objectKey, "..") {
		return "", fmt.Errorf("非法对象路径")
	}

	dest := filepath.Join(c.root, filepath.FromSlash(objectKey))
	if err := ensureUnderRoot(c.root, dest); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		_ = os.Remove(dest)
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return localURLPath + "/" + objectKey, nil
}

func (c *localClient) Delete(ctx context.Context, objectKey string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	objectKey = strings.TrimPrefix(filepath.ToSlash(objectKey), "/")
	if objectKey == "" || strings.Contains(objectKey, "..") {
		return fmt.Errorf("非法对象路径")
	}
	dest := filepath.Join(c.root, filepath.FromSlash(objectKey))
	if err := ensureUnderRoot(c.root, dest); err != nil {
		return err
	}
	err := os.Remove(dest)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

func ensureUnderRoot(root, dest string) error {
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("解析文件路径失败: %w", err)
	}
	rel, err := filepath.Rel(root, absDest)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("非法对象路径")
	}
	return nil
}
