self.addEventListener('push', function(evt) {
    var object = evt.data.json();
    var title = 'タイトルなし';
    var body = '';
    var url = '';
    var tag = '';
    var icon = '/img/icon_001500_256.png';
    if ('Title' in object) {
        title = object.Title;
    }
    if ('Body' in object) {
        body = object.Body;
    }
    if ('Url' in object) {
        url = object.Url;
    }
    if ('Icon' in object) {
        icon = object.Icon;
    }
    if ('Tag' in object) {
        tag = object.Tag;
    }

    if (body !== '') {
        evt.waitUntil(
            self.registration.showNotification(
                title,
                {
                    body:    body,
                    data:    {
                        url:       url,
                    },
                    icon:    icon,
                    tag:     tag,
                }
            )
        )
    }
});

self.addEventListener('notificationclick', function(evt) {
    var url = evt.notification.data.url;
    evt.notification.close();

    // URLが指定されていれば遷移する
    if (url !== "") {
        return clients.openWindow(url);
    }
});
