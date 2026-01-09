package s3

import (
	"bedrock/pkg/storage"
	"context"
	"fmt"
	"io"

	"time"

	// 引入 minio 官方库，它对 S3 协议的支持非常好且易用
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config 专门针对 S3 的配置
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	Region          string // AWS 必须，MinIO 可选
	Domain          string // 自定义外部访问域名（比如 CDN）
}

type Provider struct {
	client *minio.Client
	config Config
}

// NewProvider 显式构造函数
func NewProvider(c Config) storage.Provider {
	client, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKeyID, c.SecretAccessKey, ""),
		Secure: c.UseSSL,
		Region: c.Region,
	})
	if err != nil {
		return nil
	}

	return &Provider{
		client: client,
		config: c,
	}
}

// 编译时检查：确保 Provider 实现了 storage.Provider 接口
var _ storage.Provider = (*Provider)(nil)

func (p *Provider) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	// 自动处理分片上传逻辑
	_, err := p.client.PutObject(ctx, p.config.BucketName, key, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream", // 建议由上层业务探测并传入，这里简写
	})
	if err != nil {
		return "", err
	}

	// 如果配置了自定义域名（CDN），则拼接 CDN 域名
	if p.config.Domain != "" {
		return fmt.Sprintf("%s/%s", p.config.Domain, key), nil
	}
	// 否则返回 path
	return key, nil
}

func (p *Provider) Delete(ctx context.Context, key string) error {
	return p.client.RemoveObject(ctx, p.config.BucketName, key, minio.RemoveObjectOptions{})
}

func (p *Provider) GetPrivateURL(ctx context.Context, key string, expire int64) (string, error) {
	u, err := p.client.PresignedGetObject(ctx, p.config.BucketName, key, time.Duration(expire)*time.Second, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
