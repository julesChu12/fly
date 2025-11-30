package converters

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	itemsv1 "github.com/julesChu12/fly/items/api/proto/items/v1"
	"github.com/julesChu12/fly/items/internal/domain/category"
)

// DomainCategoryToProto 将 domain Category 转换为 protobuf Category
func DomainCategoryToProto(cat *category.Category) *itemsv1.Category {
	if cat == nil {
		return nil
	}

	proto := &itemsv1.Category{
		Id:          cat.ID.String(),
		Name:        cat.Name,
		Description: cat.Description,
		SortOrder:   int32(cat.SortOrder),
		Status:      CategoryStatusToProto(cat.Status),
		ItemCount:   int32(cat.ItemCount),
		CreatedAt:   timestamppb.New(cat.CreatedAt),
		UpdatedAt:   timestamppb.New(cat.UpdatedAt),
	}

	// 处理图标字段（nullable）
	if cat.Icon != nil {
		proto.Icon = *cat.Icon
	}

	// 处理父级ID
	if cat.ParentID != nil {
		proto.ParentId = cat.ParentID.String()
	}

	// 处理删除时间
	if cat.DeletedAt != nil {
		proto.DeletedAt = timestamppb.New(*cat.DeletedAt)
	}

	return proto
}

// ProtoToDomainCategory 将 protobuf Category 转换为 domain Category
func ProtoToDomainCategory(proto *itemsv1.Category) (*category.Category, error) {
	if proto == nil {
		return nil, fmt.Errorf("protobuf category is nil")
	}

	domainCat := &category.Category{
		ID:          uuid.MustParse(proto.Id),
		Name:        proto.Name,
		Description: proto.Description,
		SortOrder:   int(proto.SortOrder),
		Status:      CategoryStatusFromProto(proto.Status),
		ItemCount:   int(proto.ItemCount),
		CreatedAt:   proto.CreatedAt.AsTime(),
		UpdatedAt:   proto.UpdatedAt.AsTime(),
	}

	// 处理图标字段（nullable）
	if proto.Icon != "" {
		icon := proto.Icon
		domainCat.Icon = &icon
	}

	// 处理父级ID
	if proto.ParentId != "" {
		parentID, err := uuid.Parse(proto.ParentId)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
		domainCat.ParentID = &parentID
	}

	// 处理删除时间
	if proto.DeletedAt != nil {
		deletedAt := proto.DeletedAt.AsTime()
		domainCat.DeletedAt = &deletedAt
	}

	return domainCat, nil
}

// DomainCategoryToProtoTree 将 domain Category 转换为 protobuf CategoryTree
func DomainCategoryToProtoTree(cat *category.Category, level int, path string, children []*category.Category) *itemsv1.CategoryTree {
	if cat == nil {
		return nil
	}

	proto := &itemsv1.CategoryTree{
		Id:          cat.ID.String(),
		Name:        cat.Name,
		Description: cat.Description,
		SortOrder:   int32(cat.SortOrder),
		Status:      CategoryStatusToProto(cat.Status),
		ItemCount:   int32(cat.ItemCount),
		Level:       int32(level),
		Path:        path,
		CreatedAt:   timestamppb.New(cat.CreatedAt),
		UpdatedAt:   timestamppb.New(cat.UpdatedAt),
	}

	// 处理图标字段（nullable）
	if cat.Icon != nil {
		proto.Icon = *cat.Icon
	}

	// 处理父级ID
	if cat.ParentID != nil {
		proto.ParentId = cat.ParentID.String()
	}

	// 转换子分类
	if len(children) > 0 {
		proto.Children = make([]*itemsv1.CategoryTree, len(children))
		childPath := path
		if childPath != "" {
			childPath += " > "
		}
		childPath += cat.Name

		for i, child := range children {
			// 递归转换子分类，假设这里我们只转换直接子分类
			// 在实际使用中，调用方需要提供完整的子分类树
			proto.Children[i] = DomainCategoryToProtoTree(child, level+1, childPath, nil)
		}
	}

	return proto
}

