package item

import "errors"

// 定义错误
var (
	ErrItemNotFound        = errors.New("item not found")
	ErrInvalidItemType     = errors.New("invalid item type")
	ErrInvalidItemStatus   = errors.New("invalid item status")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrInvalidDuration     = errors.New("invalid duration")
	ErrInvalidStock        = errors.New("invalid stock")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrNotProductOrNoStock = errors.New("not a product or no stock field")
	ErrDuplicateSKU        = errors.New("duplicate SKU")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCannotDeleteActive  = errors.New("cannot delete active item")
)