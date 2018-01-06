OnIncomingMessage = function(msg) {
    console.log(msg.name, msg.data);
};

let sendBtn = document.getElementById('send');
sendBtn.onclick = function() {
    WsConnection.send(JSON.stringify({type: 'game', sub_type: 'attack', data: {}}));
};
