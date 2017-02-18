'use strict';

let _ = function(id) {return document.getElementById(id);};
let registURL = '/api/regist';
let unregistURL = '/api/unregist';
let subscription = null;
let serverKey = null;

window.addEventListener('load', function() {
    if ('serviceWorker' in navigator) {
        _('subscribe').addEventListener('click', togglePushSubscription, false);
        _('test').addEventListener('click', testPush, false);
        fetch('/api/key').then(getServerKey).then(setServerKey);
        navigator.serviceWorker.register('/push.js');
    }
}, false);

function testPush() {
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);

    fetch('/api/test', {
        method: 'post',
        body: data
    });
    document.activeElement.blur();
}

function decodeBase64URL(str) {
    let dec = atob(str.replace(/\-/g, '+').replace(/_/g, '/'));
    let buffer = new Uint8Array(dec.length);
    for(let i = 0 ; i < dec.length ; i++)
        buffer[i] = dec.charCodeAt(i);
    return buffer;
}

function decodeBase64URL(str) {
    let dec = atob(str.replace(/\-/g, '+').replace(/_/g, '/'));
    let buffer = new Uint8Array(dec.length);
    for(let i = 0 ; i < dec.length ; i++)
        buffer[i] = dec.charCodeAt(i);
    return buffer;
}

function getServerKey(resp) {
    return resp.text();
}

function setServerKey(key) {
    serverKey = decodeBase64URL(key);
    navigator.serviceWorker.ready.then(serviceWorkerReady);
}

function serviceWorkerReady(registration) {
    if ('pushManager' in registration) {
        registration.pushManager.getSubscription().then(getSubscription);
    }
    else {
        alert('プッシュ通知を有効にできません。');
    }
}

function togglePushSubscription() {
    if (!_('subscribe').classList.contains('subscribing')) {
        if (_('display').value === '') {
            alert('ニックネームを入力してください。')
            return;
        }
        _('subscribe').disabled = true;
        requestNotificationPermission();
    }
    else {
        _('subscribe').disabled = true;
        requestPushUnsubscription();
    }
}

function requestNotificationPermission() {
    Notification.requestPermission(function(permission) {
        if (permission !== 'denied') {
            requestPushPermission();
        }
    });
}

function requestPushPermission() {
    if ('permissions' in navigator)
        navigator.permissions.query({
            name: 'push',
            userVisibleOnly: true
        }).then(checkPushPermission);
    else if (Notification.permission !== 'denied') {
        navigator.serviceWorker.ready.then(requestPushSubscription);
    }
}

function checkPushPermission(evt) {
    let state = evt.state || evt.status;
    if (state !== 'denied')
        navigator.serviceWorker.ready.then(requestPushSubscription);
}

function requestPushSubscription(registration) {
    let opt = {
        userVisible: true,
        userVisibleOnly: true,
        applicationServerKey: serverKey
    };
    return registration.pushManager.subscribe(opt).then(getSubscription, errorSubscription);
}

function errorSubscription(err) {
    alert('プッシュ通知を有効にできません。' + err);
}

function getSubscription(sub) {
    if (sub) {
        enablePushRequest(sub);
    }
    else {
        disablePushRequest();
    }
}

function requestPushUnsubscription() {
    if (subscription) {
        subscription.unsubscribe();

        // subscriptionを削除する
        var data = new FormData();
        data.append('endpoint', subscription.endpoint);
        fetch(unregistURL, {
            method: 'post',
            body:   data
        }).then(res => {
        });

        subscription = null;
        disablePushRequest();

        // firefox用にクリアする
        _('display').value = '';
        location.reload();
    }
}

function disablePushRequest() {
    _('subscribe').classList.remove('subscribing');
    _('subscribe').disabled = false;
    _('test').disabled = true;
    _('message').disabled = true;
    _('display').disabled = false;
}

function enablePushRequest(sub) {
    subscription = sub;
    _('subscribe').classList.add('subscribing');
    _('subscribe').disabled = false;
    _('test').disabled = false;
    _('message').disabled = false;
    _('display').disabled = true;

    // subscriptionを登録する
    var data = new FormData();
    data.append('endpoint', subscription.endpoint);
    data.append('auth',     encodeBase64URL(subscription.getKey('auth')));
    data.append('p256dh',   encodeBase64URL(subscription.getKey('p256dh')));
    data.append('display',  _('display').value);
    data.append('roomId',   roomId);
    fetch(registURL, {
        method: 'post',
        body:   data
    }).then(function(resp) {
        return resp.text();
    }).then(function(text) {
        _('display').value = text;


        // subscriptionの登録が完了したらメンバーリストを取得する
        fetch('/api/list', {
            method: 'post',
            body: data
        }).then(function(resp) {
            return resp.json();
        }).then(function(json) {
            console.log(json);
            // 動的にリストを作る
            for (var i = 0; i < json.length; i++) {
                let display = json[i].Display;
                let endpoint = json[i].Endpoint;

                let button = document.createElement('button');
                button.type = 'button';
                button.id = endpoint;
                button.class = 'btn btn-success';
                button.textContent = '相手に送信する';
                button.onclick = function(){sendMessage(endpoint)};

                let nickname = document.createElement('a');
                nickname.href = '/d?m=' + json[i].Key;
                nickname.textContent = display;

                let tr = document.createElement('tr');
                let td1 = document.createElement('td');
                if (endpoint !== subscription.endpoint) {
                    td1.appendChild(button);
                }
                tr.appendChild(td1);
                let td2 = document.createElement('td');
                td2.appendChild(nickname);
                tr.appendChild(td2);
                let td3 = document.createElement('td');
                td3.textContent = json[i].SendCount;
                tr.appendChild(td3);
                let td4 = document.createElement('td');
                td4.textContent = json[i].RecvCount;
                tr.appendChild(td4);

                _('memberList').appendChild(tr);
            }
        });
    });
}

function sendMessage(to) {
    if (_('message').value === '') {
        alert('メッセージを入力してください。');
        return;
    }

    var data = new FormData();
    data.append('from', subscription.endpoint);
    data.append('to',   to);
    data.append('message', _('message').value);
    data.append('roomId', roomId);

    fetch('/api/send', {
        method: 'post',
        body: data
    }).then(function(){
        // メッセージ送信後に再読み込みする
        _('message').value = '';
        location.reload();
    });
}

function encodeBase64URL(buffer) {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(buffer))).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
