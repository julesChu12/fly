package converters

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	itemsv1 "github.com/julesChu12/fly/items/api/proto/items/v1"
	"github.com/julesChu12/fly/items/internal/domain/item"
	"github.com/julesChu12/fly/items/internal/application/service"
)

// DomainItemToProto 将 domain Item 转换为 protobuf Item
func DomainItemToProto(item *item.Item) *itemsv1.Item {
	if item == nil {
		return nil
	}

	proto := &itemsv1.Item{
		Id:          item.ID.String(),
		Name:        item.Name,
		Description: item.Description,
		Type:        ItemTypeToProto(item.Type),
		Price:       item.Price,
		CategoryId:  item.CategoryID.String(),
		Status:      ItemStatusToProto(item.Status),
		CreatedAt:   timestamppb.New(item.CreatedAt),
		UpdatedAt:   timestamppb.New(item.UpdatedAt),
	}

	// 可空字段处理
	if item.ImageURL != nil {
		proto.ImageUrl = *item.ImageURL
	}
	if item.Tags != nil {
		proto.Tags = *item.Tags
	}

	// 类型特有字段
	if item.IsService() {
		if item.Duration != nil {
			proto.Duration = wrapperspb.Int32(int32(*item.Duration))
		}
		if item.StaffRequired != nil {
			proto.StaffRequired = wrapperspb.Bool(*item.StaffRequired)
		}
		if item.Capacity != nil {
			proto.Capacity = wrapperspb.Int32(int32(*item.Capacity))
		}
	} else {
		if item.Stock != nil {
			proto.Stock = wrapperspb.Int32(int32(*item.Stock))
		}
		if item.CostPrice != nil {
			proto.CostPrice = wrapperspb.Double(*item.CostPrice)
		}
		if item.Weight != nil {
			proto.Weight = wrapperspb.Double(*item.Weight)
		}
		if item.SKU != nil {
			proto.Sku = wrapperspb.String(*item.SKU)
		}
	}

	return proto
}

// ProtoToDomainItem 将 protobuf Item 转换为 domain Item
func ProtoToDomainItem(proto *itemsv1.Item) (*item.Item, error) {
	if proto == nil {
		return nil, fmt.Errorf("protobuf item is nil")
	}

	// 解析 CategoryID
	categoryID, err := uuid.Parse(proto.CategoryId)
	if err != nil {
		return nil, fmt.Errorf("invalid category_id: %w", err)
	}

	domainItem := &item.Item{
		ID:          uuid.MustParse(proto.Id),
		Name:        proto.Name,
		Description: proto.Description,
		Type:        ItemTypeFromProto(proto.Type),
		Price:       proto.Price,
		CategoryID:  categoryID,
		Status:      ItemStatusFromProto(proto.Status),
		CreatedAt:   proto.CreatedAt.AsTime(),
		UpdatedAt:   proto.UpdatedAt.AsTime(),
	}

	// 可空字段处理
	if proto.ImageUrl != "" {
		imageURL := proto.ImageUrl
		domainItem.ImageURL = &imageURL
	}
	if proto.Tags != "" {
		tags := proto.Tags
		domainItem.Tags = &tags
	}

	// 类型特有字段
	if proto.Type == itemsv1.ItemType_ITEM_TYPE_SERVICE {
		if proto.Duration != nil {
			duration := int(proto.Duration.GetValue())
			domainItem.Duration = &duration
		}
		if proto.StaffRequired != nil {
			staffRequired := proto.StaffRequired.GetValue()
			domainItem.StaffRequired = &staffRequired
		}
		if proto.Capacity != nil {
			capacity := int(proto.Capacity.GetValue())
			domainItem.Capacity = &capacity
		}
	} else if proto.Type == itemsv1.ItemType_ITEM_TYPE_PRODUCT {
		if proto.Stock != nil {
			stock := int(proto.Stock.GetValue())
			domainItem.Stock = &stock
		}
		if proto.CostPrice != nil {
			costPrice := proto.CostPrice.GetValue()
			domainItem.CostPrice = &costPrice
		}
		if proto.Weight != nil {
			weight := proto.Weight.GetValue()
			domainItem.Weight = &weight
		}
		if proto.Sku != nil {
			sku := proto.Sku.GetValue()
			domainItem.SKU = &sku
		}
	}

	return domainItem, nil
}

