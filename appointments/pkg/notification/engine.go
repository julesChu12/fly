package notification

import (
	"context"
	"sync"
	"time"
)

// ChannelType 通知渠道类型
type ChannelType string

const (
	ChannelEmail   ChannelType = "email"   // 邮件
	ChannelSMS     ChannelType = "sms"     // 短信
	ChannelPush    ChannelType = "push"    // 推送通知
	ChannelWeChat  ChannelType = "wechat"  // 微信
	ChannelWebhook ChannelType = "webhook" // Webhook
	ChannelInApp   ChannelType = "in_app"  // 应用内通知
	ChannelSystem  ChannelType = "system"  // 系统通知
)

// Priority 通知优先级
type Priority string

const (
	PriorityLow    Priority = "low"    // 低优先级
	PriorityNormal Priority = "normal" // 普通优先级
	PriorityHigh   Priority = "high"   // 高优先级
	PriorityUrgent Priority = "urgent" // 紧急
)

// Status 通知状态
type Status string

const (
	StatusPending   Status = "pending"   // 待发送
	StatusSending   Status = "sending"   // 发送中
	StatusSent      Status = "sent"      // 已发送
	StatusFailed    Status = "failed"    // 发送失败
	StatusCancelled Status = "cancelled" // 已取消
	StatusRetrying  Status = "retrying"  // 重试中
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeText       MessageType = "text"       // 文本消息
	MessageTypeHTML       MessageType = "html"       // HTML消息
	MessageTypeMarkdown   MessageType = "markdown"   // Markdown消息
	MessageTypeTemplate   MessageType = "template"   // 模板消息
	MessageTypeAttachment MessageType = "attachment" // 附件消息
)

