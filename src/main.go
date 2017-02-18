package main

import (
	"net/http"
	"time"
	"math/rand"
	"html/template"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/datastore"
	"github.com/mjibson/goon"
	"src/data"
	"encoding/json"
	"src/data/endpoint"
)

func init() {
	// 乱数のシード値初期化
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/api/regist", registHandler)
	http.HandleFunc("/api/unregist", unregistHandler)
	http.HandleFunc("/api/key", keyHandler)
	http.HandleFunc("/api/add", roomAddHandler)
	http.HandleFunc("/api/send", sendHandler)
	http.HandleFunc("/api/list", listHandler)
	http.HandleFunc("/api/test", testHandler)
	http.HandleFunc("/s", showRoomHandler)
	http.HandleFunc("/e", enterRoomHandler)
}

func enterRoomHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	roomId := r.FormValue("r")

	g := goon.NewGoon(r)
	room := data.Room{RoomId:roomId}
	err := g.Get(&room)
	if err != nil {
		log.Infof(ctx, "datastore get error. %v", err)
		return;
	}

	ri := RoomInfo{
		RoomId:      room.RoomId,
		Url:         "",
		Description: room.Description,
		CreateDate:  room.CreateDate,
	}
	t, err := template.ParseFiles("templates/enter.html")
	if err != nil {
		log.Infof(ctx, "template parse file error. %v", err)
		return
	}

	err = t.Execute(w, ri)
	if err != nil {
		log.Infof(ctx, "template execute error. %v", err)
		return
	}
}

func sendHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	from := r.FormValue("from")
	to := r.FormValue("to")
	message := r.FormValue("message")
	roomId := r.FormValue("roomId")

	fromMember, err := data.GetFromEndpoint(ctx, roomId, from)
	if err != nil {
		return
	}

	toEndpoint, err := endpoint.Get(ctx, to)
	if err != nil {
		return
	}

	n := Notification{
		Title: "「" + fromMember.Display + "」さんからGood Job!",
		Body:  message,
		Url:   r.Referer(),
		Icon:  "/img/icon_001500_256.png",
	}

	SendPush(ctx, &n, toEndpoint)
	fromMember.CountUp(ctx)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	g := goon.NewGoon(r)

	//endpoint := r.FormValue("endpoint")
	roomId := r.FormValue("roomId")

	query := datastore.NewQuery("Member").Filter("room_id=", roomId)
	it := g.Run(query)

	list := []data.Member{}
	for {
		var s data.Member
		_, err := it.Next(&s)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "datastore get error.%v", err)
			break
		}

		_, err = endpoint.Get(ctx, s.Endpoint)
		if err != nil {
			// そのメンバーは存在しない
			continue;
		}

		list = append(list, s)
	}

	b, _ := json.Marshal(list)
	w.Write(b)
}

