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
	AccessDate   time.Time  `datastore:"access_date,noindex"`
	LastSendDate time.Time  `datastore:"last_send_date,noindex"`
}

// roomId + ":" + endpointは文字列として長すぎるので、ハッシュを使ってキーを作成する
func roomMemberToKeyString(roomId, endpoint string) string {
	h := fnv.New64a()
	h.Write([]byte(roomId + ":" + endpoint))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func MemberTouch(ctx context.Context, roomId, endpoint, display string) {
	key := roomMemberToKeyString(roomId, endpoint)
	m := Member{Key:key}
	g := goon.FromContext(ctx)

	err := g.Get(&m)
	if err != nil {
		// メンバー初回登録
		if display == "" {
			display = "名無し"
		}

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
	key := roomMemberToKeyString(roomId, endpoint)
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