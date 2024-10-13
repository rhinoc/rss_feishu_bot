package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rhinoc/rss_feishu_bot/handler"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/feishu", func(r chi.Router) {
		r.Post("/callback", handler.FeishuCallbackHandler)
		r.Post("/event", handler.FeishuCallbackHandler)
	})
	r.Route("/rss", func(r chi.Router) {
		r.Get("/send", handler.SendRssMessage)
	})
	r.Route("/record", func(r chi.Router) {
		r.Get("/list", handler.GetRecordList)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}
	http.ListenAndServe(":"+port, r)
}
