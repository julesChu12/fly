package constants

const (
	ServiceName = "hermes"
	Version     = "v1.0.0"

	DefaultPageSize = 20
	MaxPageSize     = 100
)

const (
	ContactTypePhone = "phone"
	ContactTypeEmail = "email"
	ContactTypeAddr  = "address"
	ContactTypeWechat = "wechat"
)

const (
	HeaderTenantID = "X-Tenant-ID"
	HeaderUserID   = "X-User-ID"
	HeaderTraceID  = "X-Trace-ID"
)

// Context key types to avoid collisions (SA1029)
type contextKey string

// Context keys
var (
	ContextKeyTenantID = contextKey("tenant_id")
	ContextKeyUserID   = contextKey("user_id")
	ContextKeyTraceID  = contextKey("trace_id")
)