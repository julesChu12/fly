package handlers

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	itemsv1 "github.com/julesChu12/fly/items/api/proto/items/v1"
	"github.com/julesChu12/fly/items/internal/application/service"
	"github.com/julesChu12/fly/items/internal/domain/item"
	"github.com/julesChu12/fly/items/internal/interface/grpc/converters"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// ItemServer 实现 gRPC ItemService
type ItemServer struct {
	itemsv1.UnimplementedItemServiceServer
	itemService service.ItemService
	logger      *logger.Logger
}

// NewItemServer 创建新的 Item gRPC 服务器
func NewItemServer(itemService service.ItemService, logger *logger.Logger) *ItemServer {
	return &ItemServer{
		itemService: itemService,
		logger:      logger,
	}
}

// CreateItem 创建商品
func (s *ItemServer) CreateItem(ctx context.Context, req *itemsv1.CreateItemRequest) (*itemsv1.CreateItemResponse, error) {
	s.logger.Infof("CreateItem request received: %s", req.Name)

	// 转换请求
	createReq, err := converters.ProtoToCreateItemRequest(req)
	if err != nil {
		s.logger.Errorf("Failed to convert CreateItemRequest: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// 调用服务
	domainItem, err := s.itemService.CreateItem(ctx, createReq)
	if err != nil {
		s.logger.Errorf("Failed to create item: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create item: %v", err)
	}

	// 转换响应
	protoItem := converters.DomainItemToProto(domainItem)

	s.logger.Infof("Item created successfully: %s", protoItem.Id)
	return &itemsv1.CreateItemResponse{
		Item: protoItem,
	}, nil
}

// GetItem 获取商品详情
func (s *ItemServer) GetItem(ctx context.Context, req *itemsv1.GetItemRequest) (*itemsv1.GetItemResponse, error) {
	s.logger.Infof("GetItem request received: %s", req.Id)

	// 解析ID
	itemID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid item ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid item ID: %v", err)
	}

	// 调用服务
	domainItem, err := s.itemService.GetItemByID(ctx, itemID)
	if err != nil {
		if err == item.ErrItemNotFound {
			s.logger.Warnf("Item not found: %s", req.Id)
			return nil, status.Errorf(codes.NotFound, "item not found")
		}
		s.logger.Errorf("Failed to get item: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get item: %v", err)
	}

	// 转换响应
	protoItem := converters.DomainItemToProto(domainItem)

	s.logger.Infof("Item retrieved successfully: %s", protoItem.Id)
	return &itemsv1.GetItemResponse{
		Item: protoItem,
	}, nil
}

// UpdateItem 更新商品
func (s *ItemServer) UpdateItem(ctx context.Context, req *itemsv1.UpdateItemRequest) (*itemsv1.UpdateItemResponse, error) {
	s.logger.Infof("UpdateItem request received: %s", req.Id)

	// 解析ID
	itemID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid item ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid item ID: %v", err)
	}

	// 转换请求
	updateReq, err := converters.ProtoToUpdateItemRequest(req, req.Id)
	if err != nil {
		s.logger.Errorf("Failed to convert UpdateItemRequest: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// 调用服务
	domainItem, err := s.itemService.UpdateItem(ctx, itemID, updateReq)
	if err != nil {
		if err == item.ErrItemNotFound {
			s.logger.Warnf("Item not found for update: %s", req.Id)
			return nil, status.Errorf(codes.NotFound, "item not found")
		}
		s.logger.Errorf("Failed to update item: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update item: %v", err)
	}

	// 转换响应
	protoItem := converters.DomainItemToProto(domainItem)

	s.logger.Infof("Item updated successfully: %s", protoItem.Id)
	return &itemsv1.UpdateItemResponse{
		Item: protoItem,
	}, nil
}

// DeleteItem 删除商品
func (s *ItemServer) DeleteItem(ctx context.Context, req *itemsv1.DeleteItemRequest) (*emptypb.Empty, error) {
	s.logger.Infof("DeleteItem request received: %s", req.Id)

	// 解析ID
	itemID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid item ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid item ID: %v", err)
	}

	// 调用服务
	err = s.itemService.DeleteItem(ctx, itemID)
	if err != nil {
		if err == item.ErrItemNotFound {
			s.logger.Warnf("Item not found for deletion: %s", req.Id)
			return nil, status.Errorf(codes.NotFound, "item not found")
		}
		s.logger.Errorf("Failed to delete item: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete item: %v", err)
	}

	s.logger.Infof("Item deleted successfully: %s", req.Id)
	return &emptypb.Empty{}, nil
}

// ListItems 获取商品列表
func (s *ItemServer) ListItems(ctx context.Context, req *itemsv1.ListItemsRequest) (*itemsv1.ListItemsResponse, error) {
	s.logger.Infof("ListItems request received: page=%d, page_size=%d", req.Page, req.PageSize)

	// 解析分类ID（如果提供）
	var categoryID *uuid.UUID
	if req.CategoryId != "" {
		id, err := uuid.Parse(req.CategoryId)
		if err != nil {
			s.logger.Errorf("Invalid category ID: %v", err)
			return nil, status.Errorf(codes.InvalidArgument, "invalid category ID: %v", err)
		}
		categoryID = &id
	}

	// 构建请求 - 根据实际服务接口调整
	itemType := converters.ItemTypeFromProto(req.Type)
	itemStatus := converters.ItemStatusFromProto(req.Status)

	listReq := &service.ListItemsRequest{
		CategoryID: categoryID,
		Type:       &itemType,
		Status:     &itemStatus,
		Limit:      int(req.PageSize),
		Offset:     int(req.Page-1) * int(req.PageSize), // 转换为offset
	}

	// 调用服务
	items, total, err := s.itemService.ListItems(ctx, listReq)
	if err != nil {
		s.logger.Errorf("Failed to list items: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list items: %v", err)
	}

	// 转换响应
	protoItems := converters.BatchDomainItemsToProto(items)

	// 计算总页数
	totalPages := int32((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	s.logger.Infof("Items listed successfully: total=%d, returned=%d", total, len(protoItems))
	return &itemsv1.ListItemsResponse{
		Items:      protoItems,
		Total:      int32(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// SearchItems 搜索商品
func (s *ItemServer) SearchItems(ctx context.Context, req *itemsv1.SearchItemsRequest) (*itemsv1.SearchItemsResponse, error) {
	s.logger.Infof("SearchItems request received: query=%s", req.Query)

	// 构建请求 - 根据实际服务接口调整
	searchReq := &service.SearchItemsRequest{
		Query:  req.Query,
		Limit:  int(req.PageSize),
	}

	// 调用服务
	items, err := s.itemService.SearchItems(ctx, searchReq)
	if err != nil {
		s.logger.Errorf("Failed to search items: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to search items: %v", err)
	}

	// 转换响应
	protoItems := converters.BatchDomainItemsToProto(items)

	s.logger.Infof("Items searched successfully: returned=%d", len(protoItems))
	return &itemsv1.SearchItemsResponse{
		Items:    protoItems,
		Total:    int32(len(items)),
		Page:     req.Page,
		PageSize: req.PageSize,
		// 注意：由于服务不返回总数和总页数，这里使用返回的数量作为总数
		TotalPages: 1,
	}, nil
}

// GetItemStats 获取商品统计信息
func (s *ItemServer) GetItemStats(ctx context.Context, req *itemsv1.GetItemStatsRequest) (*itemsv1.GetItemStatsResponse, error) {
	s.logger.Infof("GetItemStats request received")

	// 调用服务 - 注意方法名是GetItemsStats
	stats, err := s.itemService.GetItemsStats(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get item stats: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get item stats: %v", err)
	}

	// 从分布数据中计算服务数和产品数
	totalServices := int32(0)
	totalProducts := int32(0)
	if stats.TypeDistribution != nil {
		if services, exists := stats.TypeDistribution[item.ItemTypeService]; exists {
			totalServices = int32(services)
		}
		if products, exists := stats.TypeDistribution[item.ItemTypeProduct]; exists {
			totalProducts = int32(products)
		}
	}

	// 构建响应
	response := &itemsv1.GetItemStatsResponse{
		TotalItems:     int32(stats.TotalItems),
		ActiveItems:    int32(stats.ActiveItems),
		InactiveItems:  int32(stats.InactiveItems),
		// 注意：服务统计数据结构中没有这个字段，设为0
		DraftItems:     0,
		TotalServices:  totalServices,
		TotalProducts:  totalProducts,
		// 注意：服务统计数据结构中没有这个字段，设为0
		TotalValue:     0,
		// 注意：protobuf中没有AveragePrice字段
	}

	s.logger.Infof("Item stats retrieved successfully")
	return response, nil
}

// UpdateItemStatus 更新商品状态
func (s *ItemServer) UpdateItemStatus(ctx context.Context, req *itemsv1.UpdateItemStatusRequest) (*itemsv1.UpdateItemStatusResponse, error) {
	s.logger.Infof("UpdateItemStatus request received: %s -> %v", req.Id, req.Status)

	// 解析ID
	itemID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid item ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid item ID: %v", err)
	}

	// 转换状态
	newStatus := converters.ItemStatusFromProto(req.Status)

	// 调用服务
	domainItem, err := s.itemService.UpdateItemStatus(ctx, itemID, newStatus)
	if err != nil {
		if err == item.ErrItemNotFound {
			s.logger.Warnf("Item not found for status update: %s", req.Id)
			return nil, status.Errorf(codes.NotFound, "item not found")
		}
		s.logger.Errorf("Failed to update item status: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update item status: %v", err)
	}

	// 转换响应
	protoItem := converters.DomainItemToProto(domainItem)

	s.logger.Infof("Item status updated successfully: %s", protoItem.Id)
	return &itemsv1.UpdateItemStatusResponse{
		Item: protoItem,
	}, nil
}

// ActivateItem 激活商品
func (s *ItemServer) ActivateItem(ctx context.Context, req *itemsv1.ActivateItemRequest) (*itemsv1.ActivateItemResponse, error) {
	s.logger.Infof("ActivateItem request received: %s", req.Id)

	// 解析ID
	itemID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid item ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid item ID: %v", err)
	}

	// 调用服务
	domainItem, err := s.itemService.ActivateItem(ctx, itemID)
	if err != nil {
		if err == item.ErrItemNotFound {
			s.logger.Warnf("Item not found for activation: %s", req.Id)
			return nil, status.Errorf(codes.NotFound, "item not found")
		}
		s.logger.Errorf("Failed to activate item: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to activate item: %v", err)
	}

	// 转换响应
	protoItem := converters.DomainItemToProto(domainItem)

	s.logger.Infof("Item activated successfully: %s", protoItem.Id)
	return &itemsv1.ActivateItemResponse{
		Item: protoItem,
	}, nil
}

// DeactivateItem 停用商品
func (s *ItemServer) DeactivateItem(ctx context.Context, req *itemsv1.DeactivateItemRequest) (*itemsv1.DeactivateItemResponse, error) {
	s.logger.Infof("DeactivateItem request received: %s", req.Id)

	// 解析ID
	itemID, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Errorf("Invalid item ID: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid item ID: %v", err)
	}

	// 调用服务
	domainItem, err := s.itemService.DeactivateItem(ctx, itemID)
	if err != nil {
		if err == item.ErrItemNotFound {
			s.logger.Warnf("Item not found for deactivation: %s", req.Id)
			return nil, status.Errorf(codes.NotFound, "item not found")
		}
		s.logger.Errorf("Failed to deactivate item: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to deactivate item: %v", err)
	}

	// 转换响应
	protoItem := converters.DomainItemToProto(domainItem)

	s.logger.Infof("Item deactivated successfully: %s", protoItem.Id)
	return &itemsv1.DeactivateItemResponse{
		Item: protoItem,
	}, nil
}

