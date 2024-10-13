package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/rhinoc/rss_feishu_bot/config"
	"github.com/rhinoc/rss_feishu_bot/model"
	"github.com/rhinoc/rss_feishu_bot/service"
	"github.com/rhinoc/rss_feishu_bot/util"
)

type FeishuCallbackRequest struct {
	Challenge string      `json:"challenge"`
	Token     string      `json:"token"`
	Type      string      `json:"type"`
	Schema    string      `json:"schema"`
	Header    interface{} `json:"header"`
	Event     interface{} `json:"event"`
}

func (req *FeishuCallbackRequest) Bind(r *http.Request) error {
	return nil
}

func FeishuCallbackHandler(w http.ResponseWriter, r *http.Request) {
	data := &FeishuCallbackRequest{}

	if err := render.Bind(r, data); err != nil {
		log.Println("error binding request", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, data)
		return
	}

	// print req json
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("error marshalling request", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, data)
		return
	}
	log.Println("request", string(jsonData))

	req, err := service.FeishuGetMessageReq(jsonData)
	if err == nil {
		go handleMessage(req)
	}

	render.JSON(w, r, data)
}

func handleMessage(req *service.FeishuReceivedMessageRequest) {
	// parse message
	message := req.Event.Message
	text := message.Text
	log.Println("handle message text:", text)

	// parse target
	targetOpenId := req.Event.Sender.SenderId.OpenId
	isGroup := message.ChatType == "group" && strings.Contains(text, " -g")
	if isGroup {
		targetOpenId = req.Event.Message.ChatId
	}

	// handle command
	if strings.Contains(text, "/list") {
		// command: /list [-g]
		handleList(targetOpenId, isGroup, req.Event.Message.ChatId)
	} else if strings.Contains(text, "/add ") {
		// command: /add [-g] <url>
		handleAdd(text, targetOpenId, isGroup, req.Event.Message.ChatId)
	} else if strings.Contains(text, "/remove ") {
		// command: /remove [-g] <url>
		handleRemove(text, targetOpenId, isGroup, req.Event.Message.ChatId)
	} else if strings.Contains(text, "/send") {
		// command: /send [-g]
		handleSend(targetOpenId, isGroup, req.Event.Message.ChatId)
	} else if strings.Contains(text, "/help") {
		// command: /help
		handleHelp(req.Event.Message.ChatId)
	}
}

func handleList(targetOpenId string, isGroup bool, chatId string) {
	recordItem, err := service.GetRecordItem(targetOpenId, isGroup)
	if err != nil || len(recordItem.FeedList) == 0 {
		log.Println("error getting record item", err)
		service.FeishuSendMessageText(chatId, "chat_id", "No subscribed feeds found")
		return
	}

	items := []model.FeishuMessageItem{}
	for _, feed := range recordItem.FeedList {
		items = append(items, model.FeishuMessageItem{
			Title: feed.Link,
			Link:  feed.Link,
		})
	}

	content := &model.FeishuMessageContent{
		Type: "template",
		Data: model.FeishuMessageData{
			TemplateId:          config.CardTemplateId,
			TemplateVersionName: config.CardTemplateVersionName,
			TemplateVariable: map[string]interface{}{
				"cardTitle": "Subscribed Feed List",
				"cardColor": "purple",
				"itemList":  items,
			},
		},
	}

	service.FeishuSendMessage(service.FeishuSendMessageRequest{
		ReceiveId:     chatId,
		ReceiveIdType: "chat_id",
		MsgType:       "interactive",
		Content:       string(util.Must(json.Marshal(content))),
	})
}

func handleAdd(text string, targetOpenId string, isGroup bool, chatId string) {
	url := util.ExtractUrl(text)
	if url == "" {
		log.Println("no url found")
		service.FeishuSendMessageText(chatId, "chat_id", "Please provide a valid URL")
		return
	}

	recordItem, err := service.GetRecordItem(targetOpenId, isGroup)
	if err != nil {
		// create new record item
		userOpenId := targetOpenId
		if isGroup {
			userOpenId = ""
		}

		groupOpenId := ""
		if isGroup {
			groupOpenId = targetOpenId
		}

		recordItem = &service.RecordItem{
			UserOpenId:  userOpenId,
			GroupOpenId: groupOpenId,
			FeedList: []*service.RecordItemFeed{
				{
					Link: url,
				},
			},
		}

		_, err = service.AddRecordItem(*recordItem)
		if err != nil {
			log.Println("error adding record item", err)
			service.FeishuSendMessageText(chatId, "chat_id", fmt.Sprintf("Failed to add subscription: %s", url))
			return
		}

		service.FeishuSendMessageText(chatId, "chat_id", fmt.Sprintf("Successfully added subscription: %s", url))
		return
	}

	// check if the url is already in the feed list
	for _, feed := range recordItem.FeedList {
		if feed.Link == url {
			service.FeishuSendMessageText(chatId, "chat_id", fmt.Sprintf("This URL has already been subscribed: %s", url))
			return
		}
	}

	// add the url to the feed list
	recordItem.FeedList = append(recordItem.FeedList, &service.RecordItemFeed{
		Link: url,
	})

	err = service.UpdateRecordItemFeedList(*recordItem)
	if err != nil {
		log.Println("error updating record item", err)
		service.FeishuSendMessageText(chatId, "chat_id", fmt.Sprintf("Failed to add subscription: %s", url))
		return
	}

	service.FeishuSendMessageText(chatId, "chat_id", fmt.Sprintf("Successfully added subscription: %s", url))
}

func handleRemove(text string, targetOpenId string, isGroup bool, chatId string) {
	url := util.ExtractUrl(text)
	if url == "" {
		log.Println("no url found")
		service.FeishuSendMessageText(chatId, "chat_id", "Please provide a valid URL")
		return
	}
	recordItem, err := service.GetRecordItem(targetOpenId, isGroup)
	if err != nil {
		log.Println("error getting record item", err)
		service.FeishuSendMessageText(chatId, "chat_id", "No subscribed feeds found")
		return
	}

	// check if the url is in the feed list
	for i, feed := range recordItem.FeedList {
		if feed.Link == url {
			recordItem.FeedList = append(recordItem.FeedList[:i], recordItem.FeedList[i+1:]...)
			err = service.UpdateRecordItemFeedList(*recordItem)
			if err != nil {
				log.Println("error updating record item", err)
				service.FeishuSendMessageText(chatId, "chat_id", fmt.Sprintf("Failed to remove subscription: %s", url))
				return
			}

			service.FeishuSendMessageText(chatId, "chat_id", fmt.Sprintf("Successfully removed subscription: %s", url))
			return
		}
	}

	service.FeishuSendMessageText(chatId, "chat_id", fmt.Sprintf("This URL has not been subscribed yet: %s", url))
}

func handleSend(targetOpenId string, isGroup bool, chatId string) {
	recordItem, err := service.GetRecordItem(targetOpenId, isGroup)
	if err != nil || len(recordItem.FeedList) == 0 {
		log.Println("error getting record item", err)
		service.FeishuSendMessageText(chatId, "chat_id", "No subscribed feeds found")
		return
	}

	err = service.SendRssMessageByRecord(*recordItem)
	if err != nil {
		log.Println("error sending rss message", err)
		service.FeishuSendMessageText(chatId, "chat_id", "Failed to send RSS message")
	}
}

func handleHelp(chatId string) {
	service.FeishuSendMessageText(chatId, "chat_id", config.DocLink)
}
