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
	err := RegisterAI("RandomAISlow", func() AI { return new(RandomAISlow) })
	if err != nil {
		panic(err)
	}
}

// RandomAISlow is a variant of the RandomAI which has always speed 1 (and will thus never send "speed_up").
type RandomAISlow struct {
	l sync.Mutex
	i chan string
}

// GetChannel receives the answer channel.
func (r *RandomAISlow) GetChannel(c chan string) {
	r.l.Lock()
	defer r.l.Unlock()

	r.i = c
}

// GetState gets the game state and computes an answer.
func (r *RandomAISlow) GetState(g *Game) {
	r.l.Lock()
	defer r.l.Unlock()

	if r.i == nil {
		return
	}

	if g.Running {
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

		// actions
		actions := []string{ActionTurnLeft, ActionTurnRight, ActionNOOP, ActionNOOP, ActionNOOP, ActionNOOP, ActionNOOP, ActionNOOP, ActionNOOP, ActionNOOP}
		rand.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })
		fallbackAction := ""

		// test actions
		for i := range actions {
			// do action
			switch actions[i] {
			case ActionTurnLeft:
				switch g.Players[g.You].Direction {
				case DirectionLeft:
					g.Players[g.You].Direction = DirectionDown
				case DirectionRight:
					g.Players[g.You].Direction = DirectionUp
				case DirectionUp:
					g.Players[g.You].Direction = DirectionLeft
				case DirectionDown:
					g.Players[g.You].Direction = DirectionRight
				}
			case ActionTurnRight:
				switch g.Players[g.You].Direction {
				case DirectionLeft:
					g.Players[g.You].Direction = DirectionUp
				case DirectionRight:
					g.Players[g.You].Direction = DirectionDown
				case DirectionUp:
					g.Players[g.You].Direction = DirectionRight
				case DirectionDown:
					g.Players[g.You].Direction = DirectionLeft
				}
			case ActionFaster:
				g.Players[g.You].Speed++
				if g.Players[g.You].Speed > MaxSpeed {
					g.Players[g.You].Speed--
					continue
				}
			case ActionSlower:
				g.Players[g.You].Speed--
				if g.Players[g.You].Speed < 1 {
					g.Players[g.You].Speed++
					continue
				}
			case ActionNOOP:
				// Do nothing
			default:
				log.Println("random ai:", "unknown action", actions[i])
			}

			// test
			switch r.willCrash(g) {
			case randomAINoCrash:
				select {
				case r.i <- actions[i]:
				default:
				}
				return
			case randomAIMaybeCrash:
				fallbackAction = actions[i]
			}

			// undo action
			switch actions[i] {
			case ActionTurnLeft:
				switch g.Players[g.You].Direction {
				case DirectionLeft:
					g.Players[g.You].Direction = DirectionUp
				case DirectionRight:
					g.Players[g.You].Direction = DirectionDown
				case DirectionUp:
					g.Players[g.You].Direction = DirectionRight
				case DirectionDown:
					g.Players[g.You].Direction = DirectionLeft
				}
			case ActionTurnRight:
				switch g.Players[g.You].Direction {
				case DirectionLeft:
					g.Players[g.You].Direction = DirectionDown
				case DirectionRight:
					g.Players[g.You].Direction = DirectionUp
				case DirectionUp:
					g.Players[g.You].Direction = DirectionLeft
				case DirectionDown:
					g.Players[g.You].Direction = DirectionRight
				}
			case ActionFaster:
				g.Players[g.You].Speed--
			case ActionSlower:
				g.Players[g.You].Speed++
			case ActionNOOP:
				// Do nothing
			}
		}

		if fallbackAction != "" {
			select {
			case r.i <- fallbackAction:
			default:
			}
			return
		}
		// no valid actions - pick random
		select {
		case r.i <- actions[0]:
		default:
		}
		return
	}
}

// willCrash computes whether the given game state will result in a (possible) crash.
// The return codes are the same as for RandomAI.
// Not safe for concurrent use on the same game.
func (r *RandomAISlow) willCrash(g *Game) int {
	oldX, oldY := g.Players[g.You].X, g.Players[g.You].Y
	defer func() {
		g.Players[g.You].X, g.Players[g.You].Y = oldX, oldY
	}()

	var dostep func(x, y int) (int, int)
	switch g.Players[g.You].Direction {
	case DirectionUp:
		dostep = func(x, y int) (int, int) { return x, y - 1 }
	case DirectionDown:
		dostep = func(x, y int) (int, int) { return x, y + 1 }
	case DirectionLeft:
		dostep = func(x, y int) (int, int) { return x - 1, y }
	case DirectionRight:
		dostep = func(x, y int) (int, int) { return x + 1, y }
	}

	for s := 0; s < g.Players[g.You].Speed; s++ {
		g.Players[g.You].X, g.Players[g.You].Y = dostep(g.Players[g.You].X, g.Players[g.You].Y)
		if g.Players[g.You].X < 0 || g.Players[g.You].X >= g.Width || g.Players[g.You].Y < 0 || g.Players[g.You].Y >= g.Height {
			return randomAISureCrash
		}
		if g.Players[g.You].Speed >= HoleSpeed && (g.Players[g.You].stepCounter+1)%HolesEachStep == 0 && s != 0 && s != g.Players[g.You].Speed-1 {
			continue
		}
		if g.Cells[g.Players[g.You].Y][g.Players[g.You].X] == -100 {
			return randomAIMaybeCrash
		}
		if g.Cells[g.Players[g.You].Y][g.Players[g.You].X] != 0 {
			return randomAISureCrash
		}
	}

	return randomAINoCrash
}

// Name returns the name of the AI.
func (r *RandomAISlow) Name() string {
	return "RandomAISlow"
}
