package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	LocalRootDir   = "./data"
	LocalURLPrefix = "/static"
)

type localClient struct {
	root string
}

func newLocal() (Client, error) {
	root, err := filepath.Abs(LocalRootDir)
	if err != nil {
		return nil, fmt.Errorf("解析本地存储目录失败: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("创建本地存储目录失败: %w", err)
	}
	return &localClient{root: root}, nil
}

func (c *localClient) Put(ctx context.Context, key string, r io.Reader, _ int64, _ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dest, err := c.path(key)
	if err != nil {
		return err
	}
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".upload-*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpName := tmp.Name()
	_, copyErr := io.Copy(tmp, r)
	closeErr := tmp.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(tmpName)
		if copyErr != nil {
			return fmt.Errorf("写入文件失败: %w", copyErr)
		}
		return fmt.Errorf("写入文件失败: %w", closeErr)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("保存文件失败: %w", err)
	}
	return nil
}

func (c *localClient) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dest, err := c.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

func (c *localClient) URL(key string) string {
	key, err := cleanKey(key)
	if err != nil {
		return ""
	}
	return LocalURLPrefix + "/" + key
}

func (c *localClient) path(key string) (string, error) {
	key, err := cleanKey(key)
	if err != nil {
		return "", err
	}
	return filepath.Join(c.root, filepath.FromSlash(key)), nil
}
