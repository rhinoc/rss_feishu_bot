package handler

import (
	"log"
	"net/http"

	"github.com/rhinoc/rss_feishu_bot/service"
)

func SendRssMessage(w http.ResponseWriter, r *http.Request) {
	records, err := service.GetRecordList()
	if err != nil {
		log.Println("error getting record list", err)
		return
	}

	for _, record := range records {
		err := service.SendRssMessageByRecord(record)
		if err != nil {
			log.Println("error sending rss message for record", record.Id, err)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
