package oss

import (
	"bedrock/pkg/storage"
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Config 阿里云 OSS 专属配置
type Config struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string

	// Domain 是自定义域名（例如绑定了 CDN 的域名）
	// 如果为空，则默认使用阿里云的标准 Bucket 域名
	Domain string
}

type Provider struct {
	bucket *oss.Bucket
	config Config
}

// NewProvider 初始化阿里云 OSS 客户端
func NewProvider(c Config) (*Provider, error) {
	// 1. 创建 Client
	client, err := oss.New(c.Endpoint, c.AccessKeyID, c.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss init client failed: %w", err)
	}

	// 2. 获取 Bucket 实例 (注意：这里只是获取句柄，不会发生网络请求，除非调用 exists)
	bucket, err := client.Bucket(c.BucketName)
	if err != nil {
		return nil, fmt.Errorf("oss get bucket failed: %w", err)
	}

	return &Provider{
		bucket: bucket,
		config: c,
	}, nil
}

// 编译期检查接口实现
var _ storage.Provider = (*Provider)(nil)

// Upload 上传文件
func (p *Provider) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	// 组装 Options
	// oss.WithContext 让 SDK 感知上下文（超时/取消）
	options := []oss.Option{
		oss.WithContext(ctx),
		oss.ContentType("application/octet-stream"), // 最好由上层根据文件名推断并传入
	}

	// 阿里云建议明确设置 ContentLength，否则可能会采用分片上传或内存缓冲
	if size > 0 {
		options = append(options, oss.ContentLength(size))
	}

	// 执行上传
	err := p.bucket.PutObject(key, reader, options...)
	if err != nil {
		return "", fmt.Errorf("oss upload failed: %w", err)
	}

	// 返回访问链接
	return p.buildPublicURL(key), nil
}

// Delete 删除文件
func (p *Provider) Delete(ctx context.Context, key string) error {
	err := p.bucket.DeleteObject(key, oss.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("oss delete failed: %w", err)
	}
	return nil
}

// GetPrivateURL 获取私有文件的临时签名 URL
func (p *Provider) GetPrivateURL(ctx context.Context, key string, expire int64) (string, error) {
	// SignURL 默认也是支持 options 的，但通常签名操作是在本地计算，不涉及网络请求
	// 如果需要处理自定义域名的签名，可能需要额外处理，这里使用默认逻辑
	signedURL, err := p.bucket.SignURL(key, oss.HTTPGet, expire)
	if err != nil {
		return "", fmt.Errorf("oss sign url failed: %w", err)
	}

	// 注意：如果你配置了 CDN Domain，SignURL 返回的可能还是 oss 源站域名
	// 如果需要 CDN 鉴权，逻辑会完全不同（需要 URL 鉴权算法），这里仅返回 OSS 原生签名 URL
	return signedURL, nil
}

// buildPublicURL 组装公开访问的 URL
func (p *Provider) buildPublicURL(key string) string {
	// 如果配置了自定义域名 (CDN)，直接拼接
	if p.config.Domain != "" {
		// 简单处理路径拼接，防止出现双斜杠
		// 生产环境建议使用 path.Join 或 net/url 处理
		return fmt.Sprintf("%s/%s", p.config.Domain, key)
	}

	// 默认格式: https://<bucket>.<endpoint>/<key>
	// 需要处理 Endpoint 中可能包含的 http:// 前缀
	return fmt.Sprintf("https://%s.%s/%s", p.config.BucketName, removeProtocol(p.config.Endpoint), key)
}

// 辅助函数：移除 endpoint 中的 http:// 或 https://
func removeProtocol(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err == nil && u.Scheme != "" {
		return u.Host
	}
	return endpoint
}
