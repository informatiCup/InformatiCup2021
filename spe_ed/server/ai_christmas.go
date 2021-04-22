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
	err := RegisterAI("ChristmasAI", func() AI { return new(ChristmasAI) })
	if err != nil {
		panic(err)
	}
}

// ChristmasAIActions contains the different forms of the ChristmasAI (in order): Tree, candle, angel, shooting star.
var ChristmasAIActions = []string{
	"CRCL+-RRLRLRLRLLCCRRLRLRLLCRRLRLRCRLRLRRCLLRLRLRRCCLLRLRLRLRR+-LCR",
	"CCCCLCLRLRCCRLRLRLRCR+C-RLRLCL+-RR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CLL+CC-CRR+CC-CC",
	"CCCCCCCLLCCCCRCLRCRCCCCCCCCCRRCLRLCCLCCLRCCRCCRCCRLCCLCCLRLCRRCCCCCCCCCRCRLCRCCCCLLCCCCCCCLRRCRLLRLRLRLLCCCCCCCCLLRLRL",
	"CCCCLLRLRLRCRLRLRLLCCRCCLLRLRLRCRLRLRLLCCRCCLLRLRLRCRLRLRLLCCCCR+-LR+C-CRL+-CRLCCRLLRLRLLRLRCRLRLLRLCCRL++--LR+-LRCLR+",
}

// ChristmasAI is an AI that paints christmas related images.
type ChristmasAI struct {
	l sync.Mutex

	i        chan string
	counter  int
	selected string
}

// GetChannel receives the answer channel.
func (c *ChristmasAI) GetChannel(ch chan string) {
	c.l.Lock()
	defer c.l.Unlock()

	c.i = ch
}

// GetState gets the game state and computes an answer.
func (c *ChristmasAI) GetState(g *Game) {
	c.l.Lock()
	defer c.l.Unlock()

	if c.i == nil {
		return
	}

	if c.selected == "" {
		c.selected = ChristmasAIActions[rand.Intn(len(ChristmasAIActions))]
	}

	if g.Running {
		if g.Players[g.You].Active {
			if c.counter >= len(c.selected) {
				c.i <- ActionNOOP
				return
			}
			switch c.selected[c.counter] {
			case 'C', 'c', 'N', 'n':
				c.i <- ActionNOOP
			case 'L', 'l':
				c.i <- ActionTurnLeft
			case 'R', 'r':
				c.i <- ActionTurnRight
			case '+':
				c.i <- ActionFaster
			case '-':
				c.i <- ActionSlower
			default:
				log.Println("HeartAI: Unknown symbol ", c.selected[c.counter])
			}
			c.counter++
		}
	}
}

// Name returns the name of the AI.
func (c *ChristmasAI) Name() string {
	return "ChristmasAI"
}
