package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateCategory creates a new category
// @Summary Create category
// @Description Create a new category
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body object true "Category data"
// @Success 201 {object} object
// @Failure 400 {object} object
// @Router /categories [post]
func CreateCategory(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"code":    "CREATED",
		"message": "Category created successfully",
	})
}

// GetCategories retrieves a list of categories
// @Summary Get categories
// @Description Get a list of categories
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {object} object
// @Router /categories [get]
func GetCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Categories retrieved successfully",
		"data": gin.H{
			"categories": []interface{}{},
		},
	})
}

// GetCategoryTree retrieves category tree structure
// @Summary Get category tree
// @Description Get hierarchical category tree
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {object} object
// @Router /categories/tree [get]
func GetCategoryTree(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category tree retrieved successfully",
		"data": gin.H{
			"tree": []interface{}{},
		},
	})
}

// GetCategoryByID retrieves a category by ID
// @Summary Get category by ID
// @Description Get a category by its ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} object
// @Failure 404 {object} object
// @Router /categories/{id} [get]
func GetCategoryByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category retrieved successfully",
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateCategory updates an existing category
// @Summary Update category
// @Description Update an existing category
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body object true "Updated category data"
// @Success 200 {object} object
// @Failure 404 {object} object
// @Router /categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category updated successfully",
		"data": gin.H{
			"id": id,
		},
	})
}

// DeleteCategory deletes a category
// @Summary Delete category
// @Description Delete a category by its ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} object
// @Failure 404 {object} object
// @Router /categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category deleted successfully",
	})
}