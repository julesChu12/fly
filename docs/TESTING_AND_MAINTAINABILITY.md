# Fly平台测试策略和可维护性指南

## 📋 概述

本文档提供Fly平台的完整测试策略、测试框架使用指南和代码可维护性最佳实践，确保系统质量和长期可维护性。

---

## 🧪 测试策略

### 测试金字塔

```
          E2E测试 (10%)
       ┌──────────────┐
       │  端到端测试   │  完整业务场景
       └──────────────┘
     ┌────────────────────┐
     │    集成测试 (30%)   │  服务间交互
     └────────────────────┘
   ┌──────────────────────────┐
   │    单元测试 (60%)         │  函数和方法级别
   └──────────────────────────┘
```

### 测试覆盖率目标

| 层次 | 覆盖率目标 | 说明 |
|------|-----------|------|
| 单元测试 | >80% | 覆盖核心业务逻辑 |
| 集成测试 | >60% | 覆盖服务间交互 |
| E2E测试 | >40% | 覆盖关键业务流程 |
| 总体覆盖率 | >75% | 综合测试覆盖率 |

---

## 🔬 单元测试

### 1. 测试框架选择

#### Go标准测试库 + Testify

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)
```

**优势**:
- 标准库原生支持
- Testify提供丰富的断言和Mock功能
- 社区广泛使用，文档完善

### 2. 实体层测试

#### 示例: Appointment实体测试

```go
package entity_test

import (
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/julesChu12/fly/appointments/internal/domain/entity"
)

// TestAppointment_StatusTransitions 测试状态转换逻辑
func TestAppointment_StatusTransitions(t *testing.T) {
    tests := []struct {
        name        string
        fromStatus  entity.AppointmentStatus
        toStatus    entity.AppointmentStatus
        expected    bool
        description string
    }{
        {
            name:        "Pending to Confirmed",
            fromStatus:  entity.AppointmentStatusPending,
            toStatus:    entity.AppointmentStatusConfirmed,
            expected:    true,
            description: "待确认预约可以转换为已确认",
        },
        {
            name:        "Pending to Cancelled",
            fromStatus:  entity.AppointmentStatusPending,
            toStatus:    entity.AppointmentStatusCancelled,
            expected:    true,
            description: "待确认预约可以取消",
        },
        {
            name:        "Pending to Completed",
            fromStatus:  entity.AppointmentStatusPending,
            toStatus:    entity.AppointmentStatusCompleted,
            expected:    false,
            description: "待确认预约不能直接完成",
        },
        {
            name:        "Completed to Cancelled",
            fromStatus:  entity.AppointmentStatusCompleted,
            toStatus:    entity.AppointmentStatusCancelled,
            expected:    false,
            description: "已完成预约不能取消",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            appointment := &entity.Appointment{
                ID:        uuid.New(),
                Status:    tt.fromStatus,
                StartTime: time.Now().Add(24 * time.Hour),
                EndTime:   time.Now().Add(25 * time.Hour),
            }

            // Act
            result := appointment.CanTransitionTo(tt.toStatus)

            // Assert
            assert.Equal(t, tt.expected, result, tt.description)
        })
    }
}

// TestAppointment_Duration 测试持续时间计算
func TestAppointment_Duration(t *testing.T) {
    // Arrange
    startTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
    endTime := time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC)

    appointment := &entity.Appointment{
        ID:        uuid.New(),
        StartTime: startTime,
        EndTime:   endTime,
    }

    // Act
    duration := appointment.Duration()

    // Assert
    expectedDuration := 90 * time.Minute
    assert.Equal(t, expectedDuration, duration, "预约持续时间应为90分钟")
}

