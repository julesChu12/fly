package category

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewCategory(t *testing.T) {
	name := "美容美发"
	description := "美容和美发相关服务"
	parentID := uuid.New()

	category := NewCategory(name, description, &parentID)

	assert.NotEqual(t, uuid.Nil, category.ID)
	assert.Equal(t, name, category.Name)
	assert.Equal(t, description, category.Description)
	assert.Equal(t, &parentID, category.ParentID)
	assert.Equal(t, CategoryStatusActive, category.Status)
	assert.Equal(t, 0, category.SortOrder)
	assert.False(t, category.CreatedAt.IsZero())
	assert.False(t, category.UpdatedAt.IsZero())
	assert.Nil(t, category.DeletedAt)
}

func TestNewCategory_WithoutParent(t *testing.T) {
	name := "根分类"
	description := "根分类描述"

	category := NewCategory(name, description, nil)

	assert.NotEqual(t, uuid.Nil, category.ID)
	assert.Equal(t, name, category.Name)
	assert.Equal(t, description, category.Description)
	assert.Nil(t, category.ParentID)
	assert.Equal(t, CategoryStatusActive, category.Status)
	assert.Equal(t, 0, category.SortOrder)
}

func TestCategory_IsRoot(t *testing.T) {
	t.Run("root category (no parent)", func(t *testing.T) {
		category := &Category{ParentID: nil}
		assert.True(t, category.IsRoot())
	})

	t.Run("child category (has parent)", func(t *testing.T) {
		parentID := uuid.New()
		category := &Category{ParentID: &parentID}
		assert.False(t, category.IsRoot())
	})
}

func TestCategory_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status CategoryStatus
		want   bool
	}{
		{
			name:   "active status",
			status: CategoryStatusActive,
			want:   true,
		},
		{
			name:   "inactive status",
			status: CategoryStatusInactive,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := &Category{Status: tt.status}
			assert.Equal(t, tt.want, category.IsActive())
		})
	}
}

func TestCategory_GetPath(t *testing.T) {
	t.Run("root category path", func(t *testing.T) {
		category := &Category{Name: "根分类", ParentID: nil}
		assert.Equal(t, "根分类", category.GetPath())
	})

	t.Run("child category path with parent", func(t *testing.T) {
		parentID := uuid.New()
		parent := &Category{
			ID:       parentID,
			Name:     "父分类",
			ParentID: nil,
		}
		child := &Category{
			ID:       uuid.New(),
			Name:     "子分类",
			ParentID: &parentID,
			Parent:   parent,
		}

		assert.Equal(t, "父分类 > 子分类", child.GetPath())
	})

	t.Run("child category path without parent reference", func(t *testing.T) {
		child := &Category{
			Name:     "子分类",
			ParentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			Parent:   nil,
		}

		assert.Equal(t, "子分类", child.GetPath())
	})

	t.Run("nested category path", func(t *testing.T) {
		grandparentID := uuid.New()
		parentID := uuid.New()
		grandparent := &Category{
			ID:       grandparentID,
			Name:     "根分类",
			ParentID: nil,
		}
		parent := &Category{
			ID:       parentID,
			Name:     "父分类",
			ParentID: &grandparentID,
			Parent:   grandparent,
		}
		child := &Category{
			ID:       uuid.New(),
			Name:     "子分类",
			ParentID: &parentID,
			Parent:   parent,
		}

		assert.Equal(t, "根分类 > 父分类 > 子分类", child.GetPath())
	})
}

