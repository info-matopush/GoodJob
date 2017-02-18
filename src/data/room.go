package data

import "time"

type Room struct {
	RoomId       string    `datastore:"-" goon:"id"`
	Description  string    `datastore:"description,noindex"`
	CreateDate   time.Time `datastore:"create_date,noindex"`
	AccessDate   time.Time `datastore:"access_date,noindex"`
}
