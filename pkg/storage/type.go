package storage

import (
	"context"
	"io"
)

// Provider 定义存储行为的标准接口
// 设计原则：依赖抽象，不依赖具体实现
//
//go:generate mockgen -source=./type.go -package=mocks -destination=./mocks/provider_mock.go Provider
type Provider interface {
	// Upload 上传文件
	// ctx:用于控制超时或取消
	// key: 存储路径 (不包含域名)，如 "avatar/1001.jpg"
	// reader: 使用 io.Reader 而不是 []byte，支持大文件流式上传，节省内存
	// size: 文件大小，某些 SDK（如 S3）在预知大小时能优化分片上传，如果未知可传 -1
	// 返回值: (相对路径或绝对URL, 错误)
	Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error)

	// Delete 删除文件
	Delete(ctx context.Context, key string) error

	// GetPrivateURL 获取私有文件的临时访问链接
	// expire: 有效期(秒)
	GetPrivateURL(ctx context.Context, key string, expire int64) (string, error)
}
