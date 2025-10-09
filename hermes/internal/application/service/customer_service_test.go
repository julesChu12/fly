package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/julesChu12/fly/hermes/internal/domain/entity"
	"github.com/julesChu12/fly/hermes/pkg/errors"
	"github.com/julesChu12/fly/hermes/pkg/types"
)

// MockCustomerRepository 模拟客户存储库
type MockCustomerRepository struct {
	customers map[uint]*entity.Customer
	nextID    uint
}

func NewMockCustomerRepository() *MockCustomerRepository {
	return &MockCustomerRepository{
		customers: make(map[uint]*entity.Customer),
		nextID:    1,
	}
}

func (m *MockCustomerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	customer.ID = m.nextID
	m.nextID++
	m.customers[customer.ID] = customer
	return nil
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, id uint) (*entity.Customer, error) {
	customer, exists := m.customers[id]
	if !exists {
		return nil, &errors.NotFoundError{Message: "customer not found"}
	}
	return customer, nil
}

func (m *MockCustomerRepository) GetByEmail(ctx context.Context, email string) (*entity.Customer, error) {
	for _, customer := range m.customers {
		if customer.Email == email {
			return customer, nil
		}
	}
	return nil, &errors.NotFoundError{Message: "customer not found"}
}

func (m *MockCustomerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	if _, exists := m.customers[customer.ID]; !exists {
		return &errors.NotFoundError{Message: "customer not found"}
	}
	m.customers[customer.ID] = customer
	return nil
}

func (m *MockCustomerRepository) Delete(ctx context.Context, id uint) error {
	if _, exists := m.customers[id]; !exists {
		return &errors.NotFoundError{Message: "customer not found"}
	}
	delete(m.customers, id)
	return nil
}

func (m *MockCustomerRepository) List(ctx context.Context, offset, limit int) ([]*entity.Customer, int64, error) {
	var customers []*entity.Customer
	for _, customer := range m.customers {
		customers = append(customers, customer)
	}

	total := int64(len(customers))

	// 简单的分页逻辑
	start := offset
	end := offset + limit
	if start > len(customers) {
		return []*entity.Customer{}, total, nil
	}
	if end > len(customers) {
		end = len(customers)
	}

	return customers[start:end], total, nil
}

func (m *MockCustomerRepository) GetWithContacts(ctx context.Context, id uint) (*entity.Customer, error) {
	return m.GetByID(ctx, id)
}

// MockContactRepository 模拟联系方式存储库
type MockContactRepository struct {
	contacts map[uint][]*entity.Contact
	nextID   uint
}

func NewMockContactRepository() *MockContactRepository {
	return &MockContactRepository{
		contacts: make(map[uint][]*entity.Contact),
		nextID:   1,
	}
}

func (m *MockContactRepository) Create(ctx context.Context, contact *entity.Contact) error {
	contact.ID = m.nextID
	m.nextID++
	m.contacts[contact.CustomerID] = append(m.contacts[contact.CustomerID], contact)
	return nil
}

func (m *MockContactRepository) GetByID(ctx context.Context, id uint) (*entity.Contact, error) {
	for _, contacts := range m.contacts {
		for _, contact := range contacts {
			if contact.ID == id {
				return contact, nil
			}
		}
	}
	return nil, &errors.NotFoundError{Message: "contact not found"}
}

func (m *MockContactRepository) GetByCustomerID(ctx context.Context, customerID uint) ([]*entity.Contact, error) {
	contacts, exists := m.contacts[customerID]
	if !exists {
		return []*entity.Contact{}, nil
	}
	return contacts, nil
}

func (m *MockContactRepository) Update(ctx context.Context, contact *entity.Contact) error {
	for _, contacts := range m.contacts {
		for i, c := range contacts {
			if c.ID == contact.ID {
				contacts[i] = contact
				return nil
			}
		}
	}
	return &errors.NotFoundError{Message: "contact not found"}
}

func (m *MockContactRepository) Delete(ctx context.Context, id uint) error {
	for customerID, contacts := range m.contacts {
		for i, contact := range contacts {
			if contact.ID == id {
				m.contacts[customerID] = append(contacts[:i], contacts[i+1:]...)
				return nil
			}
		}
	}
	return &errors.NotFoundError{Message: "contact not found"}
}

func (m *MockContactRepository) List(ctx context.Context, offset, limit int) ([]*entity.Contact, error) {
	var allContacts []*entity.Contact
	for _, contacts := range m.contacts {
		allContacts = append(allContacts, contacts...)
	}

	start := offset
	end := offset + limit
	if start > len(allContacts) {
		return []*entity.Contact{}, nil
	}
	if end > len(allContacts) {
		end = len(allContacts)
	}

	return allContacts[start:end], nil
}

// TestCustomerService_CreateCustomer 测试创建客户
func TestCustomerService_CreateCustomer(t *testing.T) {
	customerRepo := NewMockCustomerRepository()
	contactRepo := NewMockContactRepository()
	service := NewCustomerService(customerRepo, contactRepo)

	req := &types.CreateCustomerRequest{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Phone: "13800138000",
	}

	ctx := context.Background()
	customer, err := service.CreateCustomer(ctx, req)

	if err != nil {
		t.Fatalf("创建客户失败: %v", err)
	}

	if customer.Name != req.Name {
		t.Errorf("期望客户姓名为 %s, 但得到 %s", req.Name, customer.Name)
	}

	if customer.Email != req.Email {
		t.Errorf("期望客户邮箱为 %s, 但得到 %s", req.Email, customer.Email)
	}

	if customer.Phone != req.Phone {
		t.Errorf("期望客户电话为 %s, 但得到 %s", req.Phone, customer.Phone)
	}
}

