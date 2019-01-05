Vue.component('playing-card', {
    template: '#playing-card-template',
    props: {
        card: {
            type: Object,
            required: true
        }
    },
    computed: {
        className: function () {
            const replaces = {
                'J': 'jack',
                'Q': 'queen',
                'K': 'king',
                'A': 'ace',
                '♣': 'clubs',
                '♦': 'diamonds',
                '♥': 'hearts',
                '♠': 'spades',
            };
            let cssClass = this.card.value + '_of_' + this.card.suit;
            for (let orig in replaces) {
                cssClass = cssClass.replace(new RegExp(orig, "g"), replaces[orig]);
            }
            return cssClass;
        }
    }
});

Vue.component('playing-card-inhand', {
    props: {
        card: {
            type: Object,
            required: true
        }
    },
    template: '#playing-card-inhand-template'
});

Vue.component('playing-card-battleground', {
    props: {
        card: {
            type: Object,
            required: true
        },
        defendingCard: {
            type: Object
        },
        highlighted: {
            type: Boolean
        }
    },
    template: '#playing-card-battleground-template'
});

Vue.component('playing-card-back', {
    template: '#playing-card-back-template'
});


Vue.component('opponent', {
    template: '#opponent-template',
    props: {
        handSize: {
            type: Number,
            required: true
        },
        nickname: {
            type: String,
            required: true
        },
        isAttacker: {
            type: Boolean
        },
        isDefender: {
            type: Boolean
        }
    }
});

Vue.component('deck', {
    template: '#deck-template',
    props: {
        deckSize: {
            type: Number,
            required: true
        },
        trumpCard: {
            type: Object,
            required: true
        },
        trumpSuit: {
            type: String,
            required: true
        }
    }
});

