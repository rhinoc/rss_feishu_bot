package handler

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rhinoc/rss_feishu_bot/service"
)

func GetRecordList(w http.ResponseWriter, r *http.Request) {
	recordItems, err := service.GetRecordList()
	if err != nil {
		log.Println("error getting record list", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, err)
		return
	}

	render.JSON(w, r, recordItems)
}
