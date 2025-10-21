package discovery

import (
	"context"
	"fmt"
)

// ServiceInstance 代表一个服务实例
type ServiceInstance struct {
	ID       string            // 实例唯一标识
	Name     string            // 服务名称
	Host     string            // 主机地址
	Port     int               // 端口
	Metadata map[string]string // 元数据（如版本、分组等）
	Healthy  bool              // 健康状态
}

// Address 返回完整的服务地址
func (s *ServiceInstance) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Discovery 服务发现接口
type Discovery interface {
	// GetService 根据服务名获取一个可用的服务实例（负载均衡）
	GetService(ctx context.Context, serviceName string) (*ServiceInstance, error)

	// GetServices 根据服务名获取所有可用的服务实例
	GetServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error)

	// Register 注册服务实例
	Register(ctx context.Context, instance *ServiceInstance) error

	// Deregister 注销服务实例
	Deregister(ctx context.Context, instanceID string) error

	// Close 关闭服务发现客户端
	Close() error
}

// HealthChecker 健康检查接口
type HealthChecker interface {
	// Check 执行健康检查
	Check(ctx context.Context) error
}

// LoadBalancer 负载均衡策略
type LoadBalancer interface {
	// Select 从多个实例中选择一个
	Select(instances []*ServiceInstance) (*ServiceInstance, error)
}

// ErrServiceNotFound 服务未找到错误
type ErrServiceNotFound struct {
	ServiceName string
}

func (e *ErrServiceNotFound) Error() string {
	return fmt.Sprintf("service not found: %s", e.ServiceName)
}

// ErrNoHealthyInstance 无健康实例错误
type ErrNoHealthyInstance struct {
	ServiceName string
}

func (e *ErrNoHealthyInstance) Error() string {
	return fmt.Sprintf("no healthy instance for service: %s", e.ServiceName)
}