// Notification 通知对象
type Notification struct {
	ID          string                 `json:"id"`           // 通知ID
	Title       string                 `json:"title"`        // 标题
	Content     string                 `json:"content"`      // 内容
	Type        MessageType            `json:"type"`         // 消息类型
	Channels    []ChannelType          `json:"channels"`     // 通知渠道
	Recipients  []*Recipient           `json:"recipients"`   // 接收人列表
	Priority    Priority               `json:"priority"`     // 优先级
	Status      Status                 `json:"status"`       // 状态
	Data        map[string]interface{} `json:"data"`         // 附加数据
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据
	Templates   map[string]string      `json:"templates"`    // 模板映射
	Attachments []*Attachment          `json:"attachments"`  // 附件列表
	CreatedAt   time.Time              `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time              `json:"updated_at"`   // 更新时间
	ScheduledAt *time.Time             `json:"scheduled_at"` // 计划发送时间
	ExpiresAt   *time.Time             `json:"expires_at"`   // 过期时间
	MaxRetries  int                    `json:"max_retries"`  // 最大重试次数
	RetryCount  int                    `json:"retry_count"`  // 当前重试次数
	SentAt      *time.Time             `json:"sent_at"`      // 发送时间
	CompletedAt *time.Time             `json:"completed_at"` // 完成时间

	// 运行时属性
	mu sync.RWMutex `json:"-"`
}

// Recipient 接收人
type Recipient struct {
	ID       string                 `json:"id"`        // 接收人ID
	Type     string                 `json:"type"`      // 接收人类型 (user, admin, system)
	Name     string                 `json:"name"`      // 姓名
	Email    string                 `json:"email"`     // 邮箱
	Phone    string                 `json:"phone"`     // 手机号
	WeChatID string                 `json:"wechat_id"` // 微信ID
	DeviceID string                 `json:"device_id"` // 设备ID
	Channels []ChannelType          `json:"channels"`  // 偏好的通知渠道
	Settings map[string]interface{} `json:"settings"`  // 个人设置
}

// Attachment 附件
type Attachment struct {
	ID       string                 `json:"id"`        // 附件ID
	Name     string                 `json:"name"`      // 文件名
	Type     string                 `json:"type"`      // 文件类型 (image, document, video, audio)
	URL      string                 `json:"url"`       // 文件URL
	Size     int64                  `json:"size"`      // 文件大小
	MimeType string                 `json:"mime_type"` // MIME类型
	Metadata map[string]interface{} `json:"metadata"`  // 附件元数据
}

// DeliveryResult 发送结果
type DeliveryResult struct {
	NotificationID string                 `json:"notification_id"` // 通知ID
	Channel        ChannelType            `json:"channel"`         // 渠道
	RecipientID    string                 `json:"recipient_id"`    // 接收人ID
	Status         Status                 `json:"status"`          // 状态
	Message        string                 `json:"message"`         // 结果消息
	Error          error                  `json:"error"`           // 错误信息
	SentAt         time.Time              `json:"sent_at"`         // 发送时间
	DeliveryTime   time.Duration          `json:"delivery_time"`   // 投递时间
	Metadata       map[string]interface{} `json:"metadata"`        // 附加信息
}

// NotificationEngineConfig 通知引擎配置
type NotificationEngineConfig struct {
	MaxConcurrency    int           `yaml:"max_concurrency"`    // 最大并发数
	DefaultTimeout    time.Duration `yaml:"default_timeout"`    // 默认超时时间
	MaxRetries        int           `yaml:"max_retries"`        // 最大重试次数
	RetryDelay        time.Duration `yaml:"retry_delay"`        // 重试延迟
	EnableMetrics     bool          `yaml:"enable_metrics"`     // 启用指标
	EnableTracing     bool          `yaml:"enable_tracing"`     // 启用跟踪
	BatchSize         int           `yaml:"batch_size"`         // 批处理大小
	BatchTimeout      time.Duration `yaml:"batch_timeout"`      // 批处理超时
	QueueSize         int           `yaml:"queue_size"`         // 队列大小
	EnablePersistence bool          `yaml:"enable_persistence"` // 启用持久化
	CleanupInterval   time.Duration `yaml:"cleanup_interval"`   // 清理间隔
	RetentionPeriod   time.Duration `yaml:"retention_period"`   // 保留期限

	// 渠道配置
	ChannelConfigs map[ChannelType]*ChannelConfig `yaml:"channel_configs"`

	// 模板配置
	TemplateConfig *TemplateConfig `yaml:"template_config"`
}

// ChannelConfig 渠道配置
type ChannelConfig struct {
	Enabled        bool                   `yaml:"enabled"`         // 是否启用
	RateLimit      int                    `yaml:"rate_limit"`      // 速率限制
	BurstSize      int                    `yaml:"burst_size"`      // 突发大小
	Timeout        time.Duration          `yaml:"timeout"`         // 超时时间
	MaxRetries     int                    `yaml:"max_retries"`     // 最大重试次数
	RetryDelay     time.Duration          `yaml:"retry_delay"`     // 重试延迟
	Config         map[string]interface{} `yaml:"config"`          // 渠道特定配置
	WebhookURL     string                 `yaml:"webhook_url"`     // Webhook URL
	APIKey         string                 `yaml:"api_key"`         // API密钥
	SecretKey      string                 `yaml:"secret_key"`      // 密钥
	TemplateEngine string                 `yaml:"template_engine"` // 模板引擎
}

// TemplateConfig 模板配置
type TemplateConfig struct {
	Engine      string                 `yaml:"engine"`       // 模板引擎
	Directory   string                 `yaml:"directory"`    // 模板目录
	Suffix      string                 `yaml:"suffix"`       // 模板文件后缀
	Functions   map[string]string      `yaml:"functions"`    // 自定义函数
	DefaultData map[string]interface{} `yaml:"default_data"` // 默认数据
}

// DefaultNotificationEngineConfig 默认配置
func DefaultNotificationEngineConfig() *NotificationEngineConfig {
	return &NotificationEngineConfig{
		MaxConcurrency:    100,
		DefaultTimeout:    30 * time.Second,
		MaxRetries:        3,
		RetryDelay:        5 * time.Second,
		EnableMetrics:     true,
		EnableTracing:     true,
		BatchSize:         50,
		BatchTimeout:      10 * time.Second,
		QueueSize:         1000,
		EnablePersistence: true,
		CleanupInterval:   1 * time.Hour,
		RetentionPeriod:   30 * 24 * time.Hour,
		ChannelConfigs:    make(map[ChannelType]*ChannelConfig),
		TemplateConfig: &TemplateConfig{
			Engine:      "go_template",
			Directory:   "templates/notification",
			Suffix:      ".tmpl",
			Functions:   make(map[string]string),
			DefaultData: make(map[string]interface{}),
		},
	}
}

// NotificationEngine 通知引擎接口
type NotificationEngine interface {
	// SendNotification 发送通知
	SendNotification(ctx context.Context, notification *Notification) ([]*DeliveryResult, error)

	// SendBatch 批量发送通知
	SendBatch(ctx context.Context, notifications []*Notification) ([]*DeliveryResult, error)

	// ScheduleNotification 计划发送通知
	ScheduleNotification(ctx context.Context, notification *Notification, scheduleTime time.Time) error

	// CancelNotification 取消通知
	CancelNotification(ctx context.Context, notificationID string) error

	// GetNotification 获取通知
	GetNotification(ctx context.Context, notificationID string) (*Notification, error)

	// ListNotifications 列出通知
	ListNotifications(ctx context.Context, filter *NotificationFilter) ([]*Notification, error)

	// GetDeliveryResults 获取发送结果
	GetDeliveryResults(ctx context.Context, notificationID string) ([]*DeliveryResult, error)

	// GetStats 获取统计信息
	GetStats(ctx context.Context) (*NotificationStats, error)

	// RegisterChannel 注册渠道
	RegisterChannel(channel ChannelType, handler ChannelHandler) error

	// GetChannel 获取渠道
	GetChannel(channel ChannelType) (ChannelHandler, error)

	// EnableChannel 启用渠道
	EnableChannel(channel ChannelType) error

	// DisableChannel 禁用渠道
	DisableChannel(channel ChannelType) error

	// ReloadTemplates 重新加载模板
	ReloadTemplates(ctx context.Context) error

	// Cleanup 清理过期数据
	Cleanup(ctx context.Context) error
}

// ChannelHandler 渠道处理器接口
type ChannelHandler interface {
	// GetType 获取渠道类型
	GetType() ChannelType

	// Send 发送通知
	Send(ctx context.Context, notification *Notification, recipient *Recipient) (*DeliveryResult, error)

	// SendBatch 批量发送
	SendBatch(ctx context.Context, notification *Notification, recipients []*Recipient) ([]*DeliveryResult, error)

	// Validate 验证配置
	Validate(config *ChannelConfig) error

	// Initialize 初始化
	Initialize(config *ChannelConfig) error

	// Shutdown 关闭
	Shutdown() error

	// GetStats 获取渠道统计
	GetStats() (*ChannelStats, error)
}

// NotificationFilter 通知过滤器
type NotificationFilter struct {
	Status      Status      `json:"status,omitempty"`
	Channel     ChannelType `json:"channel,omitempty"`
	Priority    Priority    `json:"priority,omitempty"`
	RecipientID string      `json:"recipient_id,omitempty"`
	StartTime   *time.Time  `json:"start_time,omitempty"`
	EndTime     *time.Time  `json:"end_time,omitempty"`
	Limit       int         `json:"limit,omitempty"`
	Offset      int         `json:"offset,omitempty"`
}

// NotificationStats 通知统计信息
type NotificationStats struct {
	TotalNotifications   int64                         `json:"total_notifications"`      // 总通知数
	PendingNotifications int64                         `json:"pending_notifications"`    // 待发送通知
	SentNotifications    int64                         `json:"sent_notifications"`       // 已发送通知
	FailedNotifications  int64                         `json:"failed_notifications"`     // 失败通知
	AverageDeliveryTime  float64                       `json:"average_delivery_time_ms"` // 平均投递时间
	ChannelStats         map[ChannelType]*ChannelStats `json:"channel_stats"`            // 渠道统计
	PriorityDistribution map[Priority]int64            `json:"priority_distribution"`    // 优先级分布
	TypeDistribution     map[MessageType]int64         `json:"type_distribution"`        // 类型分布
	HourlyStats          []int64                       `json:"hourly_stats"`             // 24小时统计
	DailyStats           []int64                       `json:"daily_stats"`              // 7天统计
	LastCleanupTime      time.Time                     `json:"last_cleanup_time"`        // 最后清理时间
}

// ChannelStats 渠道统计信息
type ChannelStats struct {
	Channel             ChannelType `json:"channel"`                  // 渠道类型
	TotalSent           int64       `json:"total_sent"`               // 总发送数
	SuccessfulSends     int64       `json:"successful_sends"`         // 成功发送数
	FailedSends         int64       `json:"failed_sends"`             // 失败发送数
	AverageDeliveryTime float64     `json:"average_delivery_time_ms"` // 平均投递时间
	LastSendTime        time.Time   `json:"last_send_time"`           // 最后发送时间
	IsEnabled           bool        `json:"is_enabled"`               // 是否启用
	RateLimit           int         `json:"rate_limit"`               // 速率限制
	QueueSize           int         `json:"queue_size"`               // 队列大小
}

// TemplateEngine 模板引擎接口
type TemplateEngine interface {
	// Render 渲染模板
	Render(template string, data map[string]interface{}) (string, error)

	// RenderFile 渲染模板文件
	RenderFile(filename string, data map[string]interface{}) (string, error)

	// LoadTemplates 加载模板
	LoadTemplates(directory string) error

	// Reload 重新加载模板
	Reload() error

	// AddFunction 添加自定义函数
	AddFunction(name string, fn interface{}) error

	// Validate 验证模板
	Validate(template string) error
}
