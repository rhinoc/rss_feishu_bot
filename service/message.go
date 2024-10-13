package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rhinoc/rss_feishu_bot/util"
)

type FeishuSendMessageRequest struct {
	ReceiveIdType string `json:"receive_id_type"`
	ReceiveId     string `json:"receive_id"`
	MsgType       string `json:"msg_type"`
	Content       string `json:"content"`
}

func FeishuSendMessage(req FeishuSendMessageRequest) error {
	if req.ReceiveIdType == "" {
		req.ReceiveIdType = "open_id"
	}
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=%s", req.ReceiveIdType)

	// Convert the request to JSON
	jsonData, err := json.Marshal(req)
	fmt.Println("send message request", string(jsonData))
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	// Create a new HTTP POST request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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
	fmt.Println("send message response", string(body))

	return nil
}

func FeishuSendMessageText(receiveId, receiveIdType, content string) error {
	type TextContent struct {
		Text string `json:"text"`
	}

	textContent := TextContent{
		Text: content,
	}

	return FeishuSendMessage(FeishuSendMessageRequest{
		ReceiveId:     receiveId,
		ReceiveIdType: receiveIdType,
		MsgType:       "text",
		Content:       string(util.Must(json.Marshal(textContent))),
	})
}

type FeishuReceivedMessageRequest struct {
	Event struct {
		Message FeishuReceivedMessage `json:"message"`
		Sender  FeishuSender          `json:"sender"`
	} `json:"event"`
}

type FeishuReceivedMessage struct {
	ChatId     string `json:"chat_id"`
	ChatType   string `json:"chat_type"`
	Content    string `json:"content"`
	Text       string `json:"text"`
	CreateTime string `json:"create_time"`
}

type FeishuSender struct {
	SenderId struct {
		OpenId  string `json:"open_id"`
		UnionId string `json:"union_id"`
		UserId  string `json:"user_id"`
	} `json:"sender_id"`
	SenderType string `json:"sender_type"`
}

type FeishuReceivedMessageContent struct {
	Title   string `json:"title"`
	Text    string `json:"text"`
	Content [][]struct {
		Tag  string `json:"tag"`
		Text string `json:"text"`
	} `json:"content"`
}

func FeishuGetMessageReq(data []byte) (*FeishuReceivedMessageRequest, error) {
	var req FeishuReceivedMessageRequest
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling request: %w", err)
	}

	var content FeishuReceivedMessageContent
	err = json.Unmarshal([]byte(req.Event.Message.Content), &content)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling content: %w", err)
	}

	text := content.Text
	if text == "" && len(content.Content) > 0 {
		for _, row := range content.Content {
			for _, cell := range row {
				text += cell.Text
			}
		}
	}

	if len(text) == 0 {
		return nil, fmt.Errorf("empty message content")
	}

	req.Event.Message.Text = text

	return &req, nil
}
