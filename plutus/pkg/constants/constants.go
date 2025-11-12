package constants

// Pagination constants
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Wallet constants
const (
	DefaultCurrency = "CNY"
	MinAmount       = 0.01
	MaxAmount       = 999999999.99
)

// Transaction constants
const (
	MaxIdempotencyKeyLength = 128
	MaxReferenceNoLength    = 128
)

// Context key types to avoid collisions (SA1029)
type contextKey string

// Context keys
var (
	ContextKeyTenantID = contextKey("tenant_id")
	ContextKeyUserID   = contextKey("user_id")
	ContextKeyTraceID  = contextKey("trace_id")
)

// HTTP headers
const (
	HeaderTenantID = "X-Tenant-ID"
	HeaderUserID   = "X-User-ID"
	HeaderTraceID  = "X-Trace-ID"
)