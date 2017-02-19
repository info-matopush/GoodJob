package data

import (
	"time"
	"hash/fnv"
	"encoding/base64"
	"golang.org/x/net/context"
	"github.com/mjibson/goon"
)

type Member struct {
	Key          string     `datastore:"-" goon:"id"`
	RoomId       string     `datastore:"room_id"`
	Endpoint     string     `datastore:"endpoint,noindex"`
	Display      string     `datastore:"display,noindex"`
	CreateDate   time.Time  `datastore:"create_date,noindex"`
	Count        int64      `datastore:"count,noindex"`
	SendCount    int64      `datastore:"send_count,noindex"`
	RecvCount    int64      `datastore:"recv_count,noindex"`
	AccessDate   time.Time  `datastore:"access_date,noindex"`
	LastSendDate time.Time  `datastore:"last_send_date,noindex"`
}

// roomId + ":" + endpointは文字列として長すぎるので、ハッシュを使ってキーを作成する
func RoomMemberToKeyString(roomId, endpoint string) string {
	h := fnv.New64a()
	h.Write([]byte(roomId + ":" + endpoint))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func MemberTouch(ctx context.Context, roomId, endpoint, display string) {
	key := RoomMemberToKeyString(roomId, endpoint)
	m := Member{Key:key}
	g := goon.FromContext(ctx)

	err := g.Get(&m)
	if err != nil {
		m.RoomId = roomId
		m.Endpoint = endpoint
		m.Display = display
		m.CreateDate = time.Now()
		m.AccessDate = time.Now()
	} else {
		m.AccessDate = time.Now()
	}
	g.Put(&m)
}

func GetFromEndpoint(ctx context.Context, roomId, endpoint string) (Member, error) {
	key := RoomMemberToKeyString(roomId, endpoint)
	m := Member{Key:key}
	g := goon.FromContext(ctx)

	err := g.Get(&m)
	return m, err
}

func (m *Member) CountUp(ctx context.Context) {
	m.Count++;
	g := goon.FromContext(ctx)
	g.Put(m)
}

func (m *Member) SendIncrement(ctx context.Context) {
	m.SendCount++;
	g := goon.FromContext(ctx)
	g.Put(m)
}

func (m *Member) RecvIncrement(ctx context.Context) {
	m.RecvCount++;
	g := goon.FromContext(ctx)
	g.Put(m)
}