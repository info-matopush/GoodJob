package main

import (
	"net/http"
	"time"
	"math/rand"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/datastore"
	"github.com/mjibson/goon"
	"src/data"
	"encoding/json"
	"src/data/endpoint"
	"html/template"
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
	http.HandleFunc("/d", detailHandler)
}

type DetailData struct {
	Display      string
	CreateDate   time.Time
	SendCount    int64
	RecvCount    int64
	ToMessage    []data.Message
	FromMessage  []data.Message
}

func detailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	memberKey := r.FormValue("m")

	g := goon.NewGoon(r)
	member := data.Member{Key:memberKey}
	err := g.Get(&member)
	if err != nil {
		log.Infof(ctx, "datastore get error. %v", err)
		return;
	}
	message1, message2 := data.GetAllMessage(ctx, memberKey)

	detail := DetailData{
		Display:     member.Display,
		CreateDate:  member.CreateDate,
		SendCount:   member.SendCount,
		RecvCount:   member.RecvCount,
		ToMessage:   message1,
		FromMessage: message2,
	}

	t, err := template.ParseFiles("templates/detail.html")
	if err != nil {
		log.Infof(ctx, "template parse file error. %v", err)
		return
	}

	err = t.Execute(w, detail)
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
	// 送信カウントを+1
	fromMember.SendIncrement(ctx)

	toMember, err := data.GetFromEndpoint(ctx, roomId, to)
	if err != nil {
		return
	}
	// 受信カウントを+1
	toMember.RecvIncrement(ctx)

	// メッセージを登録
	data.AddMessage(ctx, roomId, fromMember, toMember, message)

	// 受信側へPushするのに必要な情報を取得する
	toEndpoint, err := endpoint.Get(ctx, to)
	if err != nil {
		return
	}

	// Pushを実行する
	n := Notification{
		Title: "「" + fromMember.Display + "」さんからGood Job!",
		Body:  message,
		Url:   r.Referer(),
		Icon:  "/img/icon_001500_256.png",
	}
	SendPush(ctx, &n, toEndpoint)
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

