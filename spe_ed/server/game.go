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
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	// FieldMaxSize contains the maximum size of the field (both width and height).
	FieldMaxSize = 80
	// FieldMinSize contains the minimum size of the field (both width and height).
	FieldMinSize = 40
	// PlayersPerGame contains the maximum number of players allowed in the game.
	// DO NOT CHANGE if you are not absolutely sure to know what side effects this has.
	PlayersPerGame = 6
	// MaxSpeed holds the maximum speed.
	MaxSpeed = 10
	// HolesEachStep holds after how many steps a hole might occur (if the preconditions are met).
	HolesEachStep = 6
	// RoundTimeoutMin has the minimum time a round has (in seconds).
	RoundTimeoutMin = 5
	// RoundTimeoutMax has the maximum time a round has (in seconds).
	RoundTimeoutMax = 15
	// RoundTimeoutGrace has the time after which an answer is accepted even if the deadline has passed. It is a grace period for players.
	RoundTimeoutGrace = 2
	// HoleSpeed contains the minimum speed needed for a hole.
	HoleSpeed = 3
)

var (
	// ErrFullGame is returned when a player is added despite having a full game.
	ErrFullGame = errors.New("full game")
)

// Game represents a game of speed. See https://github.com/informatiCup/InformatiCup2021/ for a description of the game.
type Game struct {
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Cells    [][]int8        `json:"cells"`
	Players  map[int]*Player `json:"players"`
	You      int             `json:"you"` // only needed for protocol, ignored everywhere else
	Running  bool            `json:"running"`
	Deadline string          `json:"deadline,omitempty"` // RFC3339

	l   sync.Mutex
	log *Logger

	MaxPlayer     int `json:"-"`
	numberPlayer  int
	playerAnswer  []string
	playerChannel []chan string
}

// AddPlayer adds a player to the game. Will return ErrFullGame instead if game is full.
func (g *Game) AddPlayer(p *Player) error {
	g.l.Lock()
	defer g.l.Unlock()

	g.setMaxPlayer()

	if g.numberPlayer == g.MaxPlayer {
		return ErrFullGame
	}

	if g.Players == nil {
		g.Players = make(map[int]*Player)
	}

	g.numberPlayer++
	g.Players[g.numberPlayer] = p

	return nil
}

// IsReady returns if the game is ready to start.
func (g *Game) IsReady() bool {
	g.l.Lock()
	defer g.l.Unlock()

	g.setMaxPlayer()

	return g.numberPlayer == g.MaxPlayer
}

