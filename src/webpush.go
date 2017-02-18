package main

import (
	"encoding/base64"
	"errors"
	webpush "github.com/SherClockHolmes/webpush-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"src/data/endpoint"
	"src/data"
	"fmt"
	"time"
)

type Notification struct {
	Title       string
	Icon        string
	Body        string
	Url         string
	Tag         string
}

func registHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	auth, _ := base64.RawURLEncoding.DecodeString(r.FormValue("auth"))
	p256dh, _ := base64.RawURLEncoding.DecodeString(r.FormValue("p256dh"))
	e := r.FormValue("endpoint")
	d := r.FormValue("display")
	if e == "" {
		// 必須データなし
		return
	}

	ei := &endpoint.EndpointInfo{
		Endpoint: e,
		Auth:     auth,
		P256dh:   p256dh,
		Display:  d,
	}

	endpoint.Touch(ctx, ei)

	roomId := r.FormValue("roomId")
	data.MemberTouch(ctx, roomId, e, d)

	// Memberに反映されるのをちょっとだけ待つ
	time.Sleep(1 * time.Second)

	// 登録されている表示名（ニックネーム）を返す
	ei, _ = endpoint.Get(ctx, ei.Endpoint)

	fmt.Fprint(w, ei.Display)
}

func unregistHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	endpoint.Delete(ctx, r.FormValue("endpoint"))
}

func SendPush(ctx context.Context, n *Notification, ei *endpoint.EndpointInfo) (err error) {
	message, _ := json.Marshal(n)

	client := urlfetch.Client(ctx)
	b64 := base64.RawURLEncoding

	var sub webpush.Subscription
	sub.Endpoint = ei.Endpoint
	sub.Keys.Auth = b64.EncodeToString(ei.Auth)
	sub.Keys.P256dh = b64.EncodeToString(ei.P256dh)

	pri, err := getPrivateKey(ctx)
	if err != nil {
		log.Errorf(ctx, "private key get error.%v", err)
		return
	}

	resp, err := webpush.SendNotification(message, &sub, &webpush.Options{
		HTTPClient:      client,
		Subscriber:      "https://push2ch.appspot.com",
		TTL:             60,
		VAPIDPrivateKey: b64.EncodeToString(pri.D.Bytes()),
	})

	if err != nil {
		log.Errorf(ctx, "SendNotification error. %v", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 0 {
		log.Errorf(ctx, "send notification return code 0.")
		err = errors.New("send notification return code 0.")
	} else if resp.StatusCode == http.StatusOK {
		return
	} else if resp.StatusCode == http.StatusCreated {
		return
	} else if resp.StatusCode == http.StatusGone {
		endpoint.Delete(ctx, ei.Endpoint)
		err = errors.New("endpoint was gone.")
	} else {
		log.Infof(ctx, "resp %s", resp)
		buf, _ := ioutil.ReadAll(resp.Body)
		log.Infof(ctx, "body %v", string(buf))
		err = errors.New("unknown response.")
	}
	return
}



