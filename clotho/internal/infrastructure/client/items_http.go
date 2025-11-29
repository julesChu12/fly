package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// ItemsHTTPClient 封装对 Items HTTP API 的调用
type ItemsHTTPClient struct {
	client  *http.Client
	baseURL string
	logger  *logger.Logger
}

// NewItemsHTTPClient 创建新的 Items HTTP 客户端
func NewItemsHTTPClient(baseURL string, timeout time.Duration, logger *logger.Logger) *ItemsHTTPClient {
	return &ItemsHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		logger:  logger,
	}
}

// Item 商品数据结构
type Item struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"` // SERVICE or PRODUCT
	Price       float64    `json:"price"`
	CategoryID  string     `json:"category_id"`
	Status      string     `json:"status"`
	ImageURL    *string    `json:"image_url,omitempty"`
	Tags        *string    `json:"tags,omitempty"`

	// 服务字段
	Duration      *int   `json:"duration,omitempty"`
	StaffRequired *bool  `json:"staff_required,omitempty"`
	Capacity      *int   `json:"capacity,omitempty"`

	// 产品字段
	Stock     *int     `json:"stock,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Weight    *float64 `json:"weight,omitempty"`
	SKU       *string  `json:"sku,omitempty"`

	// 时间戳
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Category 分类数据结构
type Category struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ParentID    *string    `json:"parent_id,omitempty"`
	Icon        *string    `json:"icon,omitempty"`
	SortOrder   int        `json:"sort_order"`
	Status      string     `json:"status"`
	ItemCount   int        `json:"item_count"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

// CategoryTree 分类树结构
type CategoryTree struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ParentID    *string        `json:"parent_id,omitempty"`
	Icon        *string        `json:"icon,omitempty"`
	SortOrder   int            `json:"sort_order"`
	Status      string         `json:"status"`
	ItemCount   int            `json:"item_count"`
	Level       int            `json:"level"`
	Path        string         `json:"path"`
	Children    []*CategoryTree `json:"children,omitempty"`
}

// CreateItemRequest 创建商品请求
type CreateItemRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	Price       float64    `json:"price"`
	CategoryID  string     `json:"category_id"`
	ImageURL    *string    `json:"image_url,omitempty"`
	Tags        *string    `json:"tags,omitempty"`

	// 服务字段
	Duration      *int   `json:"duration,omitempty"`
	StaffRequired *bool  `json:"staff_required,omitempty"`
	Capacity      *int   `json:"capacity,omitempty"`

	// 产品字段
	Stock     *int     `json:"stock,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Weight    *float64 `json:"weight,omitempty"`
	SKU       *string  `json:"sku,omitempty"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	SortOrder   int     `json:"sort_order"`
}

// ListItemsResponse 商品列表响应
type ListItemsResponse struct {
	Items      []*Item `json:"items"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	Size       int     `json:"size"`
}

// GetItemsRequest 获取商品列表请求
type GetItemsRequest struct {
	Type       *string `json:"type,omitempty"`
	Status     *string `json:"status,omitempty"`
	CategoryID *string `json:"category_id,omitempty"`
	MinPrice   *float64 `json:"min_price,omitempty"`
	MaxPrice   *float64 `json:"max_price,omitempty"`
	Search     *string `json:"search,omitempty"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
}

// SearchItemsRequest 搜索商品请求
type SearchItemsRequest struct {
	Query      string  `json:"query"`
	Type       *string `json:"type,omitempty"`
	CategoryID *string `json:"category_id,omitempty"`
	MinPrice   *float64 `json:"min_price,omitempty"`
	MaxPrice   *float64 `json:"max_price,omitempty"`
	Limit      int     `json:"limit,omitempty"`
}

// BatchDeleteItemsRequest 批量删除商品请求
type BatchDeleteItemsRequest struct {
	IDs []string `json:"ids"`
}

// BatchUpdateItemsRequest 批量更新商品请求
type BatchUpdateItemsRequest struct {
	IDs           []string    `json:"ids"`
	Status        *string     `json:"status,omitempty"`
	Price         *float64    `json:"price,omitempty"`
	CategoryID    *string     `json:"category_id,omitempty"`
	IsActive      *bool       `json:"is_active,omitempty"`
}

