package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rhinoc/rss_feishu_bot/config"
	"github.com/rhinoc/rss_feishu_bot/util"
)

type RecordItemFeed struct {
	Link         string `json:"link"`
	LastReadLink string `json:"last_read_link"`
}

type RecordItem struct {
	Id          string            `json:"id"`
	UserOpenId  string            `json:"user_open_id"`
	GroupOpenId string            `json:"group_open_id"`
	FeedList    []*RecordItemFeed `json:"feed_list"`
}

func GetRecordList() ([]RecordItem, error) {
	records, err := FeishuGetBitableRecord(config.BitableAppToken, config.BitableTableId, FeishuGetBitableRecordRequest{
		ViewId: config.BitableViewId,
		Filter: BitableRecordFilterInfo{
			Conjunction: "and",
			Conditions: []BitableRecordCondition{
				{
					FieldName: "feedList",
					Operator:  "isNotEmpty",
					Value:     []string{},
				},
				{
					FieldName: "enable",
					Operator:  "is",
					Value:     []string{"true"},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	recordItems := make([]RecordItem, 0)
	for _, record := range records.Data.Items {
		recordItems = append(recordItems, getRecordItem(record))
	}

	recordItems = util.Filter(recordItems, func(item RecordItem) bool {
		return item.GroupOpenId != "" || item.UserOpenId != ""
	})

	return recordItems, nil
}

func GetRecordItem(id string, isGroup bool) (*RecordItem, error) {
	field := "user"
	if isGroup {
		field = "group"
	}

	records, err := FeishuGetBitableRecord(config.BitableAppToken, config.BitableTableId, FeishuGetBitableRecordRequest{
		ViewId: config.BitableViewId,
		Filter: BitableRecordFilterInfo{
			Conjunction: "and",
			Conditions: []BitableRecordCondition{
				{
					FieldName: field,
					Operator:  "is",
					Value:     []string{id},
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	if len(records.Data.Items) == 0 {
		return nil, fmt.Errorf("record not found")
	}

	if len(records.Data.Items) > 1 {
		log.Println("multiple records found", records.Data.Items)
	}

	item := getRecordItem(records.Data.Items[0])
	return &item, nil
}

func getLastReadLinkMap(source []interface{}) map[string]string {
	jsonStr := ""

	for _, item := range source {
		if sourceItem, ok := item.(map[string]interface{}); ok {
			if text, ok := sourceItem["text"].(string); ok {
				jsonStr += text
			}
		}
	}

	jsonMap := make(map[string]string)
	json.Unmarshal([]byte(jsonStr), &jsonMap)

	return jsonMap
}

func getRecordItem(source FeishuGetBitableRecordItem) RecordItem {
	item := RecordItem{
		Id: source.RecordID,
	}

	// Extract user_open_id
	if users, ok := source.Fields["user"].([]interface{}); ok && len(users) > 0 {
		if user, ok := users[0].(map[string]interface{}); ok {
			if id, ok := user["id"].(string); ok {
				item.UserOpenId = id
			}
		}
	}

	// Extract group_open_id
	if groups, ok := source.Fields["group"].([]interface{}); ok && len(groups) > 0 {
		if group, ok := groups[0].(map[string]interface{}); ok {
			if id, ok := group["id"].(string); ok {
				item.GroupOpenId = id
			}
		}
	}

	// Extract feed_list
	feedList, _ := source.Fields["feedList"].([]interface{})
	lastReadLinkList := make(map[string]string)
	if source.Fields["lastReadLinkList"] != nil {
		lastReadLinkList = getLastReadLinkMap(source.Fields["lastReadLinkList"].([]interface{}))
	}

	if len(feedList) > 0 {
		item.FeedList = make([]*RecordItemFeed, 0, len(feedList))
		for _, feedLink := range feedList {
			item.FeedList = append(item.FeedList, &RecordItemFeed{
				Link:         feedLink.(string),
				LastReadLink: lastReadLinkList[feedLink.(string)],
			})
		}
	}

	return item
}

func UpdateRecordItemLastReadLink(recordItem RecordItem) error {
	lastReadLinkList := make(map[string]interface{})
	for _, feed := range recordItem.FeedList {
		lastReadLinkList[feed.Link] = feed.LastReadLink
	}

	return FeishuUpdateBitableRecord(config.BitableAppToken, config.BitableTableId, recordItem.Id, FeishuUpdateBitableRecordRequest{
		Fields: map[string]interface{}{
			"lastReadLinkList": string(util.Must(json.Marshal(lastReadLinkList))),
		},
	})
}

func UpdateRecordItemFeedList(recordItem RecordItem) error {
	feedList := make([]string, 0, len(recordItem.FeedList))
	for _, feed := range recordItem.FeedList {
		feedList = append(feedList, feed.Link)
	}
	return FeishuUpdateBitableRecord(config.BitableAppToken, config.BitableTableId, recordItem.Id, FeishuUpdateBitableRecordRequest{
		Fields: map[string]interface{}{
			"feedList": feedList,
		},
	})
}

func AddRecordItem(recordItem RecordItem) (string, error) {
	feedList := make([]string, 0, len(recordItem.FeedList))
	for _, feed := range recordItem.FeedList {
		feedList = append(feedList, feed.Link)
	}

	fields := map[string]interface{}{
		"feedList": feedList,
		"enable":   true,
	}

	if recordItem.UserOpenId != "" {
		fields["user"] = []map[string]interface{}{
			{
				"id": recordItem.UserOpenId,
			},
		}
	}

	if recordItem.GroupOpenId != "" {
		fields["group"] = []map[string]interface{}{
			{
				"id": recordItem.GroupOpenId,
			},
		}
	}

	return FeishuAddBitableRecord(config.BitableAppToken, config.BitableTableId, FeishuAddBitableRecordRequest{
		Fields: fields,
	})
}
