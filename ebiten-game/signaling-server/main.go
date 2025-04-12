package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"github.com/pion/webrtc/v4"
	"github.com/rs/cors"
)

type ClientConnection struct {
	IsHost bool
	Offer *webrtc.SessionDescription
	Answer *webrtc.SessionDescription
}

type Lobby struct {
	mutex sync.Mutex
	// host is first client in lobby.Clients
	Clients []ClientConnection
}

var lobby_list = map[string]*Lobby{}

type PlayerData struct {
	// player id is index in lobby.Clients
	Id int
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateNewLobbyId() string {
	// have random size for lobby id
	size := 6
    buffer := make([]rune, size)
    for i := range buffer {
        buffer[i] = letters[rand.Intn(len(letters))]
    }
    id := string(buffer)

	// check if room id is already in lobby_list
	_, ok := lobby_list[id]
	if ok {
		// if it already exists, call function again
		return generateNewLobbyId()
	}
	return id
}

func makeLobby() string {
	lobby := Lobby{}
	lobby.Clients = []ClientConnection{}
	// first client is always host
	lobby_id := generateNewLobbyId()
	lobby_list[lobby_id] = &lobby
	return lobby_id
}

func getLobbyIds() []string {
	lobbies := make([]string, len(lobby_list))
	i := 0
	for k := range lobby_list {
		lobbies[i] = k
		i++
	}
	return lobbies
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./public")))
	mux.HandleFunc("/lobby/host", lobbyHost)
	mux.HandleFunc("/lobby/join", lobbyJoin)
	mux.HandleFunc("/lobby/delete", lobbyDelete)
	mux.HandleFunc("/lobby/unregisteredPlayers", lobbyUnregisteredPlayers)
	mux.HandleFunc("/offer/get", offerGet)
	mux.HandleFunc("/offer/post", offerPost)
	mux.HandleFunc("/answer/get", answerGet)
	mux.HandleFunc("/answer/post", answerPost)
	mux.HandleFunc("/ice", ice)

	fmt.Println("Server started on port 3000")
	// cors.Default() setup the middleware with default options being
    // all origins accepted with simple methods (GET, POST). See
    // documentation below for more options.
    handler := cors.Default().Handler(mux)
    http.ListenAndServe(":3000", handler)
}

func lobbyHost(w http.ResponseWriter, r *http.Request) {
	lobby_id := makeLobby()
	lobby := lobby_list[lobby_id]
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()
	// host is first client in lobby.Clients
	lobby.Clients = append(lobby.Clients, ClientConnection{IsHost: true})
	// return lobby id to host
	io.Writer.Write(w, []byte(lobby_id))
	fmt.Println("lobbyHost")
	fmt.Printf("lobby added: %s\n",lobby_id)
	// print all lobbies
	fmt.Printf("lobby_list:%s\n", getLobbyIds())
}

// call "/lobby?id={lobby_id}" to connect to lobby
func lobbyJoin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("lobbyJoin")
	w.Header().Set("Content-Type", "application/json")
	// https://freshman.tech/snippets/go/extract-url-query-params/
	// get lobby id from query params
	lobby_id := r.URL.Query().Get("id")
	fmt.Printf("lobby_id: %s\n", lobby_id)

	// only continue with connection if lobby exists
	lobby, ok := lobby_list[lobby_id]
	// If the key doesn't exist, return error
	if !ok {
    	w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Lobby not found"))
		return
	}
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Failed to read body: %s", err)
		return
	}

	fmt.Printf("body: %s", body)

	// send player id once generated
	lobby.Clients = append(lobby.Clients, ClientConnection{IsHost: false})
	// player id is index in lobby.Clients
	player_id := len(lobby.Clients) - 1
	fmt.Printf("player_id: %d\n", player_id)
	fmt.Println(lobby.Clients)
	player_data := PlayerData{Id: player_id}
	jsonValue, _ := json.Marshal(player_data)
	io.Writer.Write(w, jsonValue)
}

