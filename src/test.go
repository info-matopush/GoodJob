package main

import (
	"net/http"
	"google.golang.org/appengine"
	"src/data/endpoint"
)
func testHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	n := Notification{
		Url:        "",
		Icon:       "/img/icon_001500_256.png",
		Title:      "テスト通知",
		Body:       "これはテスト通知です。",
	}

	ei, _ := endpoint.Get(ctx, r.FormValue("endpoint"))
	if ei != nil {
		SendPush(ctx, &n, ei)
	}
}
