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
        room: {},
        wantToPlay: true,
        playersInRoom: 0,
        game: {
          players: [],
          yourPlayerIndex: -1
        }
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
        },
        markWantToPlay: function () {
          app.commandWantToPlay();
          this.wantToPlay = true;
        },
        markWantToSpectate: function () {
          app.commandWantToSpectate();
          this.wantToPlay = false;
        },
        setPlayerStatus: function (memberId, status) {
          app.commandSetPlayerStatus(memberId, status);
        },
        startGame: function () {
          app.commandStartGame();
        }
      }
    });

    this.onMessage = function (msg) {
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

    this.onClientJoinedEvent = function (data) {
        app.vue.clientsInfo.yourId = data.your_id;
        app.vue.clientsInfo.yourNickname = data.your_nickname;
        app.vue.clientsInfo.clients = data.clients;
        app.vue.rooms = data.rooms;
        if (data.rooms.length === 0) {
          // auto create
          app.commandCreateRoom();
        } else if (data.rooms.length === 1 && data.rooms[0].members_num === 1) {
          // auto join
          app.commandJoinRoom(data.rooms[0].id);
          app.vue.clientsInfo.yourRoomId = data.rooms[0].id;
        }
    };

    this.onClientBroadCastJoinedEvent = function (data) {
        app.vue.clientsInfo.clients.push(data);
    };

    this.onClientLeftEvent = function (data) {
        let clients = app.vue.clientsInfo.clients;
        for (let ind = 0; ind < clients.length; ind++) {
            if (clients[ind].id === data.id) {
                clients.splice(ind, 1);
            }
        }
        app.vue.clientsInfo.clients = clients;
    }

    this.onRoomInListUpdatedEvent = function (data) {
        const index = app.getRoomIndexById(data.room.id);
        if (index > -1) {
          Vue.set(app.vue.rooms, index, data.room);
        }
    }

    this.onRoomInListRemovedEvent = function (data) {
        const index = app.getRoomIndexById(data.room_id);
        if (index > -1) {
          app.vue.rooms.splice(index, 1);
        } else {
          console.warn("Can't remove room", data.room_id);
        }
    }

    this.onRoomJoinedEvent = function (data) {
        app.vue.room = data.room;
        app.updatePlayersInRoomCounter();
    }

    this.onRoomUpdatedEvent = function (data) {
        app.vue.room = data.room;
        app.updatePlayersInRoomCounter();
    }

    this.onClientCommandError = function (data) {
        app.vue.commandError = data;
        console.error(data.message);
    }

    this.onClientCreatedRoomEvent = function (data) {
        app.vue.rooms.push(data.room);
        if (data.room.owner_id === app.vue.clientsInfo.yourId) {
          app.vue.clientsInfo.yourRoomId = data.room.id;
        }
    }

    this.onRoomMemberChangedStatusEvent = function (data) {
        if (data.member.id === app.vue.clientsInfo.yourId) {
            app.vue.wantToPlay = data.member.want_to_play;
        }
        const memberIndex = app.getRoomMemberIndexById(data.member.id);
        if (memberIndex > -1) {
          Vue.set(app.vue.room.members, memberIndex, data.member);
        }
    }

    this.onRoomMemberChangedPlayerStatusEvent = function (data) {
        const memberIndex = app.getRoomMemberIndexById(data.member.id);
        if (memberIndex > -1) {
          Vue.set(app.vue.room.members, memberIndex, data.member);
        }
        app.updatePlayersInRoomCounter();
    }

    this.onGamePlayersEvent = function (data) {
        app.vue.game.players = data.players;
        app.vue.game.yourPlayerIndex = data.your_player_index;
    }

    this.sendCommand = function (type, subType, data) {
        console.log("send", type, subType, data);
        WsConnection.send(JSON.stringify({type: type, sub_type: subType, data: data}));
    }

    this.commandJoinRoom = function (roomId) {
        app.sendCommand('lobby', 'join_room', roomId);
    }

    this.commandCreateRoom = function (roomId) {
        app.sendCommand('lobby', 'create_room', null);
    }

    this.commandWantToPlay = function () {
        app.sendCommand('room', 'want_to_play', null);
    }

    this.commandWantToSpectate = function () {
        app.sendCommand('room', 'want_to_spectate', null);
    }

    this.commandSetPlayerStatus = function (memberId, status) {
        app.sendCommand('room', 'set_player_status', {member_id: memberId, status: status});
    }

    this.commandStartGame = function () {
        app.sendCommand('room', 'start_game', null);
    }

    this.getRoomIndexById = function (roomId) {
      for (let i = 0; i < app.vue.rooms.length; i++) {
        if (app.vue.rooms[i].id === roomId) {
          return i;
        }
      }
      return -1;
    }

    this.getRoomMemberIndexById = function (memberId) {
      if (!app.vue.room) {
        console.warn("No room for event");
        return -1;
      }
      for (let i = 0; i < app.vue.room.members.length; i++) {
        if (app.vue.room.members[i].id === memberId) {
          return i;
        }
      }
      return -1;
    }

    this.updatePlayersInRoomCounter = function() {
      let playersNum = 0;
      for (let i = 0; i < app.vue.room.members.length; i++) {
        if (app.vue.room.members[i].is_player) {
          playersNum++;
        }
      }
      app.vue.playersInRoom = playersNum;
    }
}

(function(){
  var app = new App();
  OnIncomingMessage = app.onMessage;
})();
