package local

import (
	"bedrock/pkg/storage"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Config 本地存储配置
type Config struct {
	RootPath string // 文件存储的物理根目录，例如: "./uploads"
	BaseURL  string // 外网访问的基础URL，例如: "http://localhost:8080/static"
}

type Provider struct {
	config Config
}

func NewProvider(c Config) storage.Provider {
	// 确保根目录是绝对路径（可选，视项目习惯而定）
	absPath, err := filepath.Abs(c.RootPath)
	if err != nil {
		return nil
	}

	// 确保根目录存在，不存在则创建
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil
	}

	// 更新为绝对路径，防止后续工作目录切换导致问题
	c.RootPath = absPath
	return &Provider{config: c}
}

var _ storage.Provider = (*Provider)(nil)

func (p *Provider) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	// 1. 安全检查：防止 key 包含 "../" 进行目录遍历攻击
	if strings.Contains(key, "..") {
		return "", fmt.Errorf("invalid key: path traversal attempt")
	}

	// 2. 拼接完整的物理路径
	fullPath := filepath.Join(p.config.RootPath, key)

	// 3. 确保该文件所在的父目录存在 (例如 key="avatars/user1.jpg"，需确保 "avatars" 目录存在)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("local mkdir failed: %w", err)
	}

	// 4. 创建文件
	// 注意：os.Create 不支持 Context 取消。如果需要支持 Cancel，逻辑会很复杂（需自行处理 goroutine）
	// 对于本地文件写入，通常速度很快，暂不实现复杂的 Context 控制
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("local create file failed: %w", err)
	}
	defer dst.Close()

	// 5. 写入内容
	if _, err := io.Copy(dst, reader); err != nil {
		return "", fmt.Errorf("local write failed: %w", err)
	}

	// 6. 拼接返回 URL
	// 使用 path.Join 而不是 filepath.Join，因为 URL 必须使用 "/" 分隔符
	// 假设 BaseURL 没有尾部斜杠，key 也不带头部斜杠
	urlPath := strings.TrimRight(p.config.BaseURL, "/") + "/" + strings.TrimLeft(key, "/")
	return urlPath, nil
}

func (p *Provider) Delete(ctx context.Context, key string) error {
	// 安全检查
	if strings.Contains(key, "..") {
		return fmt.Errorf("invalid key")
	}

	fullPath := filepath.Join(p.config.RootPath, key)

	err := os.Remove(fullPath)
	// 如果文件本来就不存在，通常视为删除成功，不报错
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local delete failed: %w", err)
	}
	return nil
}

func (p *Provider) GetPrivateURL(ctx context.Context, key string, expire int64) (string, error) {
	// 本地存储通常是静态文件服务，很难实现真正的“带签名的临时 URL”
	// 简单实现：直接返回公开 URL，或者报错表示不支持
	// 这里为了开发方便，直接返回公开链接
	return p.Upload(ctx, key, nil, 0) // 复用 URL 拼接逻辑，但不做上传
	// 或者: return "", fmt.Errorf("local storage does not support signed URLs")
}
