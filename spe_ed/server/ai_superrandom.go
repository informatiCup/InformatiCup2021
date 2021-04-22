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
	"math/rand"
	"sync"
)

func init() {
	err := RegisterAI("SuperRandomAI", func() AI { return new(SuperRandomAI) })
	if err != nil {
		panic(err)
	}
}

const (
	superRandomAIPathLength = HolesEachStep * 2
)

type superSnailAIRevert struct {
	X, Y, Speed, stepCounter int
	Direction                string
	Cells                    []struct{ X, Y int }
}

// SuperRandomAI is an improved version of the RandomAI which does a random action with a long possible path.
type SuperRandomAI struct {
	l sync.Mutex

	i chan string
}

// GetChannel receives the answer channel.
func (sr *SuperRandomAI) GetChannel(c chan string) {
	sr.l.Lock()
	defer sr.l.Unlock()

	sr.i = c
}

// GetState gets the game state and computes an answer.
func (sr *SuperRandomAI) GetState(g *Game) {
	sr.l.Lock()
	defer sr.l.Unlock()

	if sr.i == nil {
		return
	}

	if g.Running && g.Players[g.You].Active {
		// Fill potential dead zones
		for k := range g.Players {
			if k == g.You {
				continue
			}

			if !g.Players[k].Active {
				continue
			}

			for i := 1; i <= g.Players[k].Speed+1; i++ {
				x, y := g.Players[k].X+i, g.Players[k].Y
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
					// invalid - do nothing
				} else {
					g.Cells[y][x] = -100
				}

				x, y = g.Players[k].X-i, g.Players[k].Y
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
					// invalid - do nothing
				} else {
					g.Cells[y][x] = -100
				}

				x, y = g.Players[k].X, g.Players[k].Y+i
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
					// invalid - do nothing
				} else {
					g.Cells[y][x] = -100
				}

				x, y = g.Players[k].X, g.Players[k].Y-i
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
					// invalid - do nothing
				} else {
					g.Cells[y][x] = -100
				}
			}
		}

		action := ""
		best := 0

		// Try finding best action
		actions := make([]string, 0, 5)
		actions = append(actions, ActionTurnLeft, ActionTurnRight, ActionNOOP)

		if g.Players[g.You].Speed > 1 {
			actions = append(actions, ActionSlower)
		}
		if g.Players[g.You].Speed < 5 {
			actions = append(actions, ActionFaster)
		}
		rand.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })

		for a := range actions {
			b, r := sr.progress(g, g.You, actions[a])
			if !b {
				sr.revert(g, g.You, r)
				continue
			}
			try := sr.getLength(superRandomAIPathLength, g)
			sr.revert(g, g.You, r)
			if try > best {
				best = try
				action = actions[a]
				if try == superRandomAIPathLength {
					break
				}
			}
		}
		if action == "" {
			// Try finding 1 step - reuse RandomAI
			ai := RandomAI{}
			ai.GetChannel(sr.i)
			ai.GetState(g)
			return
		}

		sr.i <- action
	}
}

// Name returns the name of the AI.
func (sr *SuperRandomAI) Name() string {
	return "SuperRandomAI"
}

// getLength returns the longest possible path from the current game state for Game.You.
// Not safe for concurrent use on the same game.
func (sr *SuperRandomAI) getLength(max int, g *Game) int {
	max--
	if max < 0 {
		return 0
	}
	actions := make([]string, 0, 5)
	actions = append(actions, ActionTurnLeft, ActionTurnRight, ActionNOOP)

	if g.Players[g.You].Speed > 1 {
		actions = append(actions, ActionSlower)
	}
	if g.Players[g.You].Speed < 5 {
		actions = append(actions, ActionFaster)
	}

	found := 0

	for i := range actions {
		result, revert := sr.progress(g, g.You, actions[i])
		if result {
			f := 1 + sr.getLength(max, g)
			if f > found {
				found = f
				if f-1 == max {
					break
				}
			}
		}
		sr.revert(g, g.You, revert)
	}

	return found
}

// progress will progress the game by one step and return the result.
// Not safe for concurrent use on the same game.
func (sr *SuperRandomAI) progress(g *Game, player int, command string) (bool, superSnailAIRevert) {
	p := g.Players[player]
	r := superSnailAIRevert{
		X:           p.X,
		Y:           p.Y,
		Speed:       p.Speed,
		stepCounter: p.stepCounter,
		Direction:   p.Direction,
		Cells:       make([]struct{ X, Y int }, 0, p.Speed),
	}
	switch command {
	case ActionTurnLeft:
		switch p.Direction {
		case DirectionLeft:
			p.Direction = DirectionDown
		case DirectionRight:
			p.Direction = DirectionUp
		case DirectionUp:
			p.Direction = DirectionLeft
		case DirectionDown:
			p.Direction = DirectionRight
		}
	case ActionTurnRight:
		switch p.Direction {
		case DirectionLeft:
			p.Direction = DirectionUp
		case DirectionRight:
			p.Direction = DirectionDown
		case DirectionUp:
			p.Direction = DirectionRight
		case DirectionDown:
			p.Direction = DirectionLeft
		}
	case ActionFaster:
		p.Speed++
		if p.Speed > MaxSpeed {
			return false, r
		}
	case ActionSlower:
		p.Speed--
		if p.Speed < 1 {
			return false, r
		}
	case ActionNOOP:
		// Do nothing
	default:
		log.Println("jump ai:", "unknown action", command)
	}

	var dostep func(x, y int) (int, int)
	switch p.Direction {
	case DirectionUp:
		dostep = func(x, y int) (int, int) { return x, y - 1 }
	case DirectionDown:
		dostep = func(x, y int) (int, int) { return x, y + 1 }
	case DirectionLeft:
		dostep = func(x, y int) (int, int) { return x - 1, y }
	case DirectionRight:
		dostep = func(x, y int) (int, int) { return x + 1, y }
	}

	p.stepCounter++

	for s := 0; s < p.Speed; s++ {
		p.X, p.Y = dostep(p.X, p.Y)
		if p.X < 0 || p.X >= g.Width || p.Y < 0 || p.Y >= g.Height {
			return false, r
		}
		if p.Speed >= HoleSpeed && p.stepCounter%HolesEachStep == 0 && s != 0 && s != p.Speed-1 {
			continue
		}
		if g.Cells[p.Y][p.X] != 0 {
			return false, r
		}
		r.Cells = append(r.Cells, struct{ X, Y int }{p.X, p.Y})
		g.Cells[p.Y][p.X] = -33
	}

	return true, r
}

// revert reverts the game state by the revert struct.
// Not safe for cocurrent use on the same game.
func (sr *SuperRandomAI) revert(g *Game, player int, r superSnailAIRevert) {
	p := g.Players[player]
	p.X = r.X
	p.Y = r.Y
	p.Speed = r.Speed
	p.stepCounter = r.stepCounter
	p.Direction = r.Direction
	for i := range r.Cells {
		g.Cells[r.Cells[i].Y][r.Cells[i].X] = 0
	}
}
