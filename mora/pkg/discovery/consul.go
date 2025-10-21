package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulDiscovery 基于 Consul 的服务发现实现
type ConsulDiscovery struct {
	client     *api.Client
	balancer   LoadBalancer
	datacenter string
}

// ConsulConfig Consul 配置
type ConsulConfig struct {
	Address    string        // Consul 地址，默认 "127.0.0.1:8500"
	Datacenter string        // 数据中心名称，默认 "dc1"
	Token      string        // ACL Token (可选)
	Timeout    time.Duration // 超时时间，默认 10s
}

// NewConsulDiscovery 创建 Consul 服务发现实例
func NewConsulDiscovery(cfg *ConsulConfig) (*ConsulDiscovery, error) {
	if cfg == nil {
		cfg = &ConsulConfig{
			Address:    "127.0.0.1:8500",
			Datacenter: "dc1",
			Timeout:    10 * time.Second,
		}
	}

	config := api.DefaultConfig()
	config.Address = cfg.Address
	config.Token = cfg.Token
	config.WaitTime = cfg.Timeout

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulDiscovery{
		client:     client,
		balancer:   NewRoundRobinBalancer(),
		datacenter: cfg.Datacenter,
	}, nil
}

// GetService 从 Consul 获取一个健康的服务实例（带负载均衡）
func (c *ConsulDiscovery) GetService(ctx context.Context, serviceName string) (*ServiceInstance, error) {
	instances, err := c.GetServices(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	if len(instances) == 0 {
		return nil, &ErrNoHealthyInstance{ServiceName: serviceName}
	}

	// 使用负载均衡器选择实例
	return c.balancer.Select(instances)
}

// GetServices 从 Consul 获取所有健康的服务实例
func (c *ConsulDiscovery) GetServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	// 查询健康的服务实例
	entries, _, err := c.client.Health().Service(serviceName, "", true, &api.QueryOptions{
		Datacenter: c.datacenter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query service %s from consul: %w", serviceName, err)
	}

	if len(entries) == 0 {
		return nil, &ErrServiceNotFound{ServiceName: serviceName}
	}

	instances := make([]*ServiceInstance, 0, len(entries))
	for _, entry := range entries {
		instance := &ServiceInstance{
			ID:       entry.Service.ID,
			Name:     entry.Service.Service,
			Host:     entry.Service.Address,
			Port:     entry.Service.Port,
			Metadata: entry.Service.Meta,
			Healthy:  true, // 已经通过健康检查过滤
		}

		// 如果 Address 为空，使用 Node 的地址
		if instance.Host == "" {
			instance.Host = entry.Node.Address
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// Register 向 Consul 注册服务实例
func (c *ConsulDiscovery) Register(ctx context.Context, instance *ServiceInstance) error {
	registration := &api.AgentServiceRegistration{
		ID:      instance.ID,
		Name:    instance.Name,
		Address: instance.Host,
		Port:    instance.Port,
		Meta:    instance.Metadata,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", instance.Host, instance.Port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service %s to consul: %w", instance.Name, err)
	}

	return nil
}

// Deregister 从 Consul 注销服务实例
func (c *ConsulDiscovery) Deregister(ctx context.Context, instanceID string) error {
	if err := c.client.Agent().ServiceDeregister(instanceID); err != nil {
		return fmt.Errorf("failed to deregister service %s from consul: %w", instanceID, err)
	}
	return nil
}

// Close 关闭 Consul 客户端
func (c *ConsulDiscovery) Close() error {
	// Consul client 不需要显式关闭
	return nil
}

// SetLoadBalancer 设置负载均衡策略
func (c *ConsulDiscovery) SetLoadBalancer(lb LoadBalancer) {
	c.balancer = lb
}