// TestCustomerService_GetCustomer 测试获取客户
func TestCustomerService_GetCustomer(t *testing.T) {
	customerRepo := NewMockCustomerRepository()
	contactRepo := NewMockContactRepository()
	service := NewCustomerService(customerRepo, contactRepo)

	// 先创建一个客户
	req := &types.CreateCustomerRequest{
		Name:  "李四",
		Email: "lisi@example.com",
		Phone: "13900139000",
	}

	ctx := context.Background()
	created, err := service.CreateCustomer(ctx, req)
	if err != nil {
		t.Fatalf("创建客户失败: %v", err)
	}

	// 获取客户
	customer, err := service.GetCustomer(ctx, created.ID)
	if err != nil {
		t.Fatalf("获取客户失败: %v", err)
	}

	if customer.ID != created.ID {
		t.Errorf("期望客户ID为 %d, 但得到 %d", created.ID, customer.ID)
	}

	if customer.Name != req.Name {
		t.Errorf("期望客户姓名为 %s, 但得到 %s", req.Name, customer.Name)
	}
}

// TestCustomerService_GetCustomer_NotFound 测试获取不存在的客户
func TestCustomerService_GetCustomer_NotFound(t *testing.T) {
	customerRepo := NewMockCustomerRepository()
	contactRepo := NewMockContactRepository()
	service := NewCustomerService(customerRepo, contactRepo)

	ctx := context.Background()
	_, err := service.GetCustomer(ctx, 999)

	if err == nil {
		t.Fatal("期望返回错误，但没有返回错误")
	}

	if _, ok := err.(*errors.NotFoundError); !ok {
		t.Errorf("期望返回 NotFoundError, 但得到 %T", err)
	}
}

// TestCustomerService_UpdateCustomer 测试更新客户
func TestCustomerService_UpdateCustomer(t *testing.T) {
	customerRepo := NewMockCustomerRepository()
	contactRepo := NewMockContactRepository()
	service := NewCustomerService(customerRepo, contactRepo)

	// 先创建一个客户
	createReq := &types.CreateCustomerRequest{
		Name:  "王五",
		Email: "wangwu@example.com",
		Phone: "14000140000",
	}

	ctx := context.Background()
	created, err := service.CreateCustomer(ctx, createReq)
	if err != nil {
		t.Fatalf("创建客户失败: %v", err)
	}

	// 更新客户
	updateReq := &types.UpdateCustomerRequest{
		Name:  "王五更新",
		Email: "wangwu_updated@example.com",
		Phone: "14100141000",
	}

	updated, err := service.UpdateCustomer(ctx, created.ID, updateReq)
	if err != nil {
		t.Fatalf("更新客户失败: %v", err)
	}

	if updated.Name != updateReq.Name {
		t.Errorf("期望更新后客户姓名为 %s, 但得到 %s", updateReq.Name, updated.Name)
	}

	if updated.Email != updateReq.Email {
		t.Errorf("期望更新后客户邮箱为 %s, 但得到 %s", updateReq.Email, updated.Email)
	}
}

// TestCustomerService_DeleteCustomer 测试删除客户
func TestCustomerService_DeleteCustomer(t *testing.T) {
	customerRepo := NewMockCustomerRepository()
	contactRepo := NewMockContactRepository()
	service := NewCustomerService(customerRepo, contactRepo)

	// 先创建一个客户
	req := &types.CreateCustomerRequest{
		Name:  "赵六",
		Email: "zhaoliu@example.com",
		Phone: "15000150000",
	}

	ctx := context.Background()
	created, err := service.CreateCustomer(ctx, req)
	if err != nil {
		t.Fatalf("创建客户失败: %v", err)
	}

	// 删除客户
	err = service.DeleteCustomer(ctx, created.ID)
	if err != nil {
		t.Fatalf("删除客户失败: %v", err)
	}

	// 尝试再次获取，应该返回错误
	_, err = service.GetCustomer(ctx, created.ID)
	if err == nil {
		t.Fatal("期望获取已删除客户时返回错误，但没有返回错误")
	}
}

// TestCustomerService_ListCustomers 测试列表查询
func TestCustomerService_ListCustomers(t *testing.T) {
	customerRepo := NewMockCustomerRepository()
	contactRepo := NewMockContactRepository()
	service := NewCustomerService(customerRepo, contactRepo)

	ctx := context.Background()

	// 创建多个客户
	for i := 0; i < 5; i++ {
		req := &types.CreateCustomerRequest{
			Name:  fmt.Sprintf("客户%d", i+1),
			Email: fmt.Sprintf("customer%d@example.com", i+1),
			Phone: fmt.Sprintf("1300013%04d", i+1),
		}
		_, err := service.CreateCustomer(ctx, req)
		if err != nil {
			t.Fatalf("创建客户%d失败: %v", i+1, err)
		}
	}

	// 测试分页查询
	listReq := &types.ListRequest{
		Page:     1,
		PageSize: 3,
	}

	result, err := service.ListCustomers(ctx, listReq)
	if err != nil {
		t.Fatalf("查询客户列表失败: %v", err)
	}

	if result.Total != 5 {
		t.Errorf("期望总数为 5, 但得到 %d", result.Total)
	}

	if len(result.Data.([]types.CustomerResponse)) != 3 {
		t.Errorf("期望当前页数据数量为 3, 但得到 %d", len(result.Data.([]types.CustomerResponse)))
	}
}
