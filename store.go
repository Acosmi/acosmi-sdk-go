package acosmi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TokenStore token 持久化接口
// 桌面智能体可自行实现 (如 macOS Keychain / Windows Credential Manager)
type TokenStore interface {
	Save(tokens *TokenSet) error
	Load() (*TokenSet, error)
	Clear() error
}

// FileTokenStore 基于文件的 token 存储 (开发/测试用)
// 生产环境建议替换为系统钥匙串实现
type FileTokenStore struct {
	path string
	mu   sync.Mutex
}

// NewFileTokenStore 创建文件 token 存储
// 默认路径: ~/.acosmi/tokens.json (合并后统一路径)
// [RC-10] 返回 error, 防止 os.UserHomeDir() 失败时路径变为 /.acosmi/
func NewFileTokenStore(path string) (*FileTokenStore, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(home, ".acosmi", "tokens.json")
	}
	return &FileTokenStore{path: path}, nil
}

func (s *FileTokenStore) Save(tokens *TokenSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tokens: %w", err)
	}

	return os.WriteFile(s.path, data, 0600)
}

func (s *FileTokenStore) Load() (*TokenSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read token file: %w", err)
	}

	var tokens TokenSet
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("unmarshal tokens: %w", err)
	}
	return &tokens, nil
}

// [RC-11] 文件不存在时不传播错误 (Logout 后 Clear 不应报错)
func (s *FileTokenStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
