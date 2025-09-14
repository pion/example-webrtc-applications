// SPDX-FileCopyrightText: 2025 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/pion/webrtc/v4"
	"github.com/rs/cors"
)

type hostConnection struct {
	Offer  *webrtc.SessionDescription
	Answer *webrtc.SessionDescription
}
type clientConnection struct {
	Offer  *webrtc.SessionDescription
	Answer *webrtc.SessionDescription
}

type lobby struct {
	mutex sync.Mutex
	// host is first client in lobby.Clients
	Host    hostConnection
	Clients []clientConnection
}

type lobbyDatabase struct {
	mutex     sync.Mutex
	lobbyList map[string]*lobby
}

type playerData struct {
	// player id is index in lobby.Clients
	ID int
}

var (
	errLobbyNotFound  = errors.New("lobby not found")
	errPlayerNotFound = errors.New("player not found")
)

func (db *lobbyDatabase) generateNewLobbyID() string {
	lobbyList := db.lobbyList
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
		return db.generateNewLobbyID()
	}

	return id
}

func (db *lobbyDatabase) makeLobby() string {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	l := lobby{}
	l.Clients = []clientConnection{}
	// first client is always host
	lobbyID := db.generateNewLobbyID()
	db.lobbyList[lobbyID] = &l

	return lobbyID
}

func (db *lobbyDatabase) deleteLobby(lobbyID string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	delete(db.lobbyList, lobbyID)
}

func (db *lobbyDatabase) getLobbyIDs() []string {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	lobbies := make([]string, len(db.lobbyList))
	i := 0
	for k := range db.lobbyList {
		lobbies[i] = k
		i++
	}

	return lobbies
}

func main() {
	db := lobbyDatabase{
		lobbyList: make(map[string]*lobby),
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./public")))
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("pong"))
		if err != nil {
			fmt.Printf("Failed to write response: %s", err)
		}
	})
	mux.HandleFunc("/host", db.hostHandler)

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

func (db *lobbyDatabase) hostHandler(w http.ResponseWriter, r *http.Request) {
	// create new lobby
	lobbyID := db.makeLobby()
	log.Printf("New lobby created: %s", lobbyID)
	defer db.deleteLobby(lobbyID)

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		println("Failed to accept websocket:", err.Error())
	}
	defer c.CloseNow()

	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		var v any
		err = wsjson.Read(context.Background(), c, &v)
		if err != nil {
			println("Failed to read websocket message:", err.Error())
		}

		log.Printf("received: %v", v)
	}

	c.Close(websocket.StatusNormalClosure, "")
}
