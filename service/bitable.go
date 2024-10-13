package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type BitableRecordCondition struct {
	FieldName string   `json:"field_name"`
	Operator  string   `json:"operator"`
	Value     []string `json:"value"`
}

type BitableRecordFilterInfo struct {
	Conjunction string                   `json:"conjunction"`
	Conditions  []BitableRecordCondition `json:"conditions"`
}

type FeishuGetBitableRecordRequest struct {
	ViewId string                  `json:"view_id"`
	Filter BitableRecordFilterInfo `json:"filter"`
}

type FeishuGetBitableRecordItem struct {
	RecordID string                 `json:"record_id"`
	Fields   map[string]interface{} `json:"fields"`
}

type FeishuGetBitableRecordData struct {
	Total     int                          `json:"total"`
	HasMore   bool                         `json:"has_more"`
	PageToken string                       `json:"page_token"`
	Items     []FeishuGetBitableRecordItem `json:"items"`
}

type FeishuGetBitableRecordResponse struct {
	Code int                        `json:"code"`
	Msg  string                     `json:"msg"`
	Data FeishuGetBitableRecordData `json:"data"`
}

// https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/bitable-v1/app-table-record/search
func FeishuGetBitableRecord(appToken string, tableId string, req FeishuGetBitableRecordRequest) (*FeishuGetBitableRecordResponse, error) {
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records/search", appToken, tableId)

	// Convert the request to JSON
	jsonData, err := json.Marshal(req)
	fmt.Println("bitable get record request", string(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create a new HTTP POST request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	accessToken, err := GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("error getting access token: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// print response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	fmt.Println("bitable get record response", string(body))

	// Parse the response
	var response FeishuGetBitableRecordResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

type FeishuUpdateBitableRecordRequest struct {
	Fields map[string]interface{} `json:"fields"`
}

func FeishuUpdateBitableRecord(appToken string, tableId string, recordId string, req FeishuUpdateBitableRecordRequest) error {
	log.Println("bitable update record", appToken, tableId, recordId, req)

	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records/%s", appToken, tableId, recordId)

	// Convert the request to JSON
	jsonData, err := json.Marshal(req)
	fmt.Println("bitable get record request", string(jsonData))
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	// Create a new HTTP PUT request
	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	accessToken, err := GetAccessToken()
	if err != nil {
		return fmt.Errorf("error getting access token: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// print response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	fmt.Println("bitable update record response", string(body))

	return nil
}

type FeishuAddBitableRecordRequest struct {
	Fields map[string]interface{} `json:"fields"`
}

type FeishuAddBitableRecordResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Record struct {
			RecordID string `json:"record_id"`
		} `json:"record"`
	} `json:"data"`
}

func FeishuAddBitableRecord(appToken string, tableId string, req FeishuAddBitableRecordRequest) (string, error) {
	log.Println("bitable add record", appToken, tableId, req)
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records", appToken, tableId)

	// Convert the request to JSON
	jsonData, err := json.Marshal(req)
	fmt.Println("bitable add record request", string(jsonData))
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	// Create a new HTTP POST request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	accessToken, err := GetAccessToken()
	if err != nil {
		return "", fmt.Errorf("error getting access token: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// print response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	fmt.Println("bitable update record response", string(body))

	// get record id from response
	var response FeishuAddBitableRecordResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	if response.Data.Record.RecordID == "" {
		return "", fmt.Errorf("no record id found in response")
	}

	return response.Data.Record.RecordID, nil
}
