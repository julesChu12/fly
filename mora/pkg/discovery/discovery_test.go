package discovery

import (
	"context"
	"os"
	"testing"
)

func TestEnvDiscovery_GetService(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		envSetup    map[string]string
		wantHost    string
		wantPort    int
		wantErr     bool
	}{
		{
			name:        "使用 ADDRESS 格式",
			serviceName: "custos",
			envSetup: map[string]string{
				"CUSTOS_ADDRESS": "custos:9001",
			},
			wantHost: "custos",
			wantPort: 9001,
			wantErr:  false,
		},
		{
			name:        "使用 GRPC_ADDRESS 格式（兼容）",
			serviceName: "custos",
			envSetup: map[string]string{
				"CUSTOS_GRPC_ADDRESS": "custos:9001",
			},
			wantHost: "custos",
			wantPort: 9001,
			wantErr:  false,
		},
		{
			name:        "使用 HOST + PORT 格式",
			serviceName: "hermes",
			envSetup: map[string]string{
				"HERMES_HOST": "hermes",
				"HERMES_PORT": "9080",
			},
			wantHost: "hermes",
			wantPort: 9080,
			wantErr:  false,
		},
		{
			name:        "服务名包含横杠",
			serviceName: "order-service",
			envSetup: map[string]string{
				"ORDER_SERVICE_ADDRESS": "orders:9002",
			},
			wantHost: "orders",
			wantPort: 9002,
			wantErr:  false,
		},
		{
			name:        "服务未配置",
			serviceName: "unknown",
			envSetup:    map[string]string{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置环境变量
			for k, v := range tt.envSetup {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			discovery := NewEnvDiscovery()
			instance, err := discovery.GetService(context.Background(), tt.serviceName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if instance.Host != tt.wantHost {
					t.Errorf("GetService() host = %v, want %v", instance.Host, tt.wantHost)
				}
				if instance.Port != tt.wantPort {
					t.Errorf("GetService() port = %v, want %v", instance.Port, tt.wantPort)
				}
				if instance.Name != tt.serviceName {
					t.Errorf("GetService() name = %v, want %v", instance.Name, tt.serviceName)
				}
			}
		})
	}
}

func TestRoundRobinBalancer(t *testing.T) {
	instances := []*ServiceInstance{
		{ID: "1", Name: "test", Host: "host1", Port: 8080, Healthy: true},
		{ID: "2", Name: "test", Host: "host2", Port: 8080, Healthy: true},
		{ID: "3", Name: "test", Host: "host3", Port: 8080, Healthy: true},
	}

	balancer := NewRoundRobinBalancer()

	// 测试轮询
	selected := make(map[string]int)
	for i := 0; i < 9; i++ {
		instance, err := balancer.Select(instances)
		if err != nil {
			t.Fatalf("Select() error = %v", err)
		}
		selected[instance.ID]++
	}

	// 每个实例应该被选中 3 次
	for id, count := range selected {
		if count != 3 {
			t.Errorf("Instance %s selected %d times, want 3", id, count)
		}
	}
}

func TestFilterHealthy(t *testing.T) {
	instances := []*ServiceInstance{
		{ID: "1", Healthy: true},
		{ID: "2", Healthy: false},
		{ID: "3", Healthy: true},
		{ID: "4", Healthy: false},
	}

	healthy := filterHealthy(instances)

	if len(healthy) != 2 {
		t.Errorf("filterHealthy() returned %d instances, want 2", len(healthy))
	}

	for _, instance := range healthy {
		if !instance.Healthy {
			t.Errorf("filterHealthy() returned unhealthy instance %s", instance.ID)
		}
	}
}
