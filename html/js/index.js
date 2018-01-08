function App() {
    var app = this;

    this.vue = new Vue({
      el: '#app',
      data: {
        clientsInfo: {
          yourId: '',
          yourNickname: '',
          clients: []
        },
        commandError: {},
        rooms: []
      }
    });

    this.onMessage = function(msg) {
        const onEventMethodName = 'on' + msg.name;
        let methodFound = false;
        for (prop in app) {
            if (prop === onEventMethodName) {
                app[onEventMethodName](msg.data);
                methodFound = true;
                break;
            }
        }
        if (!methodFound) {
          console.warn("Method not found for event", msg);
        }
        console.log(msg.name, msg.data);
    };

    this.onClientJoinedEvent = function(data) {
        app.vue.clientsInfo.yourId = data.your_id;
        app.vue.clientsInfo.yourNickname = data.your_nickname;
        app.vue.clientsInfo.clients = data.clients;
        app.vue.rooms = data.rooms;
    };

    this.onClientBroadCastJoinedEvent = function(data) {
        app.vue.clientsInfo.clients.push(data);
    };

    this.onClientLeftEvent = function(data) {
        let clients = app.vue.clientsInfo.clients;
        for (let ind = 0; ind < clients.length; ind++) {
            if (clients[ind].id === data.id) {
                clients.splice(ind, 1);
            }
        }
        app.vue.clientsInfo.clients = clients;
    }

    this.onClientCommandError = function(data) {
        app.vue.commandError = data;
        console.error(data.message);
    }

    this.onClientCreatedRoomEvent = function(data) {
        app.vue.rooms.push(data.room);
    }
}

(function(){
  var app = new App();
  OnIncomingMessage = app.onMessage;
})();


let sendBtn = document.getElementById('send');
sendBtn.onclick = function() {
    WsConnection.send(JSON.stringify({type: 'lobby', sub_type: 'create_room', data: {}}));
};
