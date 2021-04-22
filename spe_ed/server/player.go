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
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// DirectionUp contains the string value representing "up"
	DirectionUp = "up"
	// DirectionDown contains the string value representing "down"
	DirectionDown = "down"
	// DirectionLeft contains the string value representing "left"
	DirectionLeft = "left"
	// DirectionRight contains the string value representing "right"
	DirectionRight = "right"
)

// Player represents a player of the game.
// It might be a player connected through websocket or an AI.
type Player struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Speed     int    `json:"speed"`
	Active    bool   `json:"active"`
	Name      string `json:"name,omitempty"`

	// To know where wholes need to be
	stepCounter int

	// In case of an AI
	underlyingAI AI

	// Real name - use after game has finished
	realName string

	// API key
	api         string
	apiReleased bool

	// Websocket
	inputLock  sync.Mutex
	Input      chan string `json:"-"` // Must be non-blocking
	writerLock sync.Mutex
	ws         *websocket.Conn
	wsclosed   bool
	workerOnce sync.Once
}

func (p *Player) readWorker() {
	defer func() {
		if p.Input != nil {
			close(p.Input)
			p.Input = nil
		}
	}()

	alreadystarted := true
	p.workerOnce.Do(func() { alreadystarted = false })
	if alreadystarted {
		log.Println("readWorker: already started!")
	}

	for {
		time.Sleep(10 * time.Millisecond)
		p.writerLock.Lock()
		if p.wsclosed {
			p.writerLock.Unlock()
			return
		}
		p.writerLock.Unlock()

		_, b, err := p.ws.ReadMessage()
		if err != nil {
			// Stop on error - something went wrong
			p.writerLock.Lock()
			if !p.wsclosed {
				// Ok, it is not just closed
				log.Println("player read error:", p.api, "-", err)
			}
			if websocket.IsUnexpectedCloseError(err) {
				p.wsclosed = true
				go p.ReleaseAPI()
			}
			p.writerLock.Unlock()
			return
		}

		var a Action
		err = json.Unmarshal(b, &a)
		if err != nil {
			// Stop on error - something went wrong
			p.writerLock.Lock()
			if !p.wsclosed {
				// Ok, it is not just closed
				log.Println("player json error:", p.api, "-", err, "-", string(b))
			}
			p.writerLock.Unlock()
			return
		}
		if p.Input != nil {
			// Don't block
			select {
			case p.Input <- a.Action:
				// Ok
			default:
			}
		}
	}
}

// WriteState sends the given state to the player, either to the websocket or by calling the corresponding AI function.
func (p *Player) WriteState(g *Game) error {
	p.writerLock.Lock()
	defer p.writerLock.Unlock()

	if p.underlyingAI != nil {
		// Pass copy
		go p.underlyingAI.GetState(g.PublicCopy())
		return nil
	}

	if p.ws == nil {
		return nil
	}

	var err error

	if !p.wsclosed {
		err = p.ws.WriteJSON(g)
		if err != nil {
			// is closed - remove
			p.wsclosed = true
			go p.ReleaseAPI()
			return nil
		}
	}
	return err
}

// RevealName will make the pseudonym visible to everyone.
func (p *Player) RevealName() {
	p.writerLock.Lock()
	defer p.writerLock.Unlock()
	p.Name = p.realName
}

// Close will close the player, releasing all ressources (like the websocket).
// It will call ReleaseAPI autmatically
func (p *Player) Close() error {
	p.writerLock.Lock()
	defer p.writerLock.Unlock()

	if p.underlyingAI != nil {
		p.underlyingAI = nil
	}

	if p.ws == nil {
		return nil
	}

	err := p.ws.Close()
	p.wsclosed = true
	go p.ReleaseAPI()
	p.ws = nil
	return err
}

// ReleaseAPI will release the API key of the player.
// Multiple calls will have no effect after the first one.
func (p *Player) ReleaseAPI() {
	p.writerLock.Lock()
	defer p.writerLock.Unlock()

	if !p.apiReleased && p.api != "" {
		p.apiReleased = true
		ReleaseKey(p.api)
	}
}
