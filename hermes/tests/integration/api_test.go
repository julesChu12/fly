//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// APITestSuite for integration tests
type APITestSuite struct {
	suite.Suite
	baseURL    string
	httpClient *http.Client
}

// SetupSuite runs once before all tests
func (suite *APITestSuite) SetupSuite() {
	suite.baseURL = "http://localhost:8080"
	suite.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Wait for service to be ready
	suite.waitForService()
}

// TearDownSuite runs once after all tests
func (suite *APITestSuite) TearDownSuite() {
	// Cleanup if needed
}

// waitForService waits for the service to be ready
func (suite *APITestSuite) waitForService() {
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := suite.httpClient.Get(fmt.Sprintf("%s/health", suite.baseURL))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			suite.T().Log("Service is ready")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(retryDelay)
	}

	suite.T().Fatal("Service did not become ready within expected time")
}

// TestHealthCheck tests the health endpoint
func (suite *APITestSuite) TestHealthCheck() {
	resp, err := suite.httpClient.Get(fmt.Sprintf("%s/health", suite.baseURL))

	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	suite.NoError(err)
	suite.Equal("ok", result["status"])
	suite.Equal("hermes", result["service"])
}

// TestCreateCustomer tests creating a customer via REST API
func (suite *APITestSuite) TestCreateCustomer() {
	payload := map[string]interface{}{
		"name":  "Integration Test Customer",
		"phone": "+1234567890",
		"email": "integration@example.com",
		"tags":  "integration,test",
	}

	jsonData, err := json.Marshal(payload)
	suite.NoError(err)

	resp, err := suite.httpClient.Post(
		fmt.Sprintf("%s/api/customers", suite.baseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	suite.NoError(err)

	data, ok := result["data"].(map[string]interface{})
	suite.True(ok)
	suite.Equal("Integration Test Customer", data["name"])
	suite.Equal("integration@example.com", data["email"])
}

// TestGetCustomer tests retrieving a customer
func (suite *APITestSuite) TestGetCustomer() {
	// First create a customer
	createPayload := map[string]interface{}{
		"name":  "Get Test Customer",
		"phone": "+0987654321",
		"email": "gettest@example.com",
		"tags":  "get,test",
	}

	jsonData, _ := json.Marshal(createPayload)
	resp, err := suite.httpClient.Post(
		fmt.Sprintf("%s/api/customers", suite.baseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	suite.NoError(err)
	defer resp.Body.Close()
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var createResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResult)
	suite.NoError(err)

	customerID := int(createResult["data"].(map[string]interface{})["id"].(float64))

	// Now get the customer
	resp, err = suite.httpClient.Get(
		fmt.Sprintf("%s/api/customers/%d", suite.baseURL, customerID),
	)

	suite.NoError(err)
	defer resp.Body.Close()
	suite.Equal(http.StatusOK, resp.StatusCode)

	var getResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&getResult)
	suite.NoError(err)

	data := getResult["data"].(map[string]interface{})
	suite.Equal("Get Test Customer", data["name"])
	suite.Equal("gettest@example.com", data["email"])
}

// TestListCustomers tests listing customers
func (suite *APITestSuite) TestListCustomers() {
	resp, err := suite.httpClient.Get(
		fmt.Sprintf("%s/api/customers?page=1&page_size=10", suite.baseURL),
	)

	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	suite.NoError(err)

	suite.Contains(result, "data")
	suite.Contains(result, "total")
	suite.Contains(result, "page")
	suite.Contains(result, "size")
}

// TestSwaggerDocumentation tests Swagger documentation endpoint
func (suite *APITestSuite) TestSwaggerDocumentation() {
	resp, err := suite.httpClient.Get(fmt.Sprintf("%s/swagger/index.html", suite.baseURL))

	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	// Verify it's HTML content
	contentType := resp.Header.Get("Content-Type")
	suite.Contains(contentType, "text/html")
}

// TestErrorHandling tests API error handling
func (suite *APITestSuite) TestErrorHandling() {
	// Test getting non-existent customer
	resp, err := suite.httpClient.Get(fmt.Sprintf("%s/api/customers/99999", suite.baseURL))

	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode)

	var errorResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResult)
	suite.NoError(err)

	suite.Contains(errorResult, "error")
}

// TestCustomerWorkflow tests complete customer workflow
func (suite *APITestSuite) TestCustomerWorkflow() {
	// 1. Create customer
	customer := map[string]interface{}{
		"name":  "Workflow Test Customer",
		"phone": "+1122334455",
		"email": "workflow@example.com",
		"tags":  "workflow,test",
	}

	jsonData, _ := json.Marshal(customer)
	resp, err := suite.httpClient.Post(
		fmt.Sprintf("%s/api/customers", suite.baseURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	suite.NoError(err)
	defer resp.Body.Close()
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var createResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResult)
	suite.NoError(err)

	customerData := createResult["data"].(map[string]interface{})
	customerID := int(customerData["id"].(float64))

	// 2. Get customer
	resp, err = suite.httpClient.Get(
		fmt.Sprintf("%s/api/customers/%d", suite.baseURL, customerID),
	)

	suite.NoError(err)
	defer resp.Body.Close()
	suite.Equal(http.StatusOK, resp.StatusCode)

	// 3. Update customer
	updatePayload := map[string]interface{}{
		"name":  "Updated Workflow Customer",
		"phone": "+9988776655",
		"email": "updated@example.com",
		"tags":  "updated,workflow",
	}

	updateJsonData, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/api/customers/%d", suite.baseURL, customerID),
		bytes.NewBuffer(updateJsonData),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err = suite.httpClient.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()
	suite.Equal(http.StatusOK, resp.StatusCode)

	// 4. Verify update
	var updateResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updateResult)
	suite.NoError(err)

	updatedData := updateResult["data"].(map[string]interface{})
	suite.Equal("Updated Workflow Customer", updatedData["name"])
	suite.Equal("updated@example.com", updatedData["email"])

	// 5. Delete customer
	req, _ = http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/api/customers/%d", suite.baseURL, customerID),
		nil,
	)

	resp, err = suite.httpClient.Do(req)
	suite.NoError(err)
	defer resp.Body.Close()
	suite.Equal(http.StatusOK, resp.StatusCode)

	// 6. Verify deletion
	resp, err = suite.httpClient.Get(
		fmt.Sprintf("%s/api/customers/%d", suite.baseURL, customerID),
	)

	suite.NoError(err)
	defer resp.Body.Close()
	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

// TestAPITestSuite runs the test suite
func TestAPITestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(APITestSuite))
}