// === 商品相关方法 ===

// CreateItem 创建商品
func (c *ItemsHTTPClient) CreateItem(ctx context.Context, req *CreateItemRequest) (*Item, error) {
	var resp Item
	err := c.postRequest(ctx, "/api/v1/items", req, &resp)
	return &resp, err
}

// GetItems 获取商品列表
func (c *ItemsHTTPClient) GetItems(ctx context.Context, req *GetItemsRequest) (*ListItemsResponse, error) {
	var resp ListItemsResponse
	query := c.buildItemsQuery(req)
	path := "/api/v1/items" + query
	err := c.getRequest(ctx, path, &resp)
	return &resp, err
}

// GetItemByID 根据ID获取商品
func (c *ItemsHTTPClient) GetItemByID(ctx context.Context, id string) (*Item, error) {
	var resp Item
	path := fmt.Sprintf("/api/v1/items/%s", id)
	err := c.getRequest(ctx, path, &resp)
	return &resp, err
}

// UpdateItem 更新商品
func (c *ItemsHTTPClient) UpdateItem(ctx context.Context, id string, req *CreateItemRequest) (*Item, error) {
	var resp Item
	path := fmt.Sprintf("/api/v1/items/%s", id)
	err := c.putRequest(ctx, path, req, &resp)
	return &resp, err
}

// DeleteItem 删除商品
func (c *ItemsHTTPClient) DeleteItem(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/items/%s", id)
	return c.deleteRequest(ctx, path, nil)
}

// SearchItems 搜索商品
func (c *ItemsHTTPClient) SearchItems(ctx context.Context, req *SearchItemsRequest) ([]*Item, error) {
	var resp struct {
		Items []*Item `json:"items"`
	}
	query := c.buildSearchQuery(req)
	path := "/api/v1/search/items" + query
	err := c.getRequest(ctx, path, &resp)
	return resp.Items, err
}

// === 分类相关方法 ===

// CreateCategory 创建分类
func (c *ItemsHTTPClient) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*Category, error) {
	var resp Category
	err := c.postRequest(ctx, "/api/v1/categories", req, &resp)
	return &resp, err
}

// GetCategories 获取分类列表
func (c *ItemsHTTPClient) GetCategories(ctx context.Context) ([]*Category, error) {
	var resp struct {
		Categories []*Category `json:"categories"`
	}
	err := c.getRequest(ctx, "/api/v1/categories", &resp)
	return resp.Categories, err
}

// GetCategoryTree 获取分类树
func (c *ItemsHTTPClient) GetCategoryTree(ctx context.Context) ([]*CategoryTree, error) {
	var resp struct {
		Tree []*CategoryTree `json:"tree"`
	}
	err := c.getRequest(ctx, "/api/v1/categories/tree", &resp)
	return resp.Tree, err
}

// GetCategoryByID 根据ID获取分类
func (c *ItemsHTTPClient) GetCategoryByID(ctx context.Context, id string) (*Category, error) {
	var resp Category
	path := fmt.Sprintf("/api/v1/categories/%s", id)
	err := c.getRequest(ctx, path, &resp)
	return &resp, err
}

// UpdateCategory 更新分类
func (c *ItemsHTTPClient) UpdateCategory(ctx context.Context, id string, req *CreateCategoryRequest) (*Category, error) {
	var resp Category
	path := fmt.Sprintf("/api/v1/categories/%s", id)
	err := c.putRequest(ctx, path, req, &resp)
	return &resp, err
}

// DeleteCategory 删除分类
func (c *ItemsHTTPClient) DeleteCategory(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/categories/%s", id)
	return c.deleteRequest(ctx, path, nil)
}

// === HTTP 请求方法 ===

func (c *ItemsHTTPClient) postRequest(ctx context.Context, path string, reqBody interface{}, respBody interface{}) error {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal request body", "error", err)
		return fmt.Errorf("marshal request: %w", err)
	}

	url, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		c.logger.Error("Failed to join URL path", "baseURL", c.baseURL, "path", path, "error", err)
		return fmt.Errorf("join URL path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Error("Failed to create request", "error", err)
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, respBody)
}