// TestAppointment_IsInPast 测试过期检查
func TestAppointment_IsInPast(t *testing.T) {
    tests := []struct {
        name     string
        endTime  time.Time
        expected bool
    }{
        {
            name:     "Past appointment",
            endTime:  time.Now().Add(-1 * time.Hour),
            expected: true,
        },
        {
            name:     "Future appointment",
            endTime:  time.Now().Add(1 * time.Hour),
            expected: false,
        },
        {
            name:     "Current appointment",
            endTime:  time.Now().Add(5 * time.Minute),
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            appointment := &entity.Appointment{
                ID:        uuid.New(),
                StartTime: tt.endTime.Add(-1 * time.Hour),
                EndTime:   tt.endTime,
            }

            // Act
            result := appointment.IsInPast()

            // Assert
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 3. 服务层测试

#### Mock依赖注入

```go
package service_test

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/julesChu12/fly/appointments/internal/application/service"
    "github.com/julesChu12/fly/appointments/internal/domain/dto"
    "github.com/julesChu12/fly/appointments/internal/domain/entity"
)

// MockAppointmentRepository Mock仓储接口
type MockAppointmentRepository struct {
    mock.Mock
}

func (m *MockAppointmentRepository) Create(appointment *entity.Appointment) error {
    args := m.Called(appointment)
    return args.Error(0)
}

func (m *MockAppointmentRepository) GetByID(id string) (*entity.Appointment, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) Update(appointment *entity.Appointment) error {
    args := m.Called(appointment)
    return args.Error(0)
}

func (m *MockAppointmentRepository) Delete(id string) error {
    args := m.Called(id)
    return args.Error(0)
}

// TestAppointmentService_Create 测试预约创建
func TestAppointmentService_Create(t *testing.T) {
    // Arrange
    mockRepo := new(MockAppointmentRepository)
    mockLogger := new(MockLogger)

    appointmentService := service.NewAppointmentService(mockRepo, mockLogger)

    req := &dto.CreateAppointmentRequest{
        CustomerID: uuid.New(),
        StaffID:    uuid.New(),
        ServiceID:  uuid.New(),
        StartTime:  time.Now().Add(24 * time.Hour),
        EndTime:    time.Now().Add(25 * time.Hour),
        Notes:      stringPtr("测试预约"),
    }

    // 设置Mock期望
    mockRepo.On("Create", mock.AnythingOfType("*entity.Appointment")).
        Return(nil).
        Once()

    // Act
    ctx := context.Background()
    response, err := appointmentService.Create(ctx, req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.NotEqual(t, uuid.Nil, response.ID)
    assert.Equal(t, entity.AppointmentStatusPending, response.Status)

    // 验证Mock调用
    mockRepo.AssertExpectations(t)
}

// TestAppointmentService_Create_ValidationError 测试验证错误
func TestAppointmentService_Create_ValidationError(t *testing.T) {
    tests := []struct {
        name        string
        request     *dto.CreateAppointmentRequest
        expectedErr string
    }{
        {
            name: "Missing CustomerID",
            request: &dto.CreateAppointmentRequest{
                StaffID:   uuid.New(),
                ServiceID: uuid.New(),
                StartTime: time.Now().Add(24 * time.Hour),
                EndTime:   time.Now().Add(25 * time.Hour),
            },
            expectedErr: "customer_id is required",
        },
        {
            name: "Invalid Time Range",
            request: &dto.CreateAppointmentRequest{
                CustomerID: uuid.New(),
                StaffID:    uuid.New(),
                ServiceID:  uuid.New(),
                StartTime:  time.Now().Add(25 * time.Hour),
                EndTime:    time.Now().Add(24 * time.Hour),
            },
            expectedErr: "end_time must be after start_time",
        },
        {
            name: "Past Start Time",
            request: &dto.CreateAppointmentRequest{
                CustomerID: uuid.New(),
                StaffID:    uuid.New(),
                ServiceID:  uuid.New(),
                StartTime:  time.Now().Add(-1 * time.Hour),
                EndTime:    time.Now().Add(1 * time.Hour),
            },
            expectedErr: "start_time cannot be in the past",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockRepo := new(MockAppointmentRepository)
            mockLogger := new(MockLogger)
            appointmentService := service.NewAppointmentService(mockRepo, mockLogger)

            // Act
            ctx := context.Background()
            response, err := appointmentService.Create(ctx, tt.request)

            // Assert
            assert.Error(t, err)
            assert.Nil(t, response)
            assert.Contains(t, err.Error(), tt.expectedErr)

            // 验证Repository没有被调用
            mockRepo.AssertNotCalled(t, "Create")
        })
    }
}
```

### 4. 表驱动测试

```go
// TestCalculateServicePrice 表驱动测试示例
func TestCalculateServicePrice(t *testing.T) {
    tests := []struct {
        name          string
        serviceID     string
        expectedPrice float64
        description   string
    }{
        {
            name:          "Cardiology Consultation",
            serviceID:     "cardiology-consultation",
            expectedPrice: 300.00,
            description:   "心脏科咨询价格",
        },
        {
            name:          "General Consultation",
            serviceID:     "general-consultation",
            expectedPrice: 150.00,
            description:   "普通门诊价格",
        },
        {
            name:          "Unknown Service",
            serviceID:     "unknown-service",
            expectedPrice: 200.00,
            description:   "未知服务使用默认价格",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            service := NewOrderIntegrationService(nil, nil, nil, nil)

            // Act
            price := service.calculateServicePrice(tt.serviceID)

            // Assert
            assert.Equal(t, tt.expectedPrice, price, tt.description)
        })
    }
}
```

---

## 🔗 集成测试

### 1. 数据库集成测试

#### 使用测试数据库

```go
package integration_test

import (
    "testing"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "github.com/julesChu12/fly/appointments/internal/domain/entity"
    "github.com/julesChu12/fly/appointments/internal/infrastructure/database"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    // 自动迁移
    err = db.AutoMigrate(&entity.Appointment{})
    if err != nil {
        t.Fatalf("Failed to migrate test database: %v", err)
    }

    return db
}

// teardownTestDB 清理测试数据库
func teardownTestDB(t *testing.T, db *gorm.DB) {
    sqlDB, err := db.DB()
    if err != nil {
        t.Errorf("Failed to get database connection: %v", err)
        return
    }

    if err := sqlDB.Close(); err != nil {
        t.Errorf("Failed to close database connection: %v", err)
    }
}

// TestAppointmentRepository_Integration 仓储集成测试
func TestAppointmentRepository_Integration(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    repo := database.NewAppointmentRepository(db)

    t.Run("Create and Retrieve", func(t *testing.T) {
        // Arrange
        appointment := &entity.Appointment{
            ID:         uuid.New(),
            CustomerID: uuid.New(),
            StaffID:    uuid.New(),
            ServiceID:  uuid.New(),
            StartTime:  time.Now().Add(24 * time.Hour),
            EndTime:    time.Now().Add(25 * time.Hour),
            Status:     entity.AppointmentStatusPending,
        }

        // Act - Create
        err := repo.Create(appointment)
        assert.NoError(t, err)

        // Act - Retrieve
        retrieved, err := repo.GetByID(appointment.ID.String())

        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, retrieved)
        assert.Equal(t, appointment.ID, retrieved.ID)
        assert.Equal(t, appointment.CustomerID, retrieved.CustomerID)
        assert.Equal(t, appointment.Status, retrieved.Status)
    })

    t.Run("Update", func(t *testing.T) {
        // Arrange
        appointment := &entity.Appointment{
            ID:         uuid.New(),
            CustomerID: uuid.New(),
            StaffID:    uuid.New(),
            ServiceID:  uuid.New(),
            StartTime:  time.Now().Add(24 * time.Hour),
            EndTime:    time.Now().Add(25 * time.Hour),
            Status:     entity.AppointmentStatusPending,
        }

        err := repo.Create(appointment)
        assert.NoError(t, err)

        // Act - Update Status
        appointment.Status = entity.AppointmentStatusConfirmed
        err = repo.Update(appointment)
        assert.NoError(t, err)

        // Assert
        retrieved, err := repo.GetByID(appointment.ID.String())
        assert.NoError(t, err)
        assert.Equal(t, entity.AppointmentStatusConfirmed, retrieved.Status)
    })

    t.Run("Delete", func(t *testing.T) {
        // Arrange
        appointment := &entity.Appointment{
            ID:         uuid.New(),
            CustomerID: uuid.New(),
            StaffID:    uuid.New(),
            ServiceID:  uuid.New(),
            StartTime:  time.Now().Add(24 * time.Hour),
            EndTime:    time.Now().Add(25 * time.Hour),
            Status:     entity.AppointmentStatusPending,
        }

        err := repo.Create(appointment)
        assert.NoError(t, err)

        // Act - Delete
        err = repo.Delete(appointment.ID.String())
        assert.NoError(t, err)

        // Assert
        retrieved, err := repo.GetByID(appointment.ID.String())
        assert.Error(t, err)
        assert.Nil(t, retrieved)
    })
}
```

### 2. HTTP API集成测试

```go
package integration_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"

    "github.com/julesChu12/fly/appointments/internal/interface/http"
)

// setupTestRouter 设置测试路由
func setupTestRouter(t *testing.T) *gin.Engine {
    gin.SetMode(gin.TestMode)

    db := setupTestDB(t)
    repo := database.NewAppointmentRepository(db)
    service := service.NewAppointmentService(repo, logger)
    handler := http.NewAppointmentHandler(service)

    router := gin.Default()
    handler.RegisterRoutes(router)

    return router
}

// TestAppointmentAPI_Integration API集成测试
func TestAppointmentAPI_Integration(t *testing.T) {
    router := setupTestRouter(t)

    t.Run("POST /appointments - Success", func(t *testing.T) {
        // Arrange
        requestBody := map[string]interface{}{
            "customer_id": uuid.New().String(),
            "staff_id":    uuid.New().String(),
            "service_id":  uuid.New().String(),
            "start_time":  time.Now().Add(24 * time.Hour).Format(time.RFC3339),
            "end_time":    time.Now().Add(25 * time.Hour).Format(time.RFC3339),
            "notes":       "Test appointment",
        }

        jsonBody, _ := json.Marshal(requestBody)
        req, _ := http.NewRequest("POST", "/appointments", bytes.NewBuffer(jsonBody))
        req.Header.Set("Content-Type", "application/json")

        w := httptest.NewRecorder()

        // Act
        router.ServeHTTP(w, req)

        // Assert
        assert.Equal(t, http.StatusCreated, w.Code)

        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.NotEmpty(t, response["id"])
        assert.Equal(t, "pending", response["status"])
    })

    t.Run("POST /appointments - Validation Error", func(t *testing.T) {
        // Arrange
        requestBody := map[string]interface{}{
            "customer_id": "", // Empty customer_id
            "staff_id":    uuid.New().String(),
            "service_id":  uuid.New().String(),
        }

        jsonBody, _ := json.Marshal(requestBody)
        req, _ := http.NewRequest("POST", "/appointments", bytes.NewBuffer(jsonBody))
        req.Header.Set("Content-Type", "application/json")

        w := httptest.NewRecorder()

        // Act
        router.ServeHTTP(w, req)

        // Assert
        assert.Equal(t, http.StatusBadRequest, w.Code)

        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Contains(t, response["error"], "customer_id")
    })
}
```

---

## 🌐 E2E测试

### 1. 业务流程E2E测试

```go
package e2e_test

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"

    "github.com/julesChu12/fly/appointments/internal/application/service"
)

// AppointmentE2ETestSuite E2E测试套件
type AppointmentE2ETestSuite struct {
    suite.Suite
    appointmentService *service.AppointmentService
    orderService       *service.OrderIntegrationService
    eventBus           *events.EventBus
}

// SetupSuite 测试套件初始化
func (suite *AppointmentE2ETestSuite) SetupSuite() {
    // 初始化真实的服务依赖
    db := setupTestDB(suite.T())
    repo := database.NewAppointmentRepository(db)
    logger := setupTestLogger()
    eventBus := events.NewEventBus()

    suite.appointmentService = service.NewAppointmentService(repo, logger)
    suite.orderService = service.NewOrderIntegrationService(suite.appointmentService, nil, nil, eventBus, logger)
    suite.eventBus = eventBus
}

// TestCompleteAppointmentFlow 测试完整预约流程
func (suite *AppointmentE2ETestSuite) TestCompleteAppointmentFlow() {
    ctx := context.Background()

    // 步骤1: 创建预约
    suite.T().Log("步骤1: 创建预约")
    createReq := &dto.CreateAppointmentRequest{
        CustomerID: uuid.New(),
        StaffID:    uuid.New(),
        ServiceID:  uuid.New(),
        StartTime:  time.Now().Add(24 * time.Hour),
        EndTime:    time.Now().Add(25 * time.Hour),
        Notes:      stringPtr("E2E测试预约"),
    }

    appointment, err := suite.orderService.CreateAppointmentWithOrder(ctx, createReq)
    suite.NoError(err)
    suite.NotNil(appointment)
    suite.Equal(entity.AppointmentStatusPending, appointment.Status)

    appointmentID := appointment.ID.String()

    // 步骤2: 确认预约
    suite.T().Log("步骤2: 确认预约")
    confirmed, err := suite.appointmentService.UpdateStatus(ctx, appointmentID, entity.AppointmentStatusConfirmed)
    suite.NoError(err)
    suite.Equal(entity.AppointmentStatusConfirmed, confirmed.Status)

    // 步骤3: 开始预约
    suite.T().Log("步骤3: 开始预约")
    inProgress, err := suite.appointmentService.UpdateStatus(ctx, appointmentID, entity.AppointmentStatusInProgress)
    suite.NoError(err)
    suite.Equal(entity.AppointmentStatusInProgress, inProgress.Status)

    // 步骤4: 完成预约
    suite.T().Log("步骤4: 完成预约")
    completed, err := suite.appointmentService.UpdateStatus(ctx, appointmentID, entity.AppointmentStatusCompleted)
    suite.NoError(err)
    suite.Equal(entity.AppointmentStatusCompleted, completed.Status)

    // 步骤5: 验证事件发布
    suite.T().Log("步骤5: 验证事件发布")
    time.Sleep(100 * time.Millisecond) // 等待异步事件处理

    stats := suite.eventBus.GetStats()
    suite.T().Logf("事件统计: %+v", stats)
    suite.Greater(stats.TotalEvents, int64(0), "应该有事件被发布")
}

// TestAppointmentCancellationFlow 测试预约取消流程
func (suite *AppointmentE2ETestSuite) TestAppointmentCancellationFlow() {
    ctx := context.Background()

    // 创建预约
    createReq := &dto.CreateAppointmentRequest{
        CustomerID: uuid.New(),
        StaffID:    uuid.New(),
        ServiceID:  uuid.New(),
        StartTime:  time.Now().Add(24 * time.Hour),
        EndTime:    time.Now().Add(25 * time.Hour),
    }

    appointment, err := suite.appointmentService.Create(ctx, createReq)
    suite.NoError(err)

    // 取消预约
    cancelled, err := suite.appointmentService.UpdateStatus(ctx, appointment.ID.String(), entity.AppointmentStatusCancelled)
    suite.NoError(err)
    suite.Equal(entity.AppointmentStatusCancelled, cancelled.Status)

    // 验证无法再次修改已取消的预约
    _, err = suite.appointmentService.UpdateStatus(ctx, appointment.ID.String(), entity.AppointmentStatusConfirmed)
    suite.Error(err, "已取消的预约不应该能被重新确认")
}

// 运行测试套件
func TestAppointmentE2ETestSuite(t *testing.T) {
    suite.Run(t, new(AppointmentE2ETestSuite))
}
```

---

## 📊 测试覆盖率

### 生成覆盖率报告

```bash
# 运行所有测试并生成覆盖率报告
go test ./... -coverprofile=coverage.out -covermode=atomic

# 查看总体覆盖率
go tool cover -func=coverage.out

# 生成HTML覆盖率报告
go tool cover -html=coverage.out -o coverage.html

# 只测试特定包
go test ./appointments/internal/domain/entity/... -cover

# 运行基准测试
go test -bench=. -benchmem ./...
```

### CI/CD集成

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21

      - name: Run Unit Tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload Coverage
        uses: codecov/codecov-action@v2
        with:
          file: ./coverage.out

      - name: Check Coverage Threshold
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 75" | bc -l) )); then
            echo "Coverage $coverage% is below 75% threshold"
            exit 1
          fi
```

---

## 🔧 可维护性最佳实践

### 1. 代码组织

#### 清晰的目录结构
```
appointments/
├── api/
│   └── proto/              # Protocol Buffers定义
├── cmd/
│   └── server/             # 应用入口
├── configs/                # 配置文件
├── internal/
│   ├── application/        # 应用层
│   │   └── service/        # 应用服务
│   ├── domain/             # 领域层
│   │   ├── entity/         # 实体
│   │   ├── dto/            # 数据传输对象
│   │   └── repository/     # 仓储接口
│   ├── infrastructure/     # 基础设施层
│   │   ├── database/       # 数据库实现
│   │   └── cache/          # 缓存实现
│   └── interface/          # 接口层
│       ├── http/           # HTTP处理器
│       └── grpc/           # gRPC处理器
├── pkg/                    # 共享包
└── test/                   # 测试辅助
```

### 2. 依赖注入

#### 使用依赖注入容器

```go
// pkg/di/container.go
package di

import (
    "github.com/google/wire"

    "github.com/julesChu12/fly/appointments/internal/application/service"
    "github.com/julesChu12/fly/appointments/internal/infrastructure/database"
)

// ProviderSet Wire依赖注入集合
var ProviderSet = wire.NewSet(
    // Database
    database.NewAppointmentRepository,

    // Services
    service.NewAppointmentService,
    service.NewEventService,
    service.NewOrderIntegrationService,

    // Handlers
    http.NewAppointmentHandler,
)

// InitializeAppointmentService 初始化预约服务
func InitializeAppointmentService(db *gorm.DB, logger *logger.Logger) (*service.AppointmentService, error) {
    wire.Build(ProviderSet)
    return nil, nil
}
```

### 3. 配置管理

#### 统一的配置结构

```go
// internal/config/config.go
package config

import (
    "time"

    "github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Cache    CacheConfig    `mapstructure:"cache"`
    Event    EventConfig    `mapstructure:"event"`
    Log      LogConfig      `mapstructure:"log"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigFile(configPath)
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
    // 配置验证逻辑
    return nil
}
```

### 4. 错误处理

#### 统一的错误处理模式

```go
// pkg/errors/handler.go
package errors

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

// HandleError HTTP错误处理中间件
func HandleError() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            var appErr *AppError
            if errors.As(err, &appErr) {
                c.JSON(getHTTPStatus(appErr.Code), ErrorResponse{
                    Code:    string(appErr.Code),
                    Message: appErr.Message,
                    Details: appErr.Context,
                })
                return
            }

            // 未知错误
            c.JSON(http.StatusInternalServerError, ErrorResponse{
                Code:    "INTERNAL_ERROR",
                Message: "Internal server error",
            })
        }
    }
}

func getHTTPStatus(code ErrorCode) int {
    switch code {
    case CodeInvalidArgument, CodeValidationError:
        return http.StatusBadRequest
    case CodeNotFound:
        return http.StatusNotFound
    case CodePermissionDenied, CodeAuthorizationError:
        return http.StatusForbidden
    case CodeAuthenticationError:
        return http.StatusUnauthorized
    case CodeTimeout:
        return http.StatusRequestTimeout
    case CodeRateLimited:
        return http.StatusTooManyRequests
    default:
        return http.StatusInternalServerError
    }
}
```

### 5. 文档化

#### 自动化API文档

```go
// 使用Swaggo生成API文档
// internal/interface/http/appointment_handler.go

// CreateAppointment godoc
// @Summary      创建预约
// @Description  创建新的预约记录
// @Tags         appointments
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateAppointmentRequest  true  "预约创建请求"
// @Success      201      {object}  dto.AppointmentResponse
// @Failure      400      {object}  errors.ErrorResponse
// @Failure      500      {object}  errors.ErrorResponse
// @Router       /appointments [post]
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
    // 实现逻辑
}
```

---

## 🎯 测试最佳实践总结

### 1. 测试命名规范
- 使用描述性的测试名称
- 格式: `Test<功能>_<场景>_<期望结果>`
- 例如: `TestAppointmentService_Create_Success`

### 2. AAA模式
- **Arrange**: 准备测试数据和依赖
- **Act**: 执行被测试的操作
- **Assert**: 验证结果

### 3. 测试隔离
- 每个测试应该独立运行
- 使用Setup和Teardown清理状态
- 避免测试间的依赖

### 4. Mock使用
- 只Mock外部依赖
- 不要过度Mock
- 使用接口提供可测试性

### 5. 持续集成
- 每次提交运行全部测试
- 强制覆盖率阈值
- 自动化测试报告

---

## 📝 维护性检查清单

- [ ] 所有公共API都有文档注释
- [ ] 关键业务逻辑有单元测试
- [ ] 复杂流程有集成测试
- [ ] 测试覆盖率 >75%
- [ ] 错误处理完整且一致
- [ ] 日志记录充分且结构化
- [ ] 配置文件有说明
- [ ] 代码通过lint检查
- [ ] 依赖注入正确使用
- [ ] 接口设计清晰合理

通过遵循这些测试策略和可维护性实践，可以确保Fly平台的代码质量和长期可维护性。