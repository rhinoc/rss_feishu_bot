package service

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rhinoc/rss_feishu_bot/config"
	"github.com/rhinoc/rss_feishu_bot/model"
	"github.com/rhinoc/rss_feishu_bot/util"
)

type RssFeed struct {
	Title     string
	Link      string
	UpdatedAt *time.Time
	Items     []RssFeedItem
}

type RssFeedItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
}

func GetRssFeed(url string, limit int) (*RssFeed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)

	if err != nil {
		return nil, fmt.Errorf("error parsing rss feed: %w", err)
	}

	items := []RssFeedItem{}
	for i, item := range feed.Items {
		if i >= limit {
			break
		}
		items = append(items, RssFeedItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Published,
		})
	}

	return &RssFeed{
		Title:     feed.Title,
		Link:      feed.Link,
		UpdatedAt: feed.UpdatedParsed,
		Items:     items,
	}, nil
}

func FetchRSSFeedsParallel(feedURLs []string) ([]*gofeed.Feed, error) {
	var wg sync.WaitGroup
	feeds := make([]*gofeed.Feed, len(feedURLs))
	errors := make([]error, len(feedURLs))

	for i, url := range feedURLs {
		wg.Add(1)
		go func(index int, feedURL string) {
			defer wg.Done()
			fp := gofeed.NewParser()
			feed, err := fp.ParseURL(feedURL)
			if err != nil {
				errors[index] = err
				return
			}
			feeds[index] = feed
		}(i, url)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return feeds, nil
}

func SendRssMessageByRecord(recordItem RecordItem) error {
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	rssResults := make([][]model.FeishuMessageItem, len(recordItem.FeedList))

	for recordIndex, recordItemFeed := range recordItem.FeedList {
		wg.Add(1)
		go func() {
			defer wg.Done()
			feed, err := GetRssFeedByRecordItemFeed(recordItemFeed)
			if err != nil {
				log.Println("error getting rss feed", err)
				return
			}
			rssResult := []model.FeishuMessageItem{}
			for _, item := range feed.Items {
				rssResult = append(rssResult, model.FeishuMessageItem{
					Title:            item.Title,
					Link:             item.Link,
					PrimaryDesc:      feed.Title,
					PrimaryDescColor: util.GetColorByIndex(recordIndex),
					SecondaryDesc:    item.Description,
				})
			}
			mu.Lock()
			rssResults[recordIndex] = rssResult
			if len(feed.Items) > 0 {
				recordItemFeed.LastReadLink = feed.Items[0].Link
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	items := []model.FeishuMessageItem{}
	for _, rssResult := range rssResults {
		items = append(items, rssResult...)
	}

	if len(items) == 0 {
		log.Println("no newer feeds found for record", recordItem.Id)
		return nil
	}

	date := time.Now().Format("2006-01-02")
	content := &model.FeishuMessageContent{
		Type: "template",
		Data: model.FeishuMessageData{
			TemplateId:          config.CardTemplateId,
			TemplateVersionName: config.CardTemplateVersionName,
			TemplateVariable: map[string]interface{}{
				// example: 2006-01-02 | Explore 10 New Updates
				"cardTitle": fmt.Sprintf("%s | Explore %d New Updates", date, len(items)),
				"cardColor": "blue",
				"itemList":  items,
			},
		},
	}

	targetOpenId := recordItem.UserOpenId
	if recordItem.GroupOpenId != "" {
		targetOpenId = recordItem.GroupOpenId
	}

	receiveIdType := "open_id"
	if recordItem.GroupOpenId != "" {
		receiveIdType = "chat_id"
	}

	err := FeishuSendMessage(FeishuSendMessageRequest{
		ReceiveId:     targetOpenId,
		ReceiveIdType: receiveIdType,
		MsgType:       "interactive",
		Content:       string(util.Must(json.Marshal(content))),
	})

	if err != nil {
		return fmt.Errorf("error sending message for record %s: %w", recordItem.Id, err)
	}

	err = UpdateRecordItemLastReadLink(recordItem)
	if err != nil {
		return fmt.Errorf("error updating record item last read link for record %s: %w", recordItem.Id, err)
	}

	return nil
}

func GetRssFeedByRecordItemFeed(recordItemFeed *RecordItemFeed) (*RssFeed, error) {
	feed, err := GetRssFeed(recordItemFeed.Link, config.DefaultItemLimitPerFeed)
	if err != nil {
		return nil, err
	}

	if recordItemFeed.LastReadLink != "" {
		for i, item := range feed.Items {
			if item.Link == recordItemFeed.LastReadLink {
				feed.Items = feed.Items[:i]
				break
			}
		}
	}

	return feed, nil
}
