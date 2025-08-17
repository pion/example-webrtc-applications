// SPDX-FileCopyrightText: 2025 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

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

type clientConnection struct {
	IsHost bool
	Offer  *webrtc.SessionDescription
	Answer *webrtc.SessionDescription
}

type lobby struct {
	mutex sync.Mutex
	// host is first client in lobby.Clients
	Clients []clientConnection
}

var lobbyList = map[string]*lobby{}

type playerData struct {
	// player id is index in lobby.Clients
	Id int
}

var (
	errLobbyNotFound  = errors.New("lobby not found")
	errPlayerNotFound = errors.New("player not found")
)

func generateNewLobbyID() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// have random size for lobby id
	size := 6
	buffer := make([]rune, size)
	for i := range buffer {
		buffer[i] = letters[rand.Intn(len(letters))] //nolint:gosec
	}
	id := string(buffer)

	// check if room id is already in lobby_list
	_, ok := lobbyList[id]
	if ok {
		// if it already exists, call function again
		return generateNewLobbyID()
	}

	return id
}

func makeLobby() string {
	lobby := lobby{}
	lobby.Clients = []clientConnection{}
	// first client is always host
	lobbyID := generateNewLobbyID()
	lobbyList[lobbyID] = &lobby

	return lobbyID
}

func getLobbyIDs() []string {
	lobbies := make([]string, len(lobbyList))
	i := 0
	for k := range lobbyList {
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
	err := http.ListenAndServe(":3000", handler) //nolint:gosec
	if err != nil {
		fmt.Printf("Failed to start server: %s", err)

		return
	}
}

func lobbyHost(res http.ResponseWriter, _ *http.Request) {
	lobbyID := makeLobby()
	lobby := lobbyList[lobbyID]
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()
	// host is first client in lobby.Clients
	lobby.Clients = append(lobby.Clients, clientConnection{IsHost: true})
	// return lobby id to host
	_, err := io.Writer.Write(res, []byte(lobbyID))
	if err != nil {
		fmt.Printf("Failed to write lobby_id: %s", err)

		return
	}
	fmt.Println("lobbyHost")
	fmt.Printf("lobby added: %s\n", lobbyID)
	// print all lobbies
	fmt.Printf("lobby_list:%s\n", getLobbyIDs())
}

// call "/lobby?id={lobby_id}" to connect to lobby.
func lobbyJoin(res http.ResponseWriter, req *http.Request) {
	fmt.Println("lobbyJoin")
	res.Header().Set("Content-Type", "application/json")
	// https://freshman.tech/snippets/go/extract-url-query-params/
	// get lobby id from query params
	lobbyID := req.URL.Query().Get("id")
	fmt.Printf("lobby_id: %s\n", lobbyID)

	// only continue with connection if lobby exists
	lobby, ok := lobbyList[lobbyID]
	// If the key doesn't exist, return error
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		_, err := res.Write([]byte("404 - Lobby not found"))
		if err != nil {
			fmt.Printf("Failed to write lobby_not_found: %s", err)

			return
		}

		return
	}
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("Failed to read body: %s", err)

		return
	}

	fmt.Printf("body: %s", body)

	// send player id once generated
	lobby.Clients = append(lobby.Clients, clientConnection{IsHost: false})
	// player id is index in lobby.Clients
	playerID := len(lobby.Clients) - 1
	fmt.Printf("player_id: %d\n", playerID)
	fmt.Println(lobby.Clients)
	playerData := playerData{Id: playerID}

	jsonValue, err := json.Marshal(playerData)
	if err != nil {
		fmt.Printf("Failed to marshal player_data: %s", err)

		return
	}

	_, err = io.Writer.Write(res, jsonValue)
	if err != nil {
		fmt.Printf("Failed to write player_data: %s", err)

		return
	}
}

func lobbyDelete(res http.ResponseWriter, req *http.Request) {
	fmt.Println("lobbyDelete")
	res.Header().Set("Content-Type", "application/json")
	// https://freshman.tech/snippets/go/extract-url-query-params/
	// get lobby id from query params
	lobbyID := req.URL.Query().Get("id")
	fmt.Printf("lobby_id: %s\n", lobbyID)
	// delete lobby
	delete(lobbyList, lobbyID)
	fmt.Printf("lobby_list:%s\n", getLobbyIDs())
}

