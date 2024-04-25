package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:     1024,
	WriteBufferSize:    1024,
}

type ResponseNewRound struct {
    Flag                string
    SelectedHonestID    string
    SelectedGuesserID   string
    Question            string
    Answer              string
}

type ResponseScore struct {
    Flag    string
    Lost    bool
    Score   bool
}

type ResponseOkay struct {
    Flag    string
}

type ResponseNewLobby struct {
    Flag    string
    LobbyID string
}

type ResponseLobbyUpdate struct {
    Flag    string
    Players []string
}


func (r ResponseNewRound) to_bytes() []byte {
    data, _ := json.Marshal(r)
    return data
}

func (r ResponseScore) to_bytes() []byte {
    data, _ := json.Marshal(r)
    return data
}

func (r ResponseOkay) to_bytes() []byte {
    data, _ := json.Marshal(r)
    return data
}

func (r ResponseNewLobby) to_bytes() []byte {
    data, _ := json.Marshal(r)
    return data
}

func (r ResponseLobbyUpdate) to_bytes() []byte {
    data, _ := json.Marshal(r)
    return data
}

type Response interface  {
    to_bytes() []byte
}

// contains game state info
type Lobby struct {
    has_started         bool
    // player score map, this is initialised for every call to joinlobby
    player_score_map    map[string]int
    current_guesser     string
    current_honest      string
    number_of_rounds    int
}

var lobbies map[string]*Lobby = make(map[string]*Lobby)

func Keys[M ~map[K]V, K comparable, V any](m M) []K {
    r := make([]K, 0, len(m))
    for k := range m {
        r = append(r, k)
    }
    return r
}


func parse_setuplobby(response *map[string]interface{}) []byte {
    // expecting:
    // PlayerID: string
    player_id, _ := (*response)["PlayerID"].(string)
    // generate fresh lobby id
    lobby_id := fmt.Sprintf("%d", rand.Uint32())
    initialize_lobby := Lobby {
        has_started: false,
        player_score_map: make(map[string]int),
        current_guesser: "",
        current_honest: "",
        number_of_rounds: 0,
    }

    // add lobby
    lobbies[lobby_id] = &initialize_lobby
    // initialise player
    (*(lobbies[lobby_id])).player_score_map[player_id] = 0
    return ResponseNewLobby{Flag: "NewLobby", LobbyID: lobby_id}.to_bytes()
}

func parse_joinlobby(response *map[string]interface{}) []byte {
    // expecting:
    // LobbyID: string
    // PlayerID: string
    player_id, _ := (*response)["PlayerID"].(string)
    lobby_id, _ := (*response)["LobbyID"].(string)
    // add player to player score map,
    (*(lobbies[lobby_id])).player_score_map[player_id] = 0
    return ResponseOkay{Flag: "join lobby good"}.to_bytes()
}

func parse_startlobby(response *map[string]interface{}) []byte {
    //  expecting:
    //  LobbyID:    string
    lobby_id, _ := (*response)["LobbyID"].(string)
    lol := lobbies[lobby_id]
    lol.has_started = true
    // return new round response
    
    (*(lobbies[lobby_id])).has_started = true
    // pick new guesser and honest
    players := Keys((*(lobbies[lobby_id])).player_score_map)
    index := (*(lobbies[lobby_id])).number_of_rounds
    guesser := players[index % len(players)]
    honest := players[(len(players) - index) % len(players)]
    // set guesser and honest
    (*(lobbies[lobby_id])).current_guesser = guesser
    (*(lobbies[lobby_id])).current_honest = honest
    (*(lobbies[lobby_id])).number_of_rounds += 1
    question := "Why did the chicken cross the road"
    answer := "To get to the other side"
    reply := ResponseNewRound {
        Flag: "ResponseNewRound",
        SelectedHonestID: honest,
        SelectedGuesserID: guesser,
        Question: question,
        Answer: answer,
    }
    return reply.to_bytes()
}

type ClientReply struct {
    message     []byte
    host_cid    string
}

type ServerReply struct {
    message             []byte
    broadcast_to_all    bool
    dest_cid            string
}

func parse_command(response *map[string]interface{}, curr_cid string) ServerReply {
    flag, _ := (*response)["Flag"].(string);
    switch flag {
    case "SetupLobby":
        return ServerReply{
            message: parse_setuplobby(response),
            broadcast_to_all: false,
            dest_cid: curr_cid,
        }
    case "JoinLobby":
        return ServerReply{
            message: parse_joinlobby(response),
            broadcast_to_all: false,
            dest_cid: curr_cid,
        }
    case "StartLobby":
        return ServerReply{
            message: parse_startlobby(response),
            broadcast_to_all: true,
        }
    default:
        return ServerReply{
            message: []byte{},
            broadcast_to_all: false,
        }
    }
}


var reading_channel = make(chan ClientReply)
var writing_channel = make(chan ServerReply)


// list of connection ids
var cids = []string{}
var connection_number = 0
func connection(conn *websocket.Conn) {
    connection_number += 1
    var connection_id = fmt.Sprintf("%d", connection_number)
    cids = append(cids, connection_id)
    var reading_thread = func() {
        for {
            // Receive a message.
            _, p, err := conn.ReadMessage()
            if err != nil {
                log.Println(err)
            }
            // Send message to channel
            reading_channel <- ClientReply{message: p, host_cid: connection_id}
        }
    }

    var writing_thread = func() {
        for {
            serverReply := <-writing_channel
            if connection_id == serverReply.dest_cid {
                conn.WriteMessage(1, serverReply.message)
            }
        }
    }
    // launch threads
    go reading_thread()
    go writing_thread()
}
var current_connections = 0
func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {return true}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error")
	}
    fmt.Println("making a new connection")
    go connection(ws)
}

func connection_broker() {
    for {
        clientReply := <-reading_channel
        var json_map map[string]interface{}
        json.Unmarshal([]byte(string(clientReply.message)), &json_map)
        log.Println("received")
        log.Println(json_map)
        server_reply := parse_command(&json_map, clientReply.host_cid)
        // if it should be sent to everyone
        fmt.Println("we good ehre")
        if server_reply.broadcast_to_all {
            for cid := range cids {
                fmt.Println("writing")
                writing_channel <- ServerReply {
                message: server_reply.message,
                dest_cid: cids[cid],
                }
            }
        } else {
            writing_channel <- ServerReply {
            message: server_reply.message,
            dest_cid: clientReply.host_cid,
            }
        }
        fmt.Println("onto the other one")
    }
}

func setupRoutes() {
    // start connection broker
    go connection_broker()
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Starting backend server...")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