// ProtoToCreateItemRequest 将 protobuf CreateItemRequest 转换为 domain CreateItemRequest
func ProtoToCreateItemRequest(proto *itemsv1.CreateItemRequest) (*service.CreateItemRequest, error) {
	if proto == nil {
		return nil, fmt.Errorf("protobuf CreateItemRequest is nil")
	}

	// 解析 CategoryID
	categoryID, err := uuid.Parse(proto.CategoryId)
	if err != nil {
		return nil, fmt.Errorf("invalid category_id: %w", err)
	}

	createReq := &service.CreateItemRequest{
		Name:        proto.Name,
		Description: proto.Description,
		Type:        ItemTypeFromProto(proto.Type),
		Price:       proto.Price,
		CategoryID:  categoryID,
	}
	// 可空字段
	if proto.ImageUrl != "" {
		imageURL := proto.ImageUrl
		createReq.ImageURL = &imageURL
	}
	if proto.Tags != "" {
		tags := proto.Tags
		createReq.Tags = &tags
	}

	// 类型特有字段
	if proto.Type == itemsv1.ItemType_ITEM_TYPE_SERVICE {
		if proto.Duration != nil {
			duration := int(proto.Duration.GetValue())
			createReq.Duration = &duration
		}
		if proto.StaffRequired != nil {
			staffRequired := proto.StaffRequired.GetValue()
			createReq.StaffRequired = &staffRequired
		}
		if proto.Capacity != nil {
			capacity := int(proto.Capacity.GetValue())
			createReq.Capacity = &capacity
		}
	} else if proto.Type == itemsv1.ItemType_ITEM_TYPE_PRODUCT {
		if proto.Stock != nil {
			stock := float64(proto.Stock.GetValue())
			createReq.Stock = &stock
		}
		if proto.CostPrice != nil {
			costPrice := proto.CostPrice.GetValue()
			createReq.CostPrice = &costPrice
		}
		if proto.Weight != nil {
			weight := proto.Weight.GetValue()
			createReq.Weight = &weight
		}
		if proto.Sku != nil {
			sku := proto.Sku.GetValue()
			createReq.SKU = &sku
		}
	}

	return createReq, nil
}

// ProtoToUpdateItemRequest 将 protobuf UpdateItemRequest 转换为 domain UpdateItemRequest
func ProtoToUpdateItemRequest(proto *itemsv1.UpdateItemRequest, id string) (*service.UpdateItemRequest, error) {
	if proto == nil {
		return nil, fmt.Errorf("protobuf UpdateItemRequest is nil")
	}

	updateReq := &service.UpdateItemRequest{}

	// 名称字段（必填）
	name := proto.Name
	updateReq.Name = &name

	// 价格字段（必填）
	price := proto.Price
	updateReq.Price = &price

	// 可空字段
	if proto.Description != "" {
		description := proto.Description
		updateReq.Description = &description
	}
	if proto.ImageUrl != "" {
		imageURL := proto.ImageUrl
		updateReq.ImageURL = &imageURL
	}
	if proto.Tags != "" {
		tags := proto.Tags
		updateReq.Tags = &tags
	}

	// 类型特有字段
	if proto.Duration != nil {
		duration := int(proto.Duration.GetValue())
		updateReq.Duration = &duration
	}
	if proto.StaffRequired != nil {
		staffRequired := proto.StaffRequired.GetValue()
		updateReq.StaffRequired = &staffRequired
	}
	if proto.Capacity != nil {
		capacity := int(proto.Capacity.GetValue())
		updateReq.Capacity = &capacity
	}
	if proto.Stock != nil {
		stock := float64(proto.Stock.GetValue())
		updateReq.Stock = &stock
	}
	if proto.CostPrice != nil {
		costPrice := proto.CostPrice.GetValue()
		updateReq.CostPrice = &costPrice
	}
	if proto.Weight != nil {
		weight := proto.Weight.GetValue()
		updateReq.Weight = &weight
	}
	if proto.Sku != nil {
		sku := proto.Sku.GetValue()
		updateReq.SKU = &sku
	}

	return updateReq, nil
}