// return players who haven't been registered yet by the host.
func lobbyUnregisteredPlayers(res http.ResponseWriter, req *http.Request) {
	fmt.Println("UnregisteredPlayers")
	res.Header().Set("Content-Type", "application/json")
	// https://freshman.tech/snippets/go/extract-url-query-params/
	// get lobby id from query params
	lobbyID := req.URL.Query().Get("id")
	lobby := lobbyList[lobbyID]
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	// get all players who haven't been registered yet
	playerIDs := []int{}
	for i, client := range lobby.Clients {
		if !client.IsHost && client.Answer == nil {
			playerIDs = append(playerIDs, i)
		}
	}

	// return lobby id to host
	jsonValue, err := json.Marshal(playerIDs)
	if err != nil {
		fmt.Printf("Failed to marshal player_ids: %s", err)

		return
	}

	_, err = io.Writer.Write(res, jsonValue)
	if err != nil {
		fmt.Printf("Failed to write player_ids: %s", err)

		return
	}

	fmt.Printf("player_ids %v\n", playerIDs)
}

func validatePlayer(res http.ResponseWriter, req *http.Request) (*lobby, int, error) {
	fmt.Println("validatePlayer")
	lobbyID := req.URL.Query().Get("lobby_id")

	// only continue with connection if lobby exists
	lobby, ok := lobbyList[lobbyID]
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()
	// If the key doesn't exist, return error
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		_, err := res.Write([]byte("404 - Lobby not found"))
		if err != nil {
			fmt.Printf("Failed to write lobby_not_found: %s", err)

			return nil, 0, errLobbyNotFound
		}

		return nil, 0, errLobbyNotFound
	}

	playerIDString := req.URL.Query().Get("player_id")
	playerID, err := strconv.Atoi(playerIDString)
	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		_, err = res.Write([]byte("404 - Player not found"))
		if err != nil {
			fmt.Printf("Failed to write player_not_found: %s", err)

			return nil, 0, errPlayerNotFound
		}

		return nil, 0, errPlayerNotFound
	}

	// check if player actually exists
	if playerID < 0 || playerID >= len(lobby.Clients) {
		res.WriteHeader(http.StatusNotFound)
		_, err = res.Write([]byte("404 - Player not found"))
		if err != nil {
			fmt.Printf("Failed to write player_not_found: %s", err)

			return nil, 0, errPlayerNotFound
		}

		return nil, 0, errPlayerNotFound
	}

	return lobby, playerID, nil
}

func offerGet(res http.ResponseWriter, req *http.Request) {
	fmt.Println("offerGet")
	res.Header().Set("Content-Type", "application/json")

	lobby, playerID, err := validatePlayer(res, req)
	if err != nil {
		return
	}
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	offer := lobby.Clients[playerID].Offer
	if offer == nil {
		res.WriteHeader(http.StatusNotFound)
		_, err = res.Write([]byte("404 - Offer not found"))
		if err != nil {
			fmt.Printf("Failed to write offer: %s", err)

			return
		}

		return
	}

	jsonValue, err := json.Marshal(offer)
	if err != nil {
		fmt.Printf("Failed to marshal offer: %s", err)

		return
	}

	_, err = io.Writer.Write(res, jsonValue)
	if err != nil {
		fmt.Printf("Failed to write offer: %s", err)

		return
	}
}

func offerPost(res http.ResponseWriter, req *http.Request) {
	fmt.Println("offerPost")

	lobby, playerID, err := validatePlayer(res, req)
	if err != nil {
		return
	}
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	var sdp webrtc.SessionDescription

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(req.Body).Decode(&sdp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)

		return
	}

	lobby.Clients[playerID].Offer = &sdp
	fmt.Printf("Lobby: %+v\n", lobby.Clients)
}

func answerGet(res http.ResponseWriter, req *http.Request) {
	fmt.Println("answerGet")
	res.Header().Set("Content-Type", "application/json")

	lobby, playerID, err := validatePlayer(res, req)
	if err != nil {
		return
	}

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	answer := lobby.Clients[playerID].Answer
	if answer == nil {
		res.WriteHeader(http.StatusNotFound)
		_, err = res.Write([]byte("404 - Answer not found"))
		if err != nil {
			fmt.Printf("Failed to write answer: %s", err)

			return
		}

		return
	}

	jsonValue, err := json.Marshal(answer)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)

		return
	}

	_, err = io.Writer.Write(res, jsonValue)
	if err != nil {
		fmt.Printf("Failed to write answer: %s", err)

		return
	}
}

func answerPost(res http.ResponseWriter, req *http.Request) {
	fmt.Println("answerPost")
	res.Header().Set("Content-Type", "application/json")

	lobby, playerID, err := validatePlayer(res, req)
	if err != nil {
		return
	}

	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	var sdp webrtc.SessionDescription

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(req.Body).Decode(&sdp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)

		return
	}

	lobby.Clients[playerID].Answer = &sdp
	fmt.Printf("Lobby: %+v\n", lobby.Clients)
}

func ice(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}