func TestCategory_GetLevel(t *testing.T) {
	t.Run("root category level", func(t *testing.T) {
		category := &Category{ParentID: nil}
		assert.Equal(t, 0, category.GetLevel())
	})

	t.Run("child category level", func(t *testing.T) {
		parentID := uuid.New()
		parent := &Category{
			ID:       parentID,
			Name:     "父分类",
			ParentID: nil,
		}
		child := &Category{
			ID:       uuid.New(),
			Name:     "子分类",
			ParentID: &parentID,
			Parent:   parent,
		}

		assert.Equal(t, 1, child.GetLevel())
	})

	t.Run("nested category level", func(t *testing.T) {
		grandparentID := uuid.New()
		parentID := uuid.New()
		grandparent := &Category{
			ID:       grandparentID,
			Name:     "根分类",
			ParentID: nil,
		}
		parent := &Category{
			ID:       parentID,
			Name:     "父分类",
			ParentID: &grandparentID,
			Parent:   grandparent,
		}
		child := &Category{
			ID:       uuid.New(),
			Name:     "子分类",
			ParentID: &parentID,
			Parent:   parent,
		}

		assert.Equal(t, 2, child.GetLevel())
	})

	t.Run("child category level without parent reference", func(t *testing.T) {
		child := &Category{
			Name:     "子分类",
			ParentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			Parent:   nil,
		}

		assert.Equal(t, 0, child.GetLevel())
	})
}

func TestCategory_CanDelete(t *testing.T) {
	t.Run("can delete empty category", func(t *testing.T) {
		category := &Category{
			ItemCount: 0,
			Children:  []*Category{},
		}
		assert.True(t, category.CanDelete())
	})

	t.Run("cannot delete category with items", func(t *testing.T) {
		category := &Category{
			ItemCount: 5,
			Children:  []*Category{},
		}
		assert.False(t, category.CanDelete())
	})

	t.Run("cannot delete category with children", func(t *testing.T) {
		child := &Category{}
		category := &Category{
			ItemCount: 0,
			Children:  []*Category{child},
		}
		assert.False(t, category.CanDelete())
	})

	t.Run("cannot delete category with items and children", func(t *testing.T) {
		child := &Category{}
		category := &Category{
			ItemCount: 3,
			Children:  []*Category{child},
		}
		assert.False(t, category.CanDelete())
	})

	t.Run("nil children slice", func(t *testing.T) {
		category := &Category{
			ItemCount: 0,
			Children:  nil,
		}
		assert.True(t, category.CanDelete())
	})
}

func TestCategory_Update(t *testing.T) {
	category := NewCategory("测试分类", "测试描述", nil)
	originalUpdatedAt := category.UpdatedAt

	// 等待一毫秒确保时间戳不同
	time.Sleep(1 * time.Millisecond)

	newName := "更新分类"
	newDescription := "更新描述"
	icon := "icon.png"
	sortOrder := 10

	category.Update(newName, newDescription, &icon, sortOrder)

	assert.Equal(t, newName, category.Name)
	assert.Equal(t, newDescription, category.Description)
	assert.Equal(t, &icon, category.Icon)
	assert.Equal(t, sortOrder, category.SortOrder)
	assert.True(t, category.UpdatedAt.After(originalUpdatedAt))
}

func TestCategory_Activate(t *testing.T) {
	category := NewCategory("测试分类", "测试描述", nil)
	category.Deactivate() // 先设为非激活状态
	originalUpdatedAt := category.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	category.Activate()

	assert.Equal(t, CategoryStatusActive, category.Status)
	assert.True(t, category.UpdatedAt.After(originalUpdatedAt))
}

func TestCategory_Deactivate(t *testing.T) {
	category := NewCategory("测试分类", "测试描述", nil)
	originalUpdatedAt := category.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	category.Deactivate()

	assert.Equal(t, CategoryStatusInactive, category.Status)
	assert.True(t, category.UpdatedAt.After(originalUpdatedAt))
}

func TestCategory_UpdateItemCount(t *testing.T) {
	category := NewCategory("测试分类", "测试描述", nil)
	originalUpdatedAt := category.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	category.UpdateItemCount(15)

	assert.Equal(t, 15, category.ItemCount)
	assert.True(t, category.UpdatedAt.After(originalUpdatedAt))
}

func TestCategory_AddChild(t *testing.T) {
	parent := NewCategory("父分类", "父分类描述", nil)
	child := NewCategory("子分类", "子分类描述", nil)

	parent.AddChild(child)

	assert.Len(t, parent.Children, 1)
	assert.Equal(t, child, parent.Children[0])
	assert.Equal(t, &parent.ID, child.ParentID)
	assert.Equal(t, parent, child.Parent)
}