func (c *ItemsHTTPClient) getRequest(ctx context.Context, path string, respBody interface{}) error {
	url, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		c.logger.Error("Failed to join URL path", "baseURL", c.baseURL, "path", path, "error", err)
		return fmt.Errorf("join URL path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Error("Failed to create request", "error", err)
		return fmt.Errorf("create request: %w", err)
	}

	return c.doRequest(req, respBody)
}

func (c *ItemsHTTPClient) putRequest(ctx context.Context, path string, reqBody interface{}, respBody interface{}) error {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal request body", "error", err)
		return fmt.Errorf("marshal request: %w", err)
	}

	url, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		c.logger.Error("Failed to join URL path", "baseURL", c.baseURL, "path", path, "error", err)
		return fmt.Errorf("join URL path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Error("Failed to create request", "error", err)
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, respBody)
}

func (c *ItemsHTTPClient) deleteRequest(ctx context.Context, path string, headers map[string]string) error {
	url, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		c.logger.Error("Failed to join URL path", "baseURL", c.baseURL, "path", path, "error", err)
		return fmt.Errorf("join URL path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		c.logger.Error("Failed to create request", "error", err)
		return fmt.Errorf("create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.doRequest(req, nil)
}

func (c *ItemsHTTPClient) doRequest(req *http.Request, respBody interface{}) error {
	c.logger.Info("Sending HTTP request", "method", req.Method, "url", req.URL.String())

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Failed to send request", "error", err)
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", "error", err)
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Error("HTTP request failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			c.logger.Error("Failed to unmarshal response", "error", err, "body", string(body))
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	c.logger.Info("HTTP request successful", "status", resp.StatusCode)
	return nil
}

// === 辅助方法 ===

func (c *ItemsHTTPClient) buildItemsQuery(req *GetItemsRequest) string {
	params := url.Values{}

	if req.Type != nil {
		params.Add("type", *req.Type)
	}
	if req.Status != nil {
		params.Add("status", *req.Status)
	}
	if req.CategoryID != nil {
		params.Add("category_id", *req.CategoryID)
	}
	if req.MinPrice != nil {
		params.Add("min_price", fmt.Sprintf("%.2f", *req.MinPrice))
	}
	if req.MaxPrice != nil {
		params.Add("max_price", fmt.Sprintf("%.2f", *req.MaxPrice))
	}
	if req.Search != nil {
		params.Add("search", *req.Search)
	}
	if req.Page > 0 {
		params.Add("page", fmt.Sprintf("%d", req.Page))
	}
	if req.PageSize > 0 {
		params.Add("page_size", fmt.Sprintf("%d", req.PageSize))
	}

	return "?" + params.Encode()
}

func (c *ItemsHTTPClient) buildSearchQuery(req *SearchItemsRequest) string {
	params := url.Values{}

	params.Add("q", req.Query)

	if req.Type != nil {
		params.Add("type", *req.Type)
	}
	if req.CategoryID != nil {
		params.Add("category_id", *req.CategoryID)
	}
	if req.MinPrice != nil {
		params.Add("min_price", fmt.Sprintf("%.2f", *req.MinPrice))
	}
	if req.MaxPrice != nil {
		params.Add("max_price", fmt.Sprintf("%.2f", *req.MaxPrice))
	}
	if req.Limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", req.Limit))
	}

	return "?" + params.Encode()
}

// === 批量操作方法 ===

// BatchDeleteItems 批量删除商品
func (c *ItemsHTTPClient) BatchDeleteItems(ctx context.Context, ids []string) error {
	req := &BatchDeleteItemsRequest{
		IDs: ids,
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	path := "/api/v1/items/batch"
	err := c.postRequest(ctx, path, req, &resp)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf("items service error: %s", resp.Message)
	}

	return nil
}

// BatchUpdateItems 批量更新商品
func (c *ItemsHTTPClient) BatchUpdateItems(ctx context.Context, req *BatchUpdateItemsRequest) ([]Item, error) {
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []Item `json:"data"`
	}

	path := "/api/v1/items/batch"
	err := c.putRequest(ctx, path, req, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("items service error: %s", resp.Message)
	}

	return resp.Data, nil
}