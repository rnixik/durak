function App() {
    var app = this;

    this.vue = new Vue({
      el: '#app',
      data: {
        clientsInfo: {
          yourId: '',
          yourNickname: '',
          yourRoomId: '',
          clients: []
        },
        commandError: {},
        rooms: [],
        room: {}
      },
      methods: {
        createRoom: function (event) {
          if (event) { 
            event.preventDefault();
          }
          app.commandCreateRoom();
        },
        joinRoom: function (roomId, event) {
          if (event) { 
            event.preventDefault();
          }
          app.commandJoinRoom(roomId);
          app.vue.clientsInfo.yourRoomId = roomId;
        }
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

    this.onRoomInListUpdatedEvent = function(data) {
        const index = app.getRoomIndexById(data.room.id);
        if (index > -1) {
          Vue.set(app.vue.rooms, index, data.room);
        }
    }

    this.onRoomInListRemovedEvent = function(data) {
        const index = app.getRoomIndexById(data.room_id);
        if (index > -1) {
          app.vue.rooms.splice(index, 1);
        } else {
          console.warn("Can't remove room", data.room_id);
        }
    }

    this.onRoomJoinedEvent = function(data) {
        app.vue.room = data.room;
    }

    this.onRoomUpdatedEvent = function(data) {
        app.vue.room = data.room;
    }

    this.onClientCommandError = function(data) {
        app.vue.commandError = data;
        console.error(data.message);
    }

    this.onClientCreatedRoomEvent = function(data) {
        app.vue.rooms.push(data.room);
        if (data.room.owner_id === app.vue.clientsInfo.yourId) {
          app.vue.clientsInfo.yourRoomId = data.room.id;
        }
    }

    this.sendCommand = function(type, subType, data) {
        WsConnection.send(JSON.stringify({type: type, sub_type: subType, data: data}));
    }

    this.commandJoinRoom = function(roomId) {
        app.sendCommand('lobby', 'join_room', roomId);
    }

    this.commandCreateRoom = function(roomId) {
        app.sendCommand('lobby', 'create_room', null);
    }


    this.getRoomIndexById = function(roomId) {
      for (let i = 0; i < app.vue.rooms.length; i++) {
        if (app.vue.rooms[i].id === roomId) {
          return i;
        }
      }
      return -1;
    }
}

(function(){
  var app = new App();
  OnIncomingMessage = app.onMessage;
})();


let sendBtn = document.getElementById('send');
sendBtn.onclick = function() {
    //WsConnection.send(JSON.stringify({type: 'lobby', sub_type: 'create_room', data: {}}));
};
