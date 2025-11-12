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

// ContextKey type to avoid collisions (SA1029)
type ContextKey string

// Context keys
var (
	ContextKeyTenantID = ContextKey("tenant_id")
	ContextKeyUserID   = ContextKey("user_id")
	ContextKeyTraceID  = ContextKey("trace_id")
)

// HTTP headers
const (
	HeaderTenantID = "X-Tenant-ID"
	HeaderUserID   = "X-User-ID"
	HeaderTraceID  = "X-Trace-ID"
)