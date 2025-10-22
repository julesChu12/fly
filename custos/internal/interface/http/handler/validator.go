package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/julesChu12/fly/custos/internal/application/dto"
)

// HandleValidationError 处理验证错误,返回详细的错误信息
func HandleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		fields := make(map[string]interface{})
		for _, fieldErr := range validationErrs {
			fields[fieldErr.Field()] = getValidationErrorMessage(fieldErr)
		}

		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Fields:  fields,
		})
		return
	}

	// 如果不是验证错误,返回通用的格式错误
	c.JSON(http.StatusBadRequest, &dto.ErrorResponse{
		Code:    "INVALID_REQUEST",
		Message: fmt.Sprintf("Invalid request format: %v", err),
	})
}

// getValidationErrorMessage 根据验证标签返回友好的错误信息
func getValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		if fieldErr.Type().Kind() == 24 { // string
			return fmt.Sprintf("Minimum length is %s characters", fieldErr.Param())
		}
		return fmt.Sprintf("Minimum value is %s", fieldErr.Param())
	case "max":
		if fieldErr.Type().Kind() == 24 { // string
			return fmt.Sprintf("Maximum length is %s characters", fieldErr.Param())
		}
		return fmt.Sprintf("Maximum value is %s", fieldErr.Param())
	case "eqfield":
		return fmt.Sprintf("Must match %s field", fieldErr.Param())
	case "url":
		return "Invalid URL format"
	case "uuid":
		return "Invalid UUID format"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fieldErr.Param())
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", fieldErr.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", fieldErr.Param())
	case "gt":
		return fmt.Sprintf("Must be greater than %s", fieldErr.Param())
	case "lt":
		return fmt.Sprintf("Must be less than %s", fieldErr.Param())
	case "alphanum":
		return "Only alphanumeric characters are allowed"
	case "numeric":
		return "Only numeric characters are allowed"
	case "alpha":
		return "Only alphabetic characters are allowed"
	default:
		return fmt.Sprintf("Validation failed on '%s' tag", fieldErr.Tag())
	}
}
