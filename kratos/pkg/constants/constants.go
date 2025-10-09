package constants

// Pagination constants
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Order constants
const (
	DefaultCurrency = "CNY"
	OrderNoPrefix   = "ORD"
)

// Context keys
const (
	ContextKeyTenantID = "tenant_id"
	ContextKeyUserID   = "user_id"
	ContextKeyTraceID  = "trace_id"
)

// HTTP headers
const (
	HeaderTenantID = "X-Tenant-ID"
	HeaderUserID   = "X-User-ID"
	HeaderTraceID  = "X-Trace-ID"
)