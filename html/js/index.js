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
                clients: []
            },
            commandError: {},
            infoMessage: {},
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
            createRoom: (event) => {
                if (event) {
                    event.preventDefault();
                }
                app.commandCreateRoom();
            },
            joinRoom: (roomId, event) => {
                if (event) {
                    event.preventDefault();
                }
                app.commandJoinRoom(roomId);
            },
            markWantToPlay: () => {
                app.commandWantToPlay();
                app.vue.wantToPlay = true;
            },
            markWantToSpectate: () => {
                app.commandWantToSpectate();
                app.vue.wantToPlay = false;
            },
            setPlayerStatus: (memberId, status) => {
                app.commandSetPlayerStatus(memberId, status);
            },
            startGame: () => {
                app.commandStartGame();
            },
            deleteGame: () => {
                app.commandDeleteGame();
            },
            useCard: (card) => {
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

    this.onMessage = (msg) => {
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

    this.onClientJoinedEvent = (data) => {
        app.vue.clientsInfo.yourId = data.yourId;
        app.vue.clientsInfo.yourNickname = data.yourNickname;
        app.vue.clientsInfo.clients = data.clients;
        app.vue.rooms = data.rooms;

        const roomIdFromHash = app.getFromHash('roomId');

        if (data.rooms.length === 0) {
            // auto create
            app.commandCreateRoom();
        } else if (data.rooms.length === 1 && data.rooms[0].membersNum === 1) {
            // auto join to one
            app.commandJoinRoom(data.rooms[0].id);
        } else if (roomIdFromHash) {
            // auto join by hash
            app.commandJoinRoom(roomIdFromHash);
        }
    };

    this.onClientBroadCastJoinedEvent = (data) => {
        app.vue.clientsInfo.clients.push(data);
    };

    this.onClientLeftEvent = (data) => {
        let clients = app.vue.clientsInfo.clients;
        for (let ind = 0; ind < clients.length; ind++) {
            if (clients[ind].id === data.id) {
                clients.splice(ind, 1);
            }
        }
        app.vue.clientsInfo.clients = clients;
    };

    this.onRoomInListUpdatedEvent = (data) => {
        const index = app.getRoomIndexById(data.room.id);
        if (index > -1) {
            Vue.set(app.vue.rooms, index, data.room);
        }
    };

    this.onRoomInListRemovedEvent = (data) => {
        const index = app.getRoomIndexById(data.roomId);
        if (index > -1) {
            app.vue.rooms.splice(index, 1);
        } else {
            console.warn("Can't remove room", data.roomId);
        }
    };

    this.onRoomJoinedEvent = (data) => {
        app.vue.room = data.room;
        app.updatePlayersInRoomCounter();
        app.updateLocationWithRoomId(data.room.id);
        // TODO: remove debug
        app.commandStartGame();
    };

    this.onRoomUpdatedEvent = (data) => {
        app.vue.room = data.room;
        app.updatePlayersInRoomCounter();
    };

    this.onClientCommandError = (data) => {
        app.vue.commandError = data;
        console.error(data.message);
        window.setTimeout(() => {
            app.vue.commandError = {};
        }, 3000)
    };

    this.onClientCreatedRoomEvent = (data) => {
        app.vue.rooms.push(data.room);
    };

    this.onRoomMemberChangedStatusEvent = (data) => {
        if (data.member.id === app.vue.clientsInfo.yourId) {
            app.vue.wantToPlay = data.member.wantToPlay;
        }
        const memberIndex = app.getRoomMemberIndexById(data.member.id);
        if (memberIndex > -1) {
            Vue.set(app.vue.room.members, memberIndex, data.member);
        }
    };

    this.onRoomMemberChangedPlayerStatusEvent = (data) => {
        const memberIndex = app.getRoomMemberIndexById(data.member.id);
        if (memberIndex > -1) {
            Vue.set(app.vue.room.members, memberIndex, data.member);
        }
        app.updatePlayersInRoomCounter();
    };

    this.onGamePlayersEvent = (data) => {
        app.vue.game.players = data.players;
        app.vue.game.yourPlayerIndex = data.yourPlayerIndex;
    };

    this.updateGameStateInfo = (gameStateInfoData) => {
        for (let property in gameStateInfoData) {
            if (gameStateInfoData.hasOwnProperty(property)) {
                console.log("set gameStateInfoData", property, gameStateInfoData[property]);
                app.vue.gameStateInfo[property] = gameStateInfoData[property];
            }
        }
    };

    this.onGameDealEvent = (data) => {
        app.vue.gameState.gameEnd = false;
        if (data.gameStateInfo) {
            app.updateGameStateInfo(data.gameStateInfo);
        }
        // TODO: refactor
        for (let property in data) {
            if (data.hasOwnProperty(property)) {
                const camelizedProperty = app.camelize(property);
                app.vue.gameStateInfo[camelizedProperty] = data[property];
                console.log("set", camelizedProperty, data[property]);
            }
        }
    };

    this.onGameFirstAttackerEvent = (data) => {
        app.vue.gameStateInfo.attackerIndex = data.attackerIndex;
        app.vue.gameStateInfo.defenderIndex = data.defenderIndex;
        app.vue.gameState.firstAttackerReasonCard = data.reasonCard;
    };

    this.onGameAttackEvent = (data) => {
        if (data.gameStateInfo) {
            app.updateGameStateInfo(data.gameStateInfo);
        }
        console.log('attack', data);
    };

    this.onGameDefendEvent = (data) => {
        if (data.gameStateInfo) {
            app.updateGameStateInfo(data.gameStateInfo);
        }
        console.log('defend', data);
    };

    this.onGameStateEvent = (data) => {
        if (data.gameStateInfo) {
            app.updateGameStateInfo(data.gameStateInfo);
            app.vue.gameState.firstAttackerReasonCard = null;
        }
        console.log('state only', data);
    };

    this.onGameEndEvent = (data) => {
        app.vue.gameState.gameEnd = true;
        app.vue.gameState.loserIndex = data.loserIndex;
    };

    this.onGamePlayerLeftEvent = (data) => {
        let playerName = '';
        if (app.vue.game.players[data.playerIndex]) {
            playerName = app.vue.game.players[data.playerIndex].name;
        }

        if (data.isAfk) {
            app.showInfoMessage('player_left_afk', { playerName });
        } else {
            app.showInfoMessage('player_left', { playerName });
        }
    };

    this.showInfoMessage = (messageId, parameters) => {
        app.vue.infoMessage = { messageId, parameters };
        console.info(messageId);
        window.setTimeout(() => {
            app.vue.infoMessage = {};
        }, 10000);
    };

    this.sendCommand = (type, subType, data) => {
        console.log("send", type, subType, data);
        window.WsConnection.send(JSON.stringify({type: type, subType: subType, data: data}));
    };

    this.commandJoinRoom = (roomId) => {
        app.sendCommand('lobby', 'joinRoom', parseInt(roomId));
    };

    this.commandCreateRoom = () => {
        app.sendCommand('lobby', 'createRoom', null);
    };

    this.commandWantToPlay = () => {
        app.sendCommand('room', 'wantToPlay', null);
    };

    this.commandWantToSpectate = () => {
        app.sendCommand('room', 'wantToSpectate', null);
    };

    this.commandSetPlayerStatus = (memberId, status) => {
        app.sendCommand('room', 'setPlayerStatus', {memberId: memberId, status: status});
    };

    this.commandStartGame = () => {
        app.sendCommand('room', 'startGame', null);
    };

    this.commandDeleteGame = () => {
        app.sendCommand('room', 'deleteGame', null);
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
        app.sendCommand('game', 'pickUp');
    };

    this.commandComplete = () => {
        app.sendCommand('game', 'complete');
    };

    this.getRoomIndexById = (roomId) => {
        for (let i = 0; i < app.vue.rooms.length; i++) {
            if (app.vue.rooms[i].id === roomId) {
                return i;
            }
        }
        return -1;
    };

    this.getRoomMemberIndexById = (memberId) => {
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

    this.updatePlayersInRoomCounter = () => {
        let playersNum = 0;
        for (let i = 0; i < app.vue.room.members.length; i++) {
            if (app.vue.room.members[i].isPlayer) {
                playersNum++;
            }
        }
        app.vue.playersInRoom = playersNum;
    };

    this.camelize = (str) => {
        return str.replace(/(_)(.)/g, ($1, $2, $3) => {
            return $3.toUpperCase();
        });
    };

    this.updateLocationWithRoomId = (roomId) => {
        window.location.hash = `roomId=${roomId}`;
    };

    this.getFromHash = (property) => {
        const valuesString = window.location.hash.substr(1);
        const valuesStringPairs = valuesString.split('&');
        for (let i = 0; i < valuesStringPairs.length; i += 1) {
            const pair = valuesStringPairs[i].split('=');
            if (pair.length === 2 && pair[0] === property) {
                return pair[1];
            }
        }
    };
}

(() => {
    const app = new App();
    window.OnIncomingMessage = app.onMessage;
})();
