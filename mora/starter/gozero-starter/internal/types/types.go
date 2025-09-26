package types

// 基础响应类型
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 健康检查
type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// 登录相关
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
}

// 用户资料
type ProfileResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Subject  string `json:"subject"`
	Exp      string `json:"exp"`
	Iat      string `json:"iat"`
}

// 受保护端点
type ProtectedResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	Time    string `json:"time"`
}

// 订单相关
type Order struct {
	ID     string  `json:"id"`
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
}

type OrdersResponse struct {
	Orders []Order `json:"orders"`
	Total  int     `json:"total"`
}

type CreateOrderRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type CreateOrderResponse struct {
	Order Order `json:"order"`
}

// 用户列表
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type UsersResponse struct {
	Users     []User `json:"users"`
	Total     int    `json:"total"`
	RequestBy string `json:"request_by"`
}