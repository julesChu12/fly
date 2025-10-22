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

// Context keys
const (
	ContextKeyTenantID = "tenant_id"
	ContextKeyUserID   = "user_id"
	ContextKeyTraceID  = "trace_id"
)