func lobbyDelete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("lobbyDelete")
	w.Header().Set("Content-Type", "application/json")
	// https://freshman.tech/snippets/go/extract-url-query-params/
	// get lobby id from query params
	lobby_id := r.URL.Query().Get("id")
	fmt.Printf("lobby_id: %s\n", lobby_id)
	// delete lobby
	delete(lobby_list, lobby_id)
	fmt.Printf("lobby_list:%s\n", getLobbyIds())
}

// return players who haven't been registered yet by the host
func lobbyUnregisteredPlayers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UnregisteredPlayers")
	w.Header().Set("Content-Type", "application/json")
	// https://freshman.tech/snippets/go/extract-url-query-params/
	// get lobby id from query params
	lobby_id := r.URL.Query().Get("id")
	lobby := lobby_list[lobby_id]
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()
	
	// get all players who haven't been registered yet
	player_ids := []int{}
	for i, client := range lobby.Clients {
		if !client.IsHost && client.Answer == nil {
			player_ids = append(player_ids, i)
		}
	}

	// return lobby id to host
	jsonValue, _ := json.Marshal(player_ids)
	io.Writer.Write(w, jsonValue)
	fmt.Printf("player_ids %v\n", player_ids)
}

func validatePlayer(w http.ResponseWriter, r *http.Request) (*Lobby, int, error) {
	fmt.Println("validatePlayer")
	lobby_id := r.URL.Query().Get("lobby_id")
	//fmt.Printf("lobby_id: %s\n", lobby_id)

	// only continue with connection if lobby exists
	lobby, ok := lobby_list[lobby_id]
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()
	// If the key doesn't exist, return error
	if !ok {
    	w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Lobby not found"))
		return nil, 0, errors.New("Lobby not found")
	}

	player_id_string := r.URL.Query().Get("player_id")
	player_id, err := strconv.Atoi(player_id_string)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Player not found"))
		return nil, 0, errors.New("Player not found")
    }
	//fmt.Printf("player_id: %d\n", player_id)
	//fmt.Printf("length of lobby.Clients: %d\n", len(lobby_list[lobby_id].Clients))
	//fmt.Println(lobby.Clients)
	// check if player actually exists
	if player_id < 0 || player_id >= len(lobby.Clients) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Player not found"))
		return nil, 0, errors.New("Player not found")
	}
	return lobby, player_id, nil
}

func offerGet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("offerGet")
	w.Header().Set("Content-Type", "application/json")

	lobby, player_id, err := validatePlayer(w, r)
	if err != nil {
		return
	}
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	offer := lobby.Clients[player_id].Offer
	if offer == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Offer not found"))
		return
	}

	jsonValue, _ := json.Marshal(offer)

	io.Writer.Write(w, jsonValue)

	/*
	fmt.Println("offerGet")
	fmt.Println(jsonValue)
	*/
}

func offerPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("offerPost")
	
	lobby, player_id, err := validatePlayer(w, r)
	if err != nil {
		return
	}
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	var sdp webrtc.SessionDescription

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&sdp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lobby.Clients[player_id].Offer = &sdp
	fmt.Printf("Lobby: %+v\n", lobby.Clients)
}

func answerGet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("answerGet")
	w.Header().Set("Content-Type", "application/json")

	lobby, player_id, err := validatePlayer(w, r)
	if err != nil {
		return
	}

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	answer := lobby.Clients[player_id].Answer
	if answer == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Answer not found"))
		return
	}

	jsonValue, _ := json.Marshal(answer)

	io.Writer.Write(w, jsonValue)
}

func answerPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("answerPost")
	w.Header().Set("Content-Type", "application/json")

	lobby, player_id, err := validatePlayer(w, r)
	if err != nil {
		return
	}

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	var sdp webrtc.SessionDescription

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&sdp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lobby.Clients[player_id].Answer = &sdp
	fmt.Printf("Lobby: %+v\n", lobby.Clients)
}

func ice(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement the ice handler
	w.Header().Set("Content-Type", "application/json")
}