func TestCategory_AddChild_ToExistingChildren(t *testing.T) {
	parent := NewCategory("父分类", "父分类描述", nil)
	child1 := NewCategory("子分类1", "子分类描述1", nil)
	child2 := NewCategory("子分类2", "子分类描述2", nil)

	// 第一次添加
	parent.AddChild(child1)
	assert.Len(t, parent.Children, 1)

	// 第二次添加
	parent.AddChild(child2)
	assert.Len(t, parent.Children, 2)
	assert.Contains(t, parent.Children, child1)
	assert.Contains(t, parent.Children, child2)
}

func TestCategory_ToTree(t *testing.T) {
	// 创建父子分类结构
	parent := NewCategory("父分类", "父分类描述", nil)
	child1 := NewCategory("子分类1", "子分类描述1", nil)
	child2 := NewCategory("子分类2", "子分类描述2", nil)

	parent.AddChild(child1)
	parent.AddChild(child2)

	// 转换为树结构
	tree := parent.ToTree()

	// 验证根节点
	assert.Equal(t, parent.ID, tree.ID)
	assert.Equal(t, parent.Name, tree.Name)
	assert.Equal(t, parent.Description, tree.Description)
	assert.Equal(t, parent.ParentID, tree.ParentID)
	assert.Equal(t, parent.Icon, tree.Icon)
	assert.Equal(t, parent.SortOrder, tree.SortOrder)
	assert.Equal(t, parent.Status, tree.Status)
	assert.Equal(t, parent.ItemCount, tree.ItemCount)
	assert.Equal(t, parent.CreatedAt, tree.CreatedAt)
	assert.Equal(t, parent.UpdatedAt, tree.UpdatedAt)

	// 验证子节点
	assert.Len(t, tree.Children, 2)

	// 验证子节点数据
	child1Tree := tree.Children[0]
	assert.Equal(t, child1.ID, child1Tree.ID)
	assert.Equal(t, child1.Name, child1Tree.Name)
	assert.Equal(t, &parent.ID, child1Tree.ParentID)

	child2Tree := tree.Children[1]
	assert.Equal(t, child2.ID, child2Tree.ID)
	assert.Equal(t, child2.Name, child2Tree.Name)
	assert.Equal(t, &parent.ID, child2Tree.ParentID)
}

func TestCategory_ToTree_WithoutChildren(t *testing.T) {
	category := NewCategory("叶子分类", "叶子分类描述", nil)

	tree := category.ToTree()

	assert.Equal(t, category.ID, tree.ID)
	assert.Equal(t, category.Name, tree.Name)
	assert.Len(t, tree.Children, 0)
}

func TestCategory_TableName(t *testing.T) {
	category := &Category{}
	assert.Equal(t, "categories", category.TableName())
}

func TestCategory_Constants(t *testing.T) {
	assert.Equal(t, CategoryStatus("ACTIVE"), CategoryStatusActive)
	assert.Equal(t, CategoryStatus("INACTIVE"), CategoryStatusInactive)
}

func TestCategoryTree_Structure(t *testing.T) {
	// 验证CategoryTree结构体的字段
	tree := &CategoryTree{
		ID:          uuid.New(),
		Name:        "测试分类",
		Description: "测试描述",
		Level:       1,
		Path:        "父分类 > 测试分类",
		Children:    []*CategoryTree{},
	}

	assert.NotEqual(t, uuid.Nil, tree.ID)
	assert.Equal(t, "测试分类", tree.Name)
	assert.Equal(t, "测试描述", tree.Description)
	assert.Equal(t, 1, tree.Level)
	assert.Equal(t, "父分类 > 测试分类", tree.Path)
	assert.NotNil(t, tree.Children)
}

func TestCategoryPath_Structure(t *testing.T) {
	categoryID := uuid.New()
	path := &CategoryPath{
		CategoryID: categoryID,
		Name:       "测试分类",
		Level:      1,
		Path:       "父分类 > 测试分类",
	}

	assert.Equal(t, categoryID, path.CategoryID)
	assert.Equal(t, "测试分类", path.Name)
	assert.Equal(t, 1, path.Level)
	assert.Equal(t, "父分类 > 测试分类", path.Path)
}