// ProtoToCreateCategoryRequest 将 protobuf CreateCategoryRequest 转换为 domain CreateCategoryRequest
func ProtoToCreateCategoryRequest(proto *itemsv1.CreateCategoryRequest) (*category.CreateCategoryRequest, error) {
	if proto == nil {
		return nil, fmt.Errorf("protobuf CreateCategoryRequest is nil")
	}

	createReq := &category.CreateCategoryRequest{
		Name:        proto.Name,
		Description: proto.Description,
		SortOrder:   int(proto.SortOrder),
	}

	// 图标字段（可选）
	if proto.Icon != "" {
		icon := proto.Icon
		createReq.Icon = &icon
	}

	// 父级ID
	if proto.ParentId != "" {
		parentID, err := uuid.Parse(proto.ParentId)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
		createReq.ParentID = &parentID
	}

	return createReq, nil
}

// ProtoToUpdateCategoryRequest 将 protobuf UpdateCategoryRequest 转换为 domain UpdateCategoryRequest
func ProtoToUpdateCategoryRequest(proto *itemsv1.UpdateCategoryRequest, id string) (*category.UpdateCategoryRequest, error) {
	if proto == nil {
		return nil, fmt.Errorf("protobuf UpdateCategoryRequest is nil")
	}

	updateReq := &category.UpdateCategoryRequest{}

	// 名称字段（必填）
	name := proto.Name
	updateReq.Name = &name

	// 描述字段（可选）
	description := proto.Description
	updateReq.Description = &description

	// 图标字段（可选）
	if proto.Icon != "" {
		icon := proto.Icon
		updateReq.Icon = &icon
	}

	// 父级ID
	if proto.ParentId != "" {
		parentID, err := uuid.Parse(proto.ParentId)
		if err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
		updateReq.ParentID = &parentID
	}

	// 排序字段
	if proto.SortOrder != 0 {
		sortOrder := int(proto.SortOrder)
		updateReq.SortOrder = &sortOrder
	}

	return updateReq, nil
}

// CategoryStatusToProto 将 domain CategoryStatus 转换为 protobuf CategoryStatus
func CategoryStatusToProto(status category.CategoryStatus) itemsv1.CategoryStatus {
	switch status {
	case category.CategoryStatusActive:
		return itemsv1.CategoryStatus_CATEGORY_STATUS_ACTIVE
	case category.CategoryStatusInactive:
		return itemsv1.CategoryStatus_CATEGORY_STATUS_INACTIVE
	default:
		return itemsv1.CategoryStatus_CATEGORY_STATUS_UNSPECIFIED
	}
}

// CategoryStatusFromProto 将 protobuf CategoryStatus 转换为 domain CategoryStatus
func CategoryStatusFromProto(status itemsv1.CategoryStatus) category.CategoryStatus {
	switch status {
	case itemsv1.CategoryStatus_CATEGORY_STATUS_ACTIVE:
		return category.CategoryStatusActive
	case itemsv1.CategoryStatus_CATEGORY_STATUS_INACTIVE:
		return category.CategoryStatusInactive
	default:
		return category.CategoryStatus("") // 空字符串表示未指定
	}
}

// BatchDomainCategoriesToProto 批量转换 domain Categories 到 protobuf
func BatchDomainCategoriesToProto(categories []*category.Category) []*itemsv1.Category {
	if categories == nil {
		return nil
	}

	protoCategories := make([]*itemsv1.Category, len(categories))
	for i, cat := range categories {
		protoCategories[i] = DomainCategoryToProto(cat)
	}
	return protoCategories
}

// BatchProtoCategoriesToDomain 批量转换 protobuf Categories 到 domain
func BatchProtoCategoriesToDomain(protoCategories []*itemsv1.Category) ([]*category.Category, error) {
	if protoCategories == nil {
		return nil, nil
	}

	categories := make([]*category.Category, len(protoCategories))
	for i, protoCat := range protoCategories {
		cat, err := ProtoToDomainCategory(protoCat)
		if err != nil {
			return nil, fmt.Errorf("failed to convert proto category at index %d: %w", i, err)
		}
		categories[i] = cat
	}
	return categories, nil
}

// BatchDomainCategoriesToProtoTree 批量转换 domain Categories 到 protobuf CategoryTree
func BatchDomainCategoriesToProtoTree(categories []*category.Category) []*itemsv1.CategoryTree {
	if categories == nil {
		return nil
	}

	protoCategories := make([]*itemsv1.CategoryTree, len(categories))
	for i, cat := range categories {
		// 假设这是顶级分类，level=0，path为分类名称
		protoCategories[i] = DomainCategoryToProtoTree(cat, 0, cat.Name, nil)
	}
	return protoCategories
}

