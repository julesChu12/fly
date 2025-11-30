package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"

	itemsv1 "github.com/julesChu12/fly/items/api/proto/items/v1"
)

func main() {
	// 连接到 gRPC 服务器
	conn, err := grpc.Dial("localhost:15056", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	fmt.Println("🚀 Connected to gRPC server on localhost:15056")

	// 创建客户端
	itemClient := itemsv1.NewItemServiceClient(conn)
	categoryClient := itemsv1.NewCategoryServiceClient(conn)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试结果
	var passedTests int
	var totalTests int

	// 测试 1: 创建商品 (Item)
	fmt.Println("\n📝 Test 1: Creating an Item...")
	if testCreateItem(ctx, itemClient) {
		passedTests++
	}
	totalTests++

	// 测试 2: 创建分类 (Category)
	fmt.Println("\n📂 Test 2: Creating a Category...")
	if testCreateCategory(ctx, categoryClient) {
		passedTests++
	}
	totalTests++

	// 测试 3: 列出分类
	fmt.Println("\n📋 Test 3: Listing Categories...")
	if testListCategories(ctx, categoryClient) {
		passedTests++
	}
	totalTests++

	// 测试 4: 列出商品
	fmt.Println("\n🛍️ Test 4: Listing Items...")
	if testListItems(ctx, itemClient) {
		passedTests++
	}
	totalTests++

	// 测试 5: 获取统计信息
	fmt.Println("\n📊 Test 5: Getting Item Stats...")
	if testGetItemStats(ctx, itemClient) {
		passedTests++
	}
	totalTests++

	// 总结
	fmt.Printf("\n\n🎯 Test Results: %d/%d passed (%.1f%%)\n",
		passedTests, totalTests, float64(passedTests)/float64(totalTests)*100)

	if passedTests == totalTests {
		fmt.Println("✅ All tests passed! gRPC services are working correctly.")
	} else {
		fmt.Printf("❌ %d tests failed. Please check the service logs.\n", totalTests-passedTests)
	}
}

func testCreateItem(ctx context.Context, client itemsv1.ItemServiceClient) bool {
	req := &itemsv1.CreateItemRequest{
		Name:        "Premium Massage Service",
		Description: "Professional 60-minute full body massage",
		Type:        itemsv1.ItemType_ITEM_TYPE_SERVICE,
			Price:       299.99,
		CategoryId:  "test-category-uuid",
		ImageUrl:    "https://example.com/massage.jpg",
		Tags:        "massage,premium,spa",
		Duration:    wrapperspb.Int32(60),
		StaffRequired: wrapperspb.Bool(true),
		Capacity:    wrapperspb.Int32(1),
	}

	resp, err := client.CreateItem(ctx, req)
	if err != nil {
		fmt.Printf("❌ CreateItem failed: %v\n", err)
		return false
	}

	fmt.Printf("✅ CreateItem success! Item ID: %s, Name: %s\n",
		resp.Item.Id, resp.Item.Name)
	return true
}

func testCreateCategory(ctx context.Context, client itemsv1.CategoryServiceClient) bool {
	req := &itemsv1.CreateCategoryRequest{
		Name:        "Massage Services",
		Description: "Professional massage and spa services",
		ParentId:    "",
		Icon:        "spa-icon",
		SortOrder:   1,
	}

	resp, err := client.CreateCategory(ctx, req)
	if err != nil {
		fmt.Printf("❌ CreateCategory failed: %v\n", err)
		return false
	}

	fmt.Printf("✅ CreateCategory success! Category ID: %s, Name: %s\n",
		resp.Category.Id, resp.Category.Name)
	return true
}

func testListCategories(ctx context.Context, client itemsv1.CategoryServiceClient) bool {
	req := &itemsv1.ListCategoriesRequest{
		Page:     1,
		PageSize: 10,
		Status:   itemsv1.CategoryStatus_CATEGORY_STATUS_ACTIVE,
	}

	resp, err := client.ListCategories(ctx, req)
	if err != nil {
		fmt.Printf("❌ ListCategories failed: %v\n", err)
		return false
	}

	fmt.Printf("✅ ListCategories success! Found %d categories, Total: %d\n",
		len(resp.Categories), resp.Total)
	for i, cat := range resp.Categories {
		fmt.Printf("   %d. %s (%s)\n", i+1, cat.Name, cat.Id)
	}
	return true
}

func testListItems(ctx context.Context, client itemsv1.ItemServiceClient) bool {
	req := &itemsv1.ListItemsRequest{
		Page:     1,
		PageSize: 10,
		Status:   itemsv1.ItemStatus_ITEM_STATUS_ACTIVE,
	}

	resp, err := client.ListItems(ctx, req)
	if err != nil {
		fmt.Printf("❌ ListItems failed: %v\n", err)
		return false
	}

	fmt.Printf("✅ ListItems success! Found %d items, Total: %d\n",
		len(resp.Items), resp.Total)
	return true
}

func testGetItemStats(ctx context.Context, client itemsv1.ItemServiceClient) bool {
	req := &itemsv1.GetItemStatsRequest{}

	resp, err := client.GetItemStats(ctx, req)
	if err != nil {
		fmt.Printf("❌ GetItemStats failed: %v\n", err)
		return false
	}

	fmt.Printf("✅ GetItemStats success! Total Items: %d, Active Items: %d\n",
		resp.TotalItems, resp.ActiveItems)
	fmt.Printf("   Services: %d, Products: %d\n", resp.TotalServices, resp.TotalProducts)
	return true
}