package main

import (
	"src/data"
	"fmt"
	"net/http"
	"github.com/mjibson/goon"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"math/rand"
	"time"
	"html/template"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func roomAddHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	g := goon.NewGoon(r)

	d := r.FormValue("description")
	if d == "" {
		// descriptionがないものはエラー
		fmt.Fprint(w, "/fail.html")
		return
	}

	log.Infof(ctx, "desc %s", d)

	roomId := RandString(32)
	room := data.Room{RoomId:roomId}
	err := g.Get(&room)
	if err == nil {
		// すでに作成済みのRoomIdだった
		fmt.Fprint(w, "/fail.html")
		return
	}

	// ルームを新規作成する
	room.Description = d
	room.CreateDate = time.Now()
	_, err = g.Put(&room)
	if err != nil {
		fmt.Fprint(w, "/fail.html")
	} else {
		fmt.Fprintf(w, "/s?r=%s", roomId)
	}
}

type RoomInfo struct {
	RoomId      string
	Url         string
	Description string
	CreateDate  time.Time
}

func showRoomHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	roomId := r.FormValue("r")

	roomUrl := r.Referer() + "e?r=" + roomId

	g := goon.NewGoon(r)
	room := data.Room{RoomId:roomId}
	err := g.Get(&room)
	if err != nil {
		log.Infof(ctx, "datastore get error. %v", err)
		return;
	}

	ri := RoomInfo{
		RoomId:      room.RoomId,
		Url:         roomUrl,
		Description: room.Description,
		CreateDate:  room.CreateDate,
	}
	t, err := template.ParseFiles("templates/show.html")
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