function App() {
    const app = this;

    const i18n = new VueI18n({
        locale: window.CURRENT_LOCALE,
        fallbackLocale: 'en',
        messages: window.i18nMessages,
    });

    this.vue = new Vue({
        el: '#app',
        i18n: i18n,
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
                yourPlayerIndex: null
            },
            gameStateInfo: {
                handsSizes: [],
                deckSize: 0,
                discardPileSize: 0,
                trumpCard: null,
                trumpCardIsInDeck: false,
                trumpCardIsOwnedByPlayerIndex: -1,
                trumpSuit: null,
                attackerIndex: -1,
                defenderIndex: -1,
                yourHand: [],
                canYouPickUp: false,
                canYouComplete: false,
                battleground: [],
                defendingCards: {}, // {1: {suit, value} }
                completedPlayers: {}, // {0: true, 1: false}
                defenderPickUp: false
            },
            gameState: {
                firstAttackerReasonCard: null,
                pickedCard: null,
                gameEnd: false,
                loserIndex: -1
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
            },
            useCard: function (card) {
                if (app.vue.gameState.pickedCard == card) {
                    app.vue.gameState.pickedCard = null
                } else {
                    app.vue.gameState.pickedCard = card;
                }
            },
            attack: () => {
                if (!app.vue.areYouAttacker || !app.vue.gameState.pickedCard) {
                    return;
                }
                app.commandAttack(app.vue.gameState.pickedCard.value, app.vue.gameState.pickedCard.suit);
                app.vue.gameState.pickedCard = null;
            },
            defend: (attackingCard) => {
                if (!app.vue.areYouDefender || !app.vue.gameState.pickedCard) {
                    return;
                }
                app.commandDefend(
                    attackingCard.value,
                    attackingCard.suit,
                    app.vue.gameState.pickedCard.value,
                    app.vue.gameState.pickedCard.suit,
                );
                app.vue.gameState.pickedCard = null;
            },
            pickUp: () => {
                app.commandPickUp();
            },
            complete: () => {
                app.commandComplete();
            },
        },
        computed: {
            areYouAttacker: () => {
                return app.vue.gameStateInfo.attackerIndex === app.vue.game.yourPlayerIndex;
            },
            areYouDefender: () => {
                return app.vue.gameStateInfo.defenderIndex === app.vue.game.yourPlayerIndex;
            },
            areBeaten: () => {
                return app.vue.gameStateInfo.canYouComplete && app.vue.gameStateInfo.canYouPickUp;
            },
            isWaitingForOthers: () => {
                return app.vue.gameStateInfo.battleground.length
                    && !app.vue.gameStateInfo.canYouComplete
                    && !app.vue.gameStateInfo.canYouPickUp;
            },
            attackerNickname: () => {
                if (app.vue.gameStateInfo.attackerIndex < 0) {
                    return;
                }
                const atInd = app.vue.gameStateInfo.attackerIndex;
                if (!app.vue.game.players[atInd]) {
                    return;
                }
                return app.vue.game.players[atInd].name;
            },
            loserNickname: () => {
                if (app.vue.gameState.loserIndex < 0) {
                    return;
                }
                const index = app.vue.gameState.loserIndex;
                if (!app.vue.game.players[index]) {
                    return;
                }
                return app.vue.game.players[index].name;
            },
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
        app.vue.clientsInfo.yourId = data.yourId;
        app.vue.clientsInfo.yourNickname = data.yourNickname;
        app.vue.clientsInfo.clients = data.clients;
        app.vue.rooms = data.rooms;
        if (data.rooms.length === 0) {
            // auto create
            app.commandCreateRoom();
        } else if (data.rooms.length === 1 && data.rooms[0].membersNum === 1) {
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
    };

    this.onRoomInListUpdatedEvent = function (data) {
        const index = app.getRoomIndexById(data.room.id);
        if (index > -1) {
            Vue.set(app.vue.rooms, index, data.room);
        }
    };

    this.onRoomInListRemovedEvent = function (data) {
        const index = app.getRoomIndexById(data.room_id);
        if (index > -1) {
            app.vue.rooms.splice(index, 1);
        } else {
            console.warn("Can't remove room", data.room_id);
        }
    };

    this.onRoomJoinedEvent = function (data) {
        app.vue.room = data.room;
        app.updatePlayersInRoomCounter();
        // TODO: remove debug
        app.commandStartGame();
    };

    this.onRoomUpdatedEvent = function (data) {
        app.vue.room = data.room;
        app.updatePlayersInRoomCounter();
    };

    this.onClientCommandError = function (data) {
        app.vue.commandError = data;
        console.error(data.message);
        window.setTimeout(() => {
            app.vue.commandError = {};
        }, 3000)
    };

    this.onClientCreatedRoomEvent = function (data) {
        app.vue.rooms.push(data.room);
        if (data.room.ownerId === app.vue.clientsInfo.yourId) {
            app.vue.clientsInfo.yourRoomId = data.room.id;
        }
    };

    this.onRoomMemberChangedStatusEvent = function (data) {
        if (data.member.id === app.vue.clientsInfo.yourId) {
            app.vue.wantToPlay = data.member.wantToPlay;
        }
        const memberIndex = app.getRoomMemberIndexById(data.member.id);
        if (memberIndex > -1) {
            Vue.set(app.vue.room.members, memberIndex, data.member);
        }
    };

    this.onRoomMemberChangedPlayerStatusEvent = function (data) {
        const memberIndex = app.getRoomMemberIndexById(data.member.id);
        if (memberIndex > -1) {
            Vue.set(app.vue.room.members, memberIndex, data.member);
        }
        app.updatePlayersInRoomCounter();
    };

    this.onGamePlayersEvent = function (data) {
        app.vue.game.players = data.players;
        app.vue.game.yourPlayerIndex = data.yourPlayerIndex;
    };

    this.updategameStateInfo = (gameStateInfoData) => {
        for (let property in gameStateInfoData) {
            if (gameStateInfoData.hasOwnProperty(property)) {
                const camelizedProperty = this.camelize(property);
                console.log("set gameStateInfoData", camelizedProperty, gameStateInfoData[property]);
                app.vue.gameStateInfo[camelizedProperty] = gameStateInfoData[property];
            }
        }
    };

    this.onGameDealEvent = function (data) {
        app.vue.gameState.gameEnd = false;
        if (data.gameStateInfo) {
            app.updategameStateInfo(data.gameStateInfo);
        }
        // TODO: refactor
        for (let property in data) {
            if (data.hasOwnProperty(property)) {
                const camelizedProperty = this.camelize(property);
                app.vue.gameStateInfo[camelizedProperty] = data[property];
                console.log("set", camelizedProperty, data[property]);
            }
        }
    };

    this.onGameFirstAttackerEvent = function (data) {
        app.vue.gameStateInfo.attackerIndex = data.attackerIndex;
        app.vue.gameStateInfo.defenderIndex = data.defenderIndex;
        app.vue.gameState.firstAttackerReasonCard = data.reasonCard;
    };

    this.onGameAttackEvent = (data) => {
        if (data.gameStateInfo) {
            app.updategameStateInfo(data.gameStateInfo);
        }
        console.log('attack', data);
    };

    this.onGameDefendEvent = (data) => {
        if (data.gameStateInfo) {
            app.updategameStateInfo(data.gameStateInfo);
        }
        console.log('defend', data);
    };

    this.onGameStateEvent = (data) => {
        if (data.gameStateInfo) {
            app.updategameStateInfo(data.gameStateInfo);
            app.vue.gameState.firstAttackerReasonCard = null;
        }
        console.log('state only', data);
    };
    this.onGameEndEvent = (data) => {
        app.vue.gameState.gameEnd = true;
        app.vue.gameState.loserIndex = data.loserIndex;
    };

    this.sendCommand = function (type, subType, data) {
        console.log("send", type, subType, data);
        window.WsConnection.send(JSON.stringify({type: type, sub_type: subType, data: data}));
    };

    this.commandJoinRoom = function (roomId) {
        app.sendCommand('lobby', 'joinRoom', roomId);
    };

    this.commandCreateRoom = function (roomId) {
        app.sendCommand('lobby', 'createRoom', null);
    };

    this.commandWantToPlay = function () {
        app.sendCommand('room', 'wantToPlay', null);
    };

    this.commandWantToSpectate = function () {
        app.sendCommand('room', 'wantToSpectate', null);
    };

    this.commandSetPlayerStatus = function (memberId, status) {
        app.sendCommand('room', 'setPlayerStatus', {memberId: memberId, status: status});
    };

    this.commandStartGame = function () {
        app.sendCommand('room', 'startGame', null);
    };

    this.commandAttack = (value, suit) => {
        app.sendCommand('game', 'attack', {card: {value, suit}});
    };

    this.commandDefend = (attackingValue, attackingSuit, defendingValue, defendingSuit) => {
        const attackingCard = {value: attackingValue, suit: attackingSuit};
        const defendingCard = {value: defendingValue, suit: defendingSuit};
        app.sendCommand('game', 'defend', { attackingCard, defendingCard });
    };

    this.commandPickUp = () => {
        app.sendCommand('game', 'pick_up');
    };

    this.commandComplete = () => {
        app.sendCommand('game', 'complete');
    };

    this.getRoomIndexById = function (roomId) {
        for (let i = 0; i < app.vue.rooms.length; i++) {
            if (app.vue.rooms[i].id === roomId) {
                return i;
            }
        }
        return -1;
    };

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
    };

    this.updatePlayersInRoomCounter = function () {
        let playersNum = 0;
        for (let i = 0; i < app.vue.room.members.length; i++) {
            if (app.vue.room.members[i].isPlayer) {
                playersNum++;
            }
        }
        app.vue.playersInRoom = playersNum;
    };

    this.camelize = function (str) {
        return str.replace(/(_)(.)/g, function ($1, $2, $3) {
            return $3.toUpperCase();
        })
    }
}

(function () {
    const app = new App();
    window.OnIncomingMessage = app.onMessage;
})();
