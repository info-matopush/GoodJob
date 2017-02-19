package data

import (
	"time"
	"golang.org/x/net/context"
	"github.com/mjibson/goon"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/datastore"
)

type MessageInfo struct {
	FromDisplay string
	ToDisplay   string
	Date        string
	Message     string
}

type Message struct {
	Id          int64      `datastore:"-" goon:"id"`
	RoomId      string     `datastore:"roomId"`
	FromMember  string     `datastore:"from"`
	ToMember    string     `datastore:"to"`
	FromDisplay string     `datastore:"from_display,noindex"`
	ToDisplay   string     `datastore:"to_display,noindex"`
	Date        time.Time  `datastore:"date"`
	Message     string     `datastore:"message,noindex"`
}

func AddMessage(ctx context.Context, roomId string, fromMember, toMember Member, message string) {
	g := goon.FromContext(ctx)

	m := Message{
		RoomId:      roomId,
		FromMember:  fromMember.Key,
		ToMember:    toMember.Key,
		FromDisplay: fromMember.Display,
		ToDisplay:   toMember.Display,
		Date:        time.Now(),
		Message:     message,
	}

	g.Put(&m)
}

const (
	date_format = "2006-01-02 15:04:05"
)

func getMessageInfo(ctx context.Context, query *datastore.Query) ([]MessageInfo) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	g := goon.FromContext(ctx)
	it := g.Run(query)

	list := []MessageInfo{}
	for {
		var s Message
		_, err := it.Next(&s)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "datastore get error.%v", err)
			break
		}

		m := MessageInfo{
			FromDisplay: s.FromDisplay,
			ToDisplay:   s.ToDisplay,
			Date:        s.Date.In(jst).Format(date_format),
			Message:     s.Message,
		}
		list = append(list, m)
	}
	return list
}

func GetAllMessageInfo(ctx context.Context, memberKey string) ([]MessageInfo, []MessageInfo) {

	query1 := datastore.NewQuery("Message").Filter("from=", memberKey).Order("-date")
	message1 := getMessageInfo(ctx, query1)

	query2 := datastore.NewQuery("Message").Filter("to=", memberKey).Order("-date")
	message2 := getMessageInfo(ctx, query2)

	return message1, message2
}