// ItemTypeToProto 将 domain ItemType 转换为 protobuf ItemType
func ItemTypeToProto(itemType item.ItemType) itemsv1.ItemType {
	switch itemType {
	case item.ItemTypeService:
		return itemsv1.ItemType_ITEM_TYPE_SERVICE
	case item.ItemTypeProduct:
		return itemsv1.ItemType_ITEM_TYPE_PRODUCT
	default:
		return itemsv1.ItemType_ITEM_TYPE_UNSPECIFIED
	}
}

// ItemTypeFromProto 将 protobuf ItemType 转换为 domain ItemType
func ItemTypeFromProto(itemType itemsv1.ItemType) item.ItemType {
	switch itemType {
	case itemsv1.ItemType_ITEM_TYPE_SERVICE:
		return item.ItemTypeService
	case itemsv1.ItemType_ITEM_TYPE_PRODUCT:
		return item.ItemTypeProduct
	default:
		return item.ItemType("") // 空字符串表示未指定
	}
}

// ItemStatusToProto 将 domain ItemStatus 转换为 protobuf ItemStatus
func ItemStatusToProto(status item.ItemStatus) itemsv1.ItemStatus {
	switch status {
	case item.StatusDraft:
		return itemsv1.ItemStatus_ITEM_STATUS_DRAFT
	case item.StatusActive:
		return itemsv1.ItemStatus_ITEM_STATUS_ACTIVE
	case item.StatusInactive:
		return itemsv1.ItemStatus_ITEM_STATUS_INACTIVE
	case item.StatusArchived:
		return itemsv1.ItemStatus_ITEM_STATUS_ARCHIVED
	default:
		return itemsv1.ItemStatus_ITEM_STATUS_UNSPECIFIED
	}
}

// ItemStatusFromProto 将 protobuf ItemStatus 转换为 domain ItemStatus
func ItemStatusFromProto(status itemsv1.ItemStatus) item.ItemStatus {
	switch status {
	case itemsv1.ItemStatus_ITEM_STATUS_DRAFT:
		return item.StatusDraft
	case itemsv1.ItemStatus_ITEM_STATUS_ACTIVE:
		return item.StatusActive
	case itemsv1.ItemStatus_ITEM_STATUS_INACTIVE:
		return item.StatusInactive
	case itemsv1.ItemStatus_ITEM_STATUS_ARCHIVED:
		return item.StatusArchived
	default:
		return item.ItemStatus("") // 空字符串表示未指定
	}
}

// BatchDomainItemsToProto 批量转换 domain Items 到 protobuf
func BatchDomainItemsToProto(items []*item.Item) []*itemsv1.Item {
	if items == nil {
		return nil
	}

	protoItems := make([]*itemsv1.Item, len(items))
	for i, item := range items {
		protoItems[i] = DomainItemToProto(item)
	}
	return protoItems
}

// BatchProtoItemsToDomain 批量转换 protobuf Items 到 domain
func BatchProtoItemsToDomain(protoItems []*itemsv1.Item) ([]*item.Item, error) {
	if protoItems == nil {
		return nil, nil
	}

	items := make([]*item.Item, len(protoItems))
	for i, protoItem := range protoItems {
		item, err := ProtoToDomainItem(protoItem)
		if err != nil {
			return nil, fmt.Errorf("failed to convert proto item at index %d: %w", i, err)
		}
		items[i] = item
	}
	return items, nil
}