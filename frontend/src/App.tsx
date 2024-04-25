import './App.css';

// starts from 0

var player_id: string = "id" + Math.random().toString(16).slice(2)

class SetupLobbyMessage {
    Flag: string;
    PlayerID: string;

    constructor(flag: string, playerid: string) {
        this.Flag = flag;
        this.PlayerID = playerid;
    }
}

class JoinLobbyMessage {
    Flag: string;
    PlayerID: string;
    LobbyID: string;

    constructor(flag: string, playerid: string, lobbyid: string) {
        this.Flag = flag
        this.PlayerID = playerid
        this.LobbyID = lobbyid
    }
}

class StartLobbyMessage {
    Flag:       string;
    LobbyID:    string;

    constructor(flag: string, lobbyid: string) {
        this.Flag = flag
        this.LobbyID = lobbyid
    }
}

var lobby_id : number = 0;

function handleResponse(response: any) {
    if (response["Flag"] == "OK") {
        console.log("Ite!")
        return
    }
    
    if (response["Flag"] == "NewLobby") {
        lobby_id = response["LobbyID"]
    }

    if (response["Flag"] == "ResponseNewRound") {
        console.log("We got a new round request");
        return
    }
    if (response["Flag"] == "ResponseScore") {
        console.log("We got a response score");
        return
    }
}

// take lobby id that user entered and send it to the websocket connection


function App() {
	let socket = new WebSocket("ws://localhost:8080/ws")
    
    let sendLobbyID = function(lobbyid: string) {
        // construct joinLobbyMessage
        let joinLobbyMessage = new JoinLobbyMessage("JoinLobby", player_id, lobbyid)
        socket.send(JSON.stringify(joinLobbyMessage))
    }

    let setupLobby = function() {
        let setupLobbyMessage = new SetupLobbyMessage("SetupLobby", player_id)
        socket.send(JSON.stringify(setupLobbyMessage))
    }

    let startLobby = function() {
        let startLobbyMessage = new StartLobbyMessage("StartLobby", lobby_id.toString())
        socket.send(JSON.stringify(startLobbyMessage))
    }

	console.log("Attempting websocket connection")

    socket.onopen = () => {
        console.log("Successfully established a connection!")
        // construct message
    }

    socket.onclose = () => {
        console.log("Socket closed connection")
    }

    socket.onerror = () => {
        console.log("Error!")
    }

    socket.onmessage = (message) => {
        const response = JSON.parse(message.data);
        console.log("Recieved")
        console.log(response)
        handleResponse(response)
    }

    return (
		<div className="App">
			<header className="App-header">
				<p>
					Edit <code>src/App.js</code> and save to reload lmao.
				</p>
				<a className="App-link" href="https://reactjs.org" target="_blank" rel="noopener noreferrer">
					Learn React
				</a>
                <button
                    onClick={() => setupLobby()}
                >
                    <p> Hello </p>
                </button>
                <button
                    onClick={() => startLobby()}
                >
                <p> start lobby </p>
                </button>
                <input
                    type="text"
                    placeholder='Type LobbyID'
                    onChange={(e) => sendLobbyID(e.target.value)}
                />
            </header>
		</div>
	);
}

export default App;
