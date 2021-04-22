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
	err := RegisterAI("SuperSnailAI", func() AI { return new(SuperSnailAI) })
	if err != nil {
		panic(err)
	}
}

type supersnailAIRevert struct {
	X, Y, Speed, stepCounter int
	Direction                string
	Cells                    []struct{ X, Y int }
}

// SuperSnailAI is an AI that tries to maximise space usage by always 'holding one hand to the wall'. It will usually perform a snail-like pattern at the beginning, thus the name.
// This is an improved version of the SnailAI with a simple dead end prevention.
type SuperSnailAI struct {
	l         sync.Mutex
	i         chan string
	direction string
	round     int
}

// GetChannel receives the answer channel.
func (s *SuperSnailAI) GetChannel(c chan string) {
	s.l.Lock()
	defer s.l.Unlock()

	s.i = c
}

// GetState gets the game state and computes an answer.
func (s *SuperSnailAI) GetState(g *Game) {
	s.l.Lock()
	defer s.l.Unlock()

	if s.i == nil {
		return
	}

	if s.direction == "" {
		if rand.Float32() < 0.5 {
			s.direction = DirectionLeft
		} else {
			s.direction = DirectionRight
		}
	}

	if g.Running {
		snailaction := s.getSnailAction(g)
		if snailaction != "" {
			revert := make([]supersnailAIRevert, 0)
			_, r := s.progress(g, g.You, snailaction)
			revert = append(revert, r)
			if !s.isInSmallArea(g) {
				// Everything ok
				select {
				case s.i <- snailaction:
				default:
				}
				return
			}

			// Try to find a better action
			action := []string{ActionNOOP, ActionTurnLeft, ActionTurnRight}
			test := 0

			for a := range action {
				s.revertStack(g, g.You, revert)
				revert = revert[:0]
				newTest := 1
				alive, r := s.progress(g, g.You, action[a])
				revert = append(revert, r)
				if !alive {
					continue
				}
				for {
					alive, r = s.progress(g, g.You, s.getSnailAction(g))
					revert = append(revert, r)
					if !alive {
						break
					}
					newTest++
				}
				if newTest > test {
					test = newTest
					snailaction = action[a]
				}
			}
			select {
			case s.i <- snailaction:
			default:
			}
			return
		}
		select {
		case s.i <- ActionNOOP:
		default:
		}
	}
}

func (s *SuperSnailAI) getSnailAction(g *Game) string {
	if s.direction == DirectionLeft {
		var nextX, nextY int
		switch g.Players[g.You].Direction {
		case DirectionUp:
			nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
		case DirectionDown:
			nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
		case DirectionLeft:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
		case DirectionRight:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
		}
		if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
			return ActionTurnRight
		}

		switch g.Players[g.You].Direction {
		case DirectionUp:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
		case DirectionDown:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
		case DirectionLeft:
			nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
		case DirectionRight:
			nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
		}
		if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
			return ActionNOOP
		}

		switch g.Players[g.You].Direction {
		case DirectionUp:
			nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
		case DirectionDown:
			nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
		case DirectionLeft:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
		case DirectionRight:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
		}
		if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
			return ActionTurnLeft
		}
	}
	if s.direction == DirectionRight {
		var nextX, nextY int
		switch g.Players[g.You].Direction {
		case DirectionUp:
			nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
		case DirectionDown:
			nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
		case DirectionLeft:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
		case DirectionRight:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
		}
		if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
			return ActionTurnLeft
		}

		switch g.Players[g.You].Direction {
		case DirectionUp:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
		case DirectionDown:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
		case DirectionLeft:
			nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
		case DirectionRight:
			nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
		}
		if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
			return ActionNOOP
		}

		switch g.Players[g.You].Direction {
		case DirectionUp:
			nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
		case DirectionDown:
			nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
		case DirectionLeft:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
		case DirectionRight:
			nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
		}
		if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
			return ActionTurnRight
		}
	}
	return ""
}

// Name returns the name of the AI.
func (s *SuperSnailAI) Name() string {
	return "SuperSnailAI"
}

// progress will progress the game by one step and return the result.
// Not safe for concurrent use on the same game.
func (s *SuperSnailAI) progress(g *Game, player int, command string) (bool, supersnailAIRevert) {
	p := g.Players[player]
	r := supersnailAIRevert{
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
	case "":
		// Abort
		return false, r
	default:
		log.Println("snail ai:", "unknown action", command)
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
func (s *SuperSnailAI) revert(g *Game, player int, r supersnailAIRevert) {
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

func (s *SuperSnailAI) revertStack(g *Game, player int, rs []supersnailAIRevert) {
	for i := len(rs) - 1; i >= 0; i-- {
		s.revert(g, player, rs[i])
	}
}

func (s *SuperSnailAI) isInSmallArea(g *Game) bool {
	test := []struct{ X, Y int }{struct {
		X int
		Y int
	}{g.Players[g.You].X + 1, g.Players[g.You].Y}, struct {
		X int
		Y int
	}{g.Players[g.You].X - 1, g.Players[g.You].Y}, struct {
		X int
		Y int
	}{g.Players[g.You].X, g.Players[g.You].Y + 1}, struct {
		X int
		Y int
	}{g.Players[g.You].X, g.Players[g.You].Y - 1}}
	count := 0
	for i := range test {
		if test[i].X < 0 || test[i].X >= g.Width || test[i].Y < 0 || test[i].Y >= g.Height {
			count++
			continue
		}
		if g.Cells[test[i].Y][test[i].X] == 0 {
			count++
		}
	}
	return count <= 1
}
