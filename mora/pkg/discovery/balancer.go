package discovery

import (
	"fmt"
	"math/rand"
	"sync/atomic"
)

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	counter uint64
}

// NewRoundRobinBalancer 创建轮询负载均衡器
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

// Select 使用轮询算法选择实例
func (r *RoundRobinBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances available")
	}

	// 过滤健康实例
	healthyInstances := filterHealthy(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("no healthy instances available")
	}

	// 原子递增计数器并取模
	index := atomic.AddUint64(&r.counter, 1) % uint64(len(healthyInstances))
	return healthyInstances[index], nil
}

// RandomBalancer 随机负载均衡器
type RandomBalancer struct{}

// NewRandomBalancer 创建随机负载均衡器
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{}
}

// Select 使用随机算法选择实例
func (r *RandomBalancer) Select(instances []*ServiceInstance) (*ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances available")
	}

	healthyInstances := filterHealthy(instances)
	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("no healthy instances available")
	}

	index := rand.Intn(len(healthyInstances))
	return healthyInstances[index], nil
}

// filterHealthy 过滤出健康的实例
func filterHealthy(instances []*ServiceInstance) []*ServiceInstance {
	healthy := make([]*ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		if instance.Healthy {
			healthy = append(healthy, instance)
		}
	}
	return healthy
}
