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
	"sync"
)

func init() {
	err := RegisterAI("JumpingLargestFreeAI", func() AI { return new(JumpingLargestFreeAI) })
	if err != nil {
		panic(err)
	}
}

// JumpingLargestFreeAIJumpAtLessThanFree is the number of free cells connected at which the AI tries to jump.
const JumpingLargestFreeAIJumpAtLessThanFree = 50

// JumpingLargestFreeAI behaves like LargestFreeAI but tries to jump out of small areas.
type JumpingLargestFreeAI struct {
	l sync.Mutex

	i                 chan string
	largestfree       AI
	jump              AI
	freeCountingSlice []bool
}

// GetChannel receives the answer channel.
func (jlf *JumpingLargestFreeAI) GetChannel(c chan string) {
	jlf.l.Lock()
	defer jlf.l.Unlock()

	jlf.i = c

	if jlf.largestfree != nil {
		jlf.largestfree.GetChannel(c)
	}

	if jlf.jump != nil {
		jlf.jump.GetChannel(c)
	}
}

// GetState gets the game state and computes an answer.
func (jlf *JumpingLargestFreeAI) GetState(g *Game) {
	jlf.l.Lock()
	defer jlf.l.Unlock()

	if jlf.i == nil {
		return
	}

	if jlf.largestfree == nil {
		jlf.largestfree = new(LargestFreeAI)
		jlf.largestfree.GetChannel(jlf.i)
	}

	if g.Running && g.Players[g.You].Active {
		if jlf.freeSpaceConnected(g.Players[g.You].X, g.Players[g.You].Y, JumpingLargestFreeAIJumpAtLessThanFree+1, g) < JumpingLargestFreeAIJumpAtLessThanFree {
			if jlf.jump == nil {
				jlf.jump = new(JumpAI)
				jlf.jump.GetChannel(jlf.i)
			}
			jlf.jump.GetState(g)
		} else if g.Players[g.You].Speed > 1 {
			// Check slow_down
			var dostep func(x, y int) (int, int)
			possible := true
			x, y := g.Players[g.You].X, g.Players[g.You].Y

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

			for i := 0; i < g.Players[g.You].Speed-1; i++ {
				x, y = dostep(x, y)
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
					possible = false
					break
				}
			}

			if possible {
				select {
				case jlf.i <- ActionSlower:
				default:
				}
				return
			}

			// Check turn_left
			possible = true
			x, y = g.Players[g.You].X, g.Players[g.You].Y

			switch g.Players[g.You].Direction {
			case DirectionLeft:
				dostep = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionRight:
				dostep = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionDown:
				dostep = func(x, y int) (int, int) { return x - 1, y }
			case DirectionUp:
				dostep = func(x, y int) (int, int) { return x + 1, y }
			}

			for i := 0; i < g.Players[g.You].Speed; i++ {
				x, y = dostep(x, y)
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
					possible = false
					break
				}
			}

			if possible {
				select {
				case jlf.i <- ActionTurnLeft:
				default:
				}
				return
			}

			// Check turn_right
			possible = true
			x, y = g.Players[g.You].X, g.Players[g.You].Y

			switch g.Players[g.You].Direction {
			case DirectionRight:
				dostep = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionLeft:
				dostep = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionUp:
				dostep = func(x, y int) (int, int) { return x - 1, y }
			case DirectionDown:
				dostep = func(x, y int) (int, int) { return x + 1, y }
			}

			for i := 0; i < g.Players[g.You].Speed; i++ {
				x, y = dostep(x, y)
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
					possible = false
					break
				}
			}

			if possible {
				select {
				case jlf.i <- ActionTurnRight:
				default:
				}
				return
			}

			// nothing to survive, so....
			select {
			case jlf.i <- ActionSlower:
			default:
			}
		} else {
			jlf.jump = nil
			jlf.largestfree.GetState(g)
		}
	}
}

// Name returns the name of the AI.
func (jlf *JumpingLargestFreeAI) Name() string {
	return "JumpingLargestFreeAI"
}

// freeSpaceConnected calculates the number of free space connected to given area.
// It is not safe for concurrent usage on the same instance of JumpingLargestFreeAI.
func (jlf *JumpingLargestFreeAI) freeSpaceConnected(x, y, cutoff int, g *Game) int {
	// cutoff -1 == no cutoff
	// Not concurrent safe

	if len(jlf.freeCountingSlice) != g.Height*g.Width {
		jlf.freeCountingSlice = make([]bool, g.Height*g.Width)
	} else {
		for i := range jlf.freeCountingSlice {
			jlf.freeCountingSlice[i] = false
		}
	}

	current := 0

	current = jlf.freeSpaceConnectedInternal(x, y, cutoff, current, g)
	current = jlf.freeSpaceConnectedInternal(x-1, y, cutoff, current, g)
	current = jlf.freeSpaceConnectedInternal(x+1, y, cutoff, current, g)
	current = jlf.freeSpaceConnectedInternal(x, y-1, cutoff, current, g)
	current = jlf.freeSpaceConnectedInternal(x, y+1, cutoff, current, g)

	return current
}

func (jlf *JumpingLargestFreeAI) freeSpaceConnectedInternal(x, y, cutoff, current int, g *Game) int {
	if cutoff != -1 && current > cutoff {
		return current
	}

	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return current
	}

	cell := y*g.Width + x

	if jlf.freeCountingSlice[cell] {
		return current
	}
	jlf.freeCountingSlice[cell] = true

	if g.Cells[y][x] != 0 {
		return current
	}
	current++

	current = jlf.freeSpaceConnectedInternal(x-1, y, cutoff, current, g)
	current = jlf.freeSpaceConnectedInternal(x+1, y, cutoff, current, g)
	current = jlf.freeSpaceConnectedInternal(x, y-1, cutoff, current, g)
	current = jlf.freeSpaceConnectedInternal(x, y+1, cutoff, current, g)

	return current
}
