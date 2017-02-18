'use strict'

let _ = function(id) {return document.getElementById(id);}

window.addEventListener('load', function() {
    _('addRoom').addEventListener('click', addRoom, false);
})

function addRoom() {
    if (_('room_description').value === '') {
        alert('説明を入力してください。');
        return;
    }

    var data = new FormData();
    data.append('description', _('room_description').value);
    fetch('/api/add', {
        method: 'post',
        body: data
    }).then(function(resp){
        return resp.text();
    }).then(function(text){
        location.replace(text);
    })
}