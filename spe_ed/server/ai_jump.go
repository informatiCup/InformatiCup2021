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
	err := RegisterAI("JumpAI", func() AI { return new(JumpAI) })
	if err != nil {
		panic(err)
	}
}

const (
	// JumpAITries contains the number of iterations JumpAI uses to find a jump
	JumpAITries = 100
)

const (
	jumpAIprogressNormal = iota
	jumpAIprogressJump
	jumpAIprogressCrash
)

type jumpAIRevert struct {
	X, Y, Speed, stepCounter int
	Direction                string
	Cells                    []struct{ X, Y int }
}

// JumpAI tries to find a possible jump and then tries to execute it if possible. If no jump is found, it behaves like RandomAI.
type JumpAI struct {
	l sync.Mutex

	i    chan string
	plan []string
	r    *rand.Rand
}

// GetChannel receives the answer channel.
func (j *JumpAI) GetChannel(c chan string) {
	j.l.Lock()
	defer j.l.Unlock()

	j.i = c
}

// GetState gets the game state and computes an answer.
func (j *JumpAI) GetState(g *Game) {
	j.l.Lock()
	defer j.l.Unlock()

	if j.i == nil {
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

		if len(j.plan) != 0 {
			if !j.executePlan(g, j.plan) {
				j.plan = nil
			}
		}

		if len(j.plan) == 0 {
			if j.r == nil {
				j.r = rand.New(rand.NewSource(rand.Int63()))
			}

			length := HolesEachStep - (g.Players[g.You].stepCounter % HolesEachStep)

			// Try finding jump
			j.plan = j.findPlan(length, g.PublicCopy())

			if len(j.plan) == 0 {
				// Try finding 1 step - reuse RandomAI
				c := make(chan string, 1)
				ai := RandomAI{}
				ai.GetChannel(c)
				ai.GetState(g)
				j.plan = []string{<-c}
			}
		}
		action := j.plan[0]
		j.plan = j.plan[1:]
		j.i <- action
	}
}

// Name returns the name of the AI.
func (j *JumpAI) Name() string {
	return "JumpAI"
}

// findPlan will try to find a plan containing a jump with a maximum of length steps.
// Function will return nil if no plan is found.
// Not safe for concurrent use.
func (j *JumpAI) findPlan(length int, g *Game) []string {
	length--
	if length < 0 {
		return nil
	}
	actions := []string{ActionTurnLeft, ActionTurnRight, ActionSlower, ActionFaster, ActionNOOP}
	j.r.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })

	for i := range actions {
		result, revert := j.progress(g, g.You, actions[i])
		switch result {
		case jumpAIprogressCrash:
			j.revert(g, g.You, revert)
			continue
		case jumpAIprogressNormal:
			plan := j.findPlan(length, g)
			j.revert(g, g.You, revert)
			if plan == nil {
				continue
			}
			plan = append([]string{actions[i]}, plan...)
			return plan
		case jumpAIprogressJump:
			j.revert(g, g.You, revert)
			return []string{actions[i]}
		}
	}

	return nil
}

// progress will progress the game by one step and return the result.
// Not safe for concurrent use on the same game.
func (j *JumpAI) progress(g *Game, player int, command string) (int, jumpAIRevert) {
	p := g.Players[player]
	r := jumpAIRevert{
		X:           p.X,
		Y:           p.Y,
		Speed:       p.Speed,
		stepCounter: p.stepCounter,
		Direction:   p.Direction,
		Cells:       make([]struct{ X, Y int }, 0, p.Speed),
	}
	jump := false
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
			return jumpAIprogressCrash, r
		}
	case ActionSlower:
		p.Speed--
		if p.Speed < 1 {
			return jumpAIprogressCrash, r
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
			return jumpAIprogressCrash, r
		}
		if p.Speed >= HoleSpeed && p.stepCounter%HolesEachStep == 0 && s != 0 && s != p.Speed-1 {
			if g.Cells[p.Y][p.X] != 0 {
				jump = true
			}
			continue
		}
		if g.Cells[p.Y][p.X] != 0 {
			return jumpAIprogressCrash, r
		}
		r.Cells = append(r.Cells, struct{ X, Y int }{p.X, p.Y})
		g.Cells[p.Y][p.X] = -33
	}

	if jump {
		return jumpAIprogressJump, r
	}
	return jumpAIprogressNormal, r
}

// revert reverts the game state by the revert struct.
// Not safe for cocurrent use on the same game.
func (j *JumpAI) revert(g *Game, player int, r jumpAIRevert) {
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

// executePlan returns true if given plan jumps over SOMETHING.
// It is not safe for concurrent usage on the same game, however it will revert the game to the initial state given to the function.
func (j *JumpAI) executePlan(g *Game, plan []string) bool {
	revert := make([]struct{ X, Y int }, 0, 60)
	defer func() {
		// Revert cells
		for i := range revert {
			g.Cells[revert[i].Y][revert[i].X] = 0
		}
	}()

	sc := g.Players[g.You].stepCounter
	direction := g.Players[g.You].Direction
	speed := g.Players[g.You].Speed
	x, y := g.Players[g.You].X, g.Players[g.You].Y

	// Execute plan
	jump := false
	for i := range plan {
		switch plan[i] {
		case ActionTurnLeft:
			switch direction {
			case DirectionLeft:
				direction = DirectionDown
			case DirectionRight:
				direction = DirectionUp
			case DirectionUp:
				direction = DirectionLeft
			case DirectionDown:
				direction = DirectionRight
			}
		case ActionTurnRight:
			switch direction {
			case DirectionLeft:
				direction = DirectionUp
			case DirectionRight:
				direction = DirectionDown
			case DirectionUp:
				direction = DirectionRight
			case DirectionDown:
				direction = DirectionLeft
			}
		case ActionFaster:
			speed++
			if speed > MaxSpeed {
				return false
			}
		case ActionSlower:
			speed--
			if speed < 1 {
				return false
			}
		case ActionNOOP:
			// Do nothing
		default:
			log.Println("jump ai:", "unknown action", plan[i])
		}

		var dostep func(x, y int) (int, int)
		switch direction {
		case DirectionUp:
			dostep = func(x, y int) (int, int) { return x, y - 1 }
		case DirectionDown:
			dostep = func(x, y int) (int, int) { return x, y + 1 }
		case DirectionLeft:
			dostep = func(x, y int) (int, int) { return x - 1, y }
		case DirectionRight:
			dostep = func(x, y int) (int, int) { return x + 1, y }
		}

		sc++

		for s := 0; s < speed; s++ {
			x, y = dostep(x, y)
			if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
				return false
			}
			if speed >= HoleSpeed && sc%HolesEachStep == 0 && s != 0 && s != speed-1 {
				if g.Cells[y][x] != 0 {
					jump = true
				}
				continue
			}
			if g.Cells[y][x] != 0 {
				return false
			}
			g.Cells[y][x] = -33
			revert = append(revert, struct{ X, Y int }{x, y})
		}

	}

	return jump
}