// RunGame will (completely) run a game. You have to make sure that IsReady returns true before calling this method.
// It will return the winning player or -1 for a draw. If errors occur, this value is undefined.
func (g *Game) RunGame() (int, error) {
	g.l.Lock()
	defer g.l.Unlock()

	var err error
	var gameID string
	var statLock sync.Mutex

	g.log, gameID, err = GetLogger()
	log.Println("game:", "starting", gameID)

	if err != nil {
		log.Println("getting logger:", err)
	}
	if g.log != nil {
		defer g.log.Close()
		g.log.LogPlayer(g.Players)
	}

	g.setMaxPlayer()

	// Check player
	if g.numberPlayer != g.MaxPlayer {
		return -100, errors.New("not enough player")
	}

	// Initialise
	//// Initialise board
	g.Width = rand.Intn(FieldMaxSize-FieldMinSize) + FieldMinSize + 1
	g.Height = rand.Intn(FieldMaxSize-FieldMinSize) + FieldMinSize + 1

	g.Cells = make([][]int8, g.Height)
	for i := range g.Cells {
		g.Cells[i] = make([]int8, g.Width)
	}

	//// Initialise players
	// Quadrantenphysik
	quarterSelect := []int{0, 1, 2, 3, 4, 5, 6, 7}
	rand.Shuffle(len(quarterSelect), func(i, j int) { quarterSelect[j], quarterSelect[i] = quarterSelect[i], quarterSelect[j] })

	quarterWidth := g.Width / 4
	quarterHeight := g.Height / 2

	quarterNum := 0
	for i := range g.Players {
		g.Players[i].Speed = 1
		g.Players[i].Active = true
		x := 0
		y := 0
		switch quarterNum {
		case 0:
			x = 0
			y = 0
		case 1:
			x = quarterWidth
			y = 0
		case 2:
			x = quarterWidth * 2
			y = 0
		case 3:
			x = quarterWidth * 3
			y = 0
		case 4:
			x = 0
			y = quarterHeight
		case 5:
			x = quarterWidth
			y = quarterHeight
		case 6:
			x = quarterWidth * 2
			y = quarterHeight
		case 7:
			x = quarterWidth
			y = quarterHeight * 3
		}
		g.Players[i].X = x + rand.Intn(quarterWidth)
		g.Players[i].Y = y + rand.Intn(quarterHeight)

		g.Cells[g.Players[i].Y][g.Players[i].X] = int8(i)

		switch {
		case g.Players[i].X > g.Width/2 && g.Players[i].Y > g.Height/2:
			g.Players[i].Direction = DirectionUp
		case g.Players[i].X <= g.Width/2 && g.Players[i].Y > g.Height/2:
			g.Players[i].Direction = DirectionRight
		case g.Players[i].X > g.Width/2 && g.Players[i].Y <= g.Height/2:
			g.Players[i].Direction = DirectionLeft
		case g.Players[i].X <= g.Width/2 && g.Players[i].Y <= g.Height/2:
			g.Players[i].Direction = DirectionDown
		default:
			// Just give some direction
			g.Players[i].Direction = DirectionUp
		}

		quarterNum++
	}

	//// Initialise game
	g.playerChannel = make([]chan string, PlayersPerGame)
	for i := 1; i <= g.numberPlayer; i++ {
		g.playerChannel[i-1] = g.Players[i].Input // Used for communicating later
	}
	g.Running = true

	// Send stats
	if statsEnabled {
		gs := GameStats{
			Key:     gameID,
			Start:   time.Now(),
			Players: make(map[int]PlayerStats),
		}
		for i := range g.Players {
			ps := PlayerStats{
				Pseudonym: g.Players[i].realName,
			}
			if g.Players[i].underlyingAI != nil {
				// Is bot
				ps.Key = g.Players[i].underlyingAI.Name()
				ps.Bot = true
			} else {
				ps.Key = g.Players[i].api
				ps.Bot = false
				go func() { DeleteLobby <- ps.Key }()
			}
			gs.Players[i] = ps
		}
		statLock.Lock()
		go func() {
			SendStat <- gs
			statLock.Unlock()
		}()
	}

	// Run game

mainGame:
	for { // Loop used for rounds
		timeout := rand.Intn(RoundTimeoutMax-RoundTimeoutMin+1) + RoundTimeoutMin
		deadline := time.Now().Add(time.Duration(timeout) * time.Second).UTC()
		g.Deadline = deadline.Format(time.RFC3339)
		g.sendState()
		deadline = deadline.Add(time.Duration(RoundTimeoutGrace) * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		g.playerAnswer = make([]string, PlayersPerGame)
	innerGame:
		for { // Loop used for input
			select { // IMPORTANT: This has to be changed when the number of player changes
			// 1
			case a, ok := <-g.playerChannel[1-1]:
				player := 1
				if !ok {
					g.invalidatePlayer(player)
				} else if a == "" || g.playerAnswer[player-1] != "" || !IsValidAction(a) {
					log.Printf("Invalid answer from %s (%s)", g.Players[player].api, a)
					g.invalidatePlayer(player)
				} else {
					g.playerAnswer[player-1] = a
				}
				if g.checkEndRound() {
					break innerGame
				}
			// 2
			case a, ok := <-g.playerChannel[2-1]:
				player := 2
				if !ok {
					g.invalidatePlayer(player)
				} else if a == "" || g.playerAnswer[player-1] != "" || !IsValidAction(a) {
					log.Printf("Invalid answer from %s (%s)", g.Players[player].api, a)
					g.invalidatePlayer(player)
				} else {
					g.playerAnswer[player-1] = a
				}
				if g.checkEndRound() {
					break innerGame
				}
			// 3
			case a, ok := <-g.playerChannel[3-1]:
				player := 3
				if !ok {
					g.invalidatePlayer(player)
				} else if a == "" || g.playerAnswer[player-1] != "" || !IsValidAction(a) {
					log.Printf("Invalid answer from %s (%s)", g.Players[player].api, a)
					g.invalidatePlayer(player)
				} else {
					g.playerAnswer[player-1] = a
				}
				if g.checkEndRound() {
					break innerGame
				}
			// 4
			case a, ok := <-g.playerChannel[4-1]:
				player := 4
				if !ok {
					g.invalidatePlayer(player)
				} else if a == "" || g.playerAnswer[player-1] != "" || !IsValidAction(a) {
					log.Printf("Invalid answer from %s (%s)", g.Players[player].api, a)
					g.invalidatePlayer(player)
				} else {
					g.playerAnswer[player-1] = a
				}
				if g.checkEndRound() {
					break innerGame
				}
			// 5
			case a, ok := <-g.playerChannel[5-1]:
				player := 5
				if !ok {
					g.invalidatePlayer(player)
				} else if a == "" || g.playerAnswer[player-1] != "" || !IsValidAction(a) {
					log.Printf("Invalid answer from %s (%s)", g.Players[player].api, a)
					g.invalidatePlayer(player)
				} else {
					g.playerAnswer[player-1] = a
				}
				if g.checkEndRound() {
					break innerGame
				}
			// 6
			case a, ok := <-g.playerChannel[6-1]:
				player := 6
				if !ok {
					g.invalidatePlayer(player)
				} else if a == "" || g.playerAnswer[player-1] != "" || !IsValidAction(a) {
					log.Printf("Invalid answer from %s (%s)", g.Players[player].api, a)
					g.invalidatePlayer(player)
				} else {
					g.playerAnswer[player-1] = a
				}
				if g.checkEndRound() {
					break innerGame
				}
			case <-ctx.Done():
				break innerGame
			}
		}
		cancel()

		// Process Actions
		for i := range g.Players {
			switch g.playerAnswer[i-1] {
			case "":
				g.invalidatePlayer(i)
			case ActionTurnLeft:
				switch g.Players[i].Direction {
				case DirectionLeft:
					g.Players[i].Direction = DirectionDown
				case DirectionRight:
					g.Players[i].Direction = DirectionUp
				case DirectionUp:
					g.Players[i].Direction = DirectionLeft
				case DirectionDown:
					g.Players[i].Direction = DirectionRight
				}
			case ActionTurnRight:
				switch g.Players[i].Direction {
				case DirectionLeft:
					g.Players[i].Direction = DirectionUp
				case DirectionRight:
					g.Players[i].Direction = DirectionDown
				case DirectionUp:
					g.Players[i].Direction = DirectionRight
				case DirectionDown:
					g.Players[i].Direction = DirectionLeft
				}
			case ActionFaster:
				g.Players[i].Speed++
				if g.Players[i].Speed > MaxSpeed {
					g.invalidatePlayer(i)
				}
			case ActionSlower:
				g.Players[i].Speed--
				if g.Players[i].Speed < 1 {
					g.invalidatePlayer(i)
				}
			case ActionNOOP:
				// Do nothing
			default:
				g.invalidatePlayer(i)
			}
		}

		// Do Movement
		for i := range g.Players {
			if !g.Players[i].Active {
				continue
			}
			var dostep func(x, y int) (int, int)
			switch g.Players[i].Direction {
			case DirectionUp:
				dostep = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionDown:
				dostep = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionLeft:
				dostep = func(x, y int) (int, int) { return x - 1, y }
			case DirectionRight:
				dostep = func(x, y int) (int, int) { return x + 1, y }
			}

			g.Players[i].stepCounter++

			for s := 0; s < g.Players[i].Speed; s++ {
				g.Players[i].X, g.Players[i].Y = dostep(g.Players[i].X, g.Players[i].Y)
				if g.Players[i].X < 0 || g.Players[i].X >= g.Width || g.Players[i].Y < 0 || g.Players[i].Y >= g.Height {
					g.invalidatePlayer(i)
					break
				}
				if g.Players[i].Speed >= HoleSpeed && g.Players[i].stepCounter%HolesEachStep == 0 && s != 0 && s != g.Players[i].Speed-1 {
					continue
				}
				if g.Cells[g.Players[i].Y][g.Players[i].X] != 0 {
					g.Cells[g.Players[i].Y][g.Players[i].X] = -1
				} else {
					g.Cells[g.Players[i].Y][g.Players[i].X] = int8(i)
				}
			}
		}

		// Check crash
		for i := range g.Players {
			if !g.Players[i].Active {
				continue
			}
			var dostepback func(x, y int) (int, int)
			switch g.Players[i].Direction {
			case DirectionUp:
				dostepback = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionDown:
				dostepback = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionLeft:
				dostepback = func(x, y int) (int, int) { return x + 1, y }
			case DirectionRight:
				dostepback = func(x, y int) (int, int) { return x - 1, y }
			}

			backX := g.Players[i].X
			backY := g.Players[i].Y
			for s := 0; s < g.Players[i].Speed; s++ {
				if g.Cells[backY][backX] == -1 {
					// Crash - check hole
					if g.Players[i].Speed >= HoleSpeed && g.Players[i].stepCounter%HolesEachStep == 0 && s != 0 && s != g.Players[i].Speed-1 {
						// No crash - is hole
					} else {
						g.invalidatePlayer(i)
						break
					}
				}
				backX, backY = dostepback(backX, backY)
			}
		}

		// Check end game
		if g.checkEndGame() {
			break mainGame
		}
	}
	// Finish game
	g.Running = false

	for i := range g.Players {
		g.Players[i].RevealName()
	}

	g.Deadline = ""
	g.sendState()

	winner := -1
	for i := range g.Players {
		if g.Players[i].Active {
			winner = i
			break
		}
	}

	winnerString := "none"
	if winner != -1 {
		if g.Players[winner].underlyingAI != nil {
			winnerString = fmt.Sprintf("#AI#-%s", g.Players[winner].underlyingAI.Name())
		} else {
			winnerString = fmt.Sprintf("#Player#-%s", g.Players[winner].api)
		}
	}

	for i := range g.Players {
		err := g.Players[i].Close()
		if err != nil {
			log.Println("closing player in game:", err)
		}
	}

	log.Println("game:", "ending", gameID, "- winner", winnerString)

	// Delete stats
	if statsEnabled {
		go func() {
			statLock.Lock()
			DeleteStat <- gameID
			statLock.Unlock()
		}()
	}

	return winner, nil
}

// sendState sends the current state to all players.
// Caller has to lock the game.
func (g *Game) sendState() {
	for i := range g.Players {
		g.You = i
		err := g.Players[i].WriteState(g)
		if err != nil {
			log.Println("game sending state:", err)
		}
	}
	g.You = 0

	if g.log != nil {
		g.log.LogState(g)
	}
}

// checkEndRound checks whether the round has finished (all players have answered or are not active).
// Caller has to lock the game.
func (g *Game) checkEndRound() bool {
	t := len(g.playerChannel)
	if len(g.playerAnswer) < t {
		t = len(g.playerAnswer)
	}
	for i := 0; i < t; i++ {
		if g.playerChannel[i] != nil && g.playerAnswer[i] == "" {
			return false
		}
	}
	return true
}

// checkEndGame checks whether the game has finished (only one or none players are active).
// Caller has to lock the game.
func (g *Game) checkEndGame() bool {
	numberActive := 0
	for i := range g.Players {
		if g.Players[i].Active {
			numberActive++
		}
	}
	return numberActive <= 1
}

// invalidatePlayer removes a player from participating in the game.
// This handles setting the player inactive and removing the option to send actions.
// Caller has to lock the game.
func (g *Game) invalidatePlayer(p int) {
	_, ok := g.Players[p]
	if !ok {
		return
	}
	g.Players[p].writerLock.Lock()
	g.Players[p].Active = false
	g.Players[p].writerLock.Unlock()

	g.playerChannel[p-1] = nil
}

// MissingPlayer returns how many players are missing for a full, ready game.
func (g *Game) MissingPlayer() int {
	g.l.Lock()
	defer g.l.Unlock()
	g.setMaxPlayer()
	r := g.MaxPlayer - g.numberPlayer
	if r < 0 {
		return 0
	}
	return r
}

// setMaxPlayer sets the maximum number of players if it is zero.
// Caller has to lock the game.
func (g *Game) setMaxPlayer() {
	if g.MaxPlayer == 0 {
		g.MaxPlayer = rand.Intn(PlayersPerGame-1) + 2
	}
}

// ContainsAPI returns whether a player with the given API key is already registered in the game.
func (g *Game) ContainsAPI(api string) bool {
	g.l.Lock()
	defer g.l.Unlock()

	for k := range g.Players {
		if g.Players[k].api == api {
			return true
		}
	}
	return false
}

// PublicCopy returns a copy of the game with all private fields set to zero.
// As an exception for AIs, Player.stepCounter is also copied.
func (g *Game) PublicCopy() *Game {
	newG := Game{
		Width:    g.Width,
		Height:   g.Height,
		Cells:    make([][]int8, len(g.Cells)),
		Players:  make(map[int]*Player, len(g.Players)),
		You:      g.You,
		Running:  g.Running,
		Deadline: g.Deadline,
	}

	for i := range g.Cells {
		newG.Cells[i] = make([]int8, len(g.Cells[i]))
		copy(newG.Cells[i], g.Cells[i])
	}

	for k := range g.Players {
		newG.Players[k] = &Player{
			X:           g.Players[k].X,
			Y:           g.Players[k].Y,
			Direction:   g.Players[k].Direction,
			Speed:       g.Players[k].Speed,
			Active:      g.Players[k].Active,
			Name:        g.Players[k].Name,
			stepCounter: g.Players[k].stepCounter,
		}
	}
	return &newG
}
