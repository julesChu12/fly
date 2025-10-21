package discovery

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// EnvDiscovery 基于环境变量的服务发现实现
// 适用于开发环境和简单部署场景
type EnvDiscovery struct {
	// 服务名到环境变量键的映射规则
	// 默认: {SERVICE_NAME}_HOST, {SERVICE_NAME}_PORT
	// 或者: {SERVICE_NAME}_ADDRESS (host:port格式)
}

// NewEnvDiscovery 创建环境变量服务发现实例
func NewEnvDiscovery() *EnvDiscovery {
	return &EnvDiscovery{}
}

// GetService 从环境变量获取服务实例
// 支持以下环境变量格式:
// 1. {SERVICE_NAME}_ADDRESS=host:port  (推荐)
// 2. {SERVICE_NAME}_HOST=host + {SERVICE_NAME}_PORT=port
// 3. {SERVICE_NAME}_GRPC_ADDRESS=host:port (兼容现有配置)
func (e *EnvDiscovery) GetService(ctx context.Context, serviceName string) (*ServiceInstance, error) {
	// 标准化服务名 (转大写，替换-为_)
	envKey := normalizeServiceName(serviceName)

	// 尝试读取 ADDRESS 格式
	if addr := os.Getenv(envKey + "_ADDRESS"); addr != "" {
		return parseAddress(serviceName, addr)
	}

	// 尝试读取 GRPC_ADDRESS 格式 (兼容现有配置)
	if addr := os.Getenv(envKey + "_GRPC_ADDRESS"); addr != "" {
		return parseAddress(serviceName, addr)
	}

	// 尝试读取 HOST + PORT 格式
	host := os.Getenv(envKey + "_HOST")
	portStr := os.Getenv(envKey + "_PORT")
	if host != "" && portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port for service %s: %w", serviceName, err)
		}
		return &ServiceInstance{
			ID:      fmt.Sprintf("%s-env", serviceName),
			Name:    serviceName,
			Host:    host,
			Port:    port,
			Healthy: true,
		}, nil
	}

	return nil, &ErrServiceNotFound{ServiceName: serviceName}
}

// GetServices 返回单个服务实例（环境变量模式只支持单实例）
func (e *EnvDiscovery) GetServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	instance, err := e.GetService(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	return []*ServiceInstance{instance}, nil
}

// Register 环境变量模式不支持注册
func (e *EnvDiscovery) Register(ctx context.Context, instance *ServiceInstance) error {
	return fmt.Errorf("EnvDiscovery does not support service registration")
}

// Deregister 环境变量模式不支持注销
func (e *EnvDiscovery) Deregister(ctx context.Context, instanceID string) error {
	return fmt.Errorf("EnvDiscovery does not support service deregistration")
}

// Close 关闭连接（无操作）
func (e *EnvDiscovery) Close() error {
	return nil
}

// normalizeServiceName 标准化服务名为环境变量键
// 例: "custos" -> "CUSTOS", "order-service" -> "ORDER_SERVICE"
func normalizeServiceName(serviceName string) string {
	normalized := strings.ToUpper(serviceName)
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}

// parseAddress 解析 host:port 格式的地址
func parseAddress(serviceName, address string) (*ServiceInstance, error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid address format for service %s: %s (expected host:port)", serviceName, address)
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port in address for service %s: %w", serviceName, err)
	}

	return &ServiceInstance{
		ID:      fmt.Sprintf("%s-env", serviceName),
		Name:    serviceName,
		Host:    parts[0],
		Port:    port,
		Healthy: true,
		Metadata: map[string]string{
			"source": "environment",
		},
	}, nil
}
