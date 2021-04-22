// SPDX-License-Identifier: Apache-2.0
// Copyright 2020,2021 Philipp Naumann, Marcus Soll
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	waitMinutes = 5
	maxWaitTime = 5 * time.Minute
)

var (
	currentGameLock       = sync.Mutex{}
	currentGame     *Game = nil
	upgrader              = websocket.Upgrader{}
	newGameTime           = time.Time{}
)

func init() {
	upgrader.HandshakeTimeout = 3 * time.Second
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	go gameStarterWorker()
}

func endpoint(w http.ResponseWriter, r *http.Request) {
	// Check API key
	key := r.URL.Query().Get("key")
	switch ClaimKey(key) {
	case KeyOK:
		break
	case KeyRateLimit:
		w.WriteHeader(http.StatusTooManyRequests)
		return
	case KeyInvalid:
		w.WriteHeader(http.StatusForbidden)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currentGameLock.Lock()
	if currentGame != nil {
		if currentGame.ContainsAPI(key) {
			currentGameLock.Unlock()
			log.Println("keys (in game):", "ratelimit", key)
			w.WriteHeader(http.StatusTooManyRequests)
			ReleaseKey(key)
			return
		}
	}
	currentGameLock.Unlock()

	log.Printf("connection metadata %s: %s", key, r.Header)

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		ReleaseKey(key)
		return
	}

	if statsEnabled {
		SendLobby <- key
	}

	p := new(Player)
	p.realName = GlobalPseudonym.Get(key)
	p.ws = conn
	p.api = key
	p.Input = make(chan string, 5)
	go p.readWorker()

	// Attach to game
	currentGameLock.Lock()
	defer currentGameLock.Unlock()

	if currentGame == nil {
		currentGame = new(Game)
		newGameTime = time.Now()
	}

	err = currentGame.AddPlayer(p)
	if err == ErrFullGame {
		if currentGame.IsReady() {
			go currentGame.RunGame()
			currentGame = new(Game)
			newGameTime = time.Now()
			if err := currentGame.AddPlayer(p); err != nil {
				log.Println("endpoint:", "add player second time:", err)
				conn.Close()
				return
			}
		} else {
			log.Println("endpoint:", "full game, but not ready")
			conn.Close()
			return
		}
	}

	if currentGame.IsReady() {
		go currentGame.RunGame()
		currentGame = nil
	}
}

func gameStarterWorker() {
	for {
		time.Sleep(1 * time.Second)
		currentGameLock.Lock()
		if currentGame == nil {
			currentGameLock.Unlock()
			continue
		}

		if time.Now().Sub(newGameTime) > maxWaitTime {
			for !currentGame.IsReady() {
				missing := currentGame.MissingPlayer()
				ais := GetAI(missing)
				for i := range ais {
					p := new(Player)
					p.realName = ais[i].API
					p.underlyingAI = ais[i].AI
					p.Input = make(chan string, 5)
					p.underlyingAI.GetChannel(p.Input)
					currentGame.AddPlayer(p)
				}
			}
			go currentGame.RunGame()
			currentGame = nil
		}

		currentGameLock.Unlock()
	}
}
