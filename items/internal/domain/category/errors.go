package category

import "errors"

// 定义错误
var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrDuplicateCategory   = errors.New("duplicate category name")
	ErrHasChildren         = errors.New("category has children")
	ErrHasItems            = errors.New("category has items")
	ErrCircularReference   = errors.New("circular reference detected")
	ErrCannotDeleteRoot    = errors.New("cannot delete root category")
	ErrInvalidParent       = errors.New("invalid parent category")
	ErrMaxDepthExceeded    = errors.New("maximum category depth exceeded")
	ErrCategoryInUse       = errors.New("category is in use")
)