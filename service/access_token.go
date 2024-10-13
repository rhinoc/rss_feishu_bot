package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rhinoc/rss_feishu_bot/config"
)

// Add these variables at the package level
var (
	cachedToken string
	tokenExpiry time.Time
)

func GetAccessToken() (string, error) {
	url := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"

	// Check if we have a cached token that's still valid
	if cachedToken != "" && time.Now().Before(tokenExpiry) {
		return cachedToken, nil
	}

	// Prepare the request body
	requestBody := map[string]string{
		"app_id":     config.AppID,
		"app_secret": config.AppSecret,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Make the HTTP POST request
	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}

	fmt.Println("access token response", string(body))
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		return "", fmt.Errorf("API error: %s", response.Msg)
	}

	// Cache the token and set expiry time
	cachedToken = response.TenantAccessToken
	tokenExpiry = time.Now().Add(time.Duration(response.Expire) * time.Second)

	return response.TenantAccessToken, nil
}
