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
    return []byte{}
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

func parse_command(response *map[string]interface{}) []byte {
    flag, _ := (*response)["Flag"].(string);
    switch flag {
    case "SetupLobby":
        return parse_setuplobby(response)
    case "JoinLobby":
        return parse_joinlobby(response)
    case "StartLobby":
        return parse_startlobby(response)


    /*
    case "SubmitGuess":
        return parse_submitguess(response)
    */

    default:
        return []byte{}
    }
}

func reader(conn *websocket.Conn) {
	// while true
	for {
        i, p, err := conn.ReadMessage()

        if err != nil {
            log.Println(err)
        }

        // decode the json string
        var json_map map[string]interface{}
        json.Unmarshal([]byte(string(p) ), &json_map)
        log.Println(json_map)
        message := parse_command(&json_map)
        for _, val := range lobbies {
            log.Println(*val)
        }
        
        conn.WriteMessage(i, []byte(message))
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {return true}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error")
	}

	log.Println("Client successfully connected!")
	reader(ws)
}

func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Starting backend server...")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
