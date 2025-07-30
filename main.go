package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

const NutsPerBolt = 4
const EmptyStartingBolts = 2

var bestSolutionDepth int = 50

type State struct {
	nuts  [][]byte
	depth int
}
type Nut struct {
	colour byte
	pos    int
}

var queue chan State
var seen map[string]int

//var queue chan bool

func main() {
	//queue = make(chan bool)
	debug.SetGCPercent(50)
	seen = make(map[string][int])
	queue = make(chan State, 1000)
	initialState := loadFile("test.nuts")
	initialState.printState()
	queue <- initialState
	for {
		//state.printState()
		if len(queue) > 0 {
			state := <-queue
			go state.computeMoves()
		}
	}
}

func (s *State) stateToStr() string {
	string result
	for _, bolt := range s.nuts {
		result.append(string(bolt))
	}
	return result
}

func loadFile(path string) State {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	buffer, err := io.ReadAll(file)
	inputString := strings.ReplaceAll(string(buffer), "\n", "")
	inputString = strings.ReplaceAll(inputString, " ", "")
	inputLine := []byte(inputString)
	if len(inputLine)%4 != 0 {
		log.Printf("State length not divisible by nuts per bolt!\n%s", string(inputLine))
	}
	state2d := make([][]byte, len(inputLine)/NutsPerBolt)
	for bolt, nut := 0, 0; bolt < len(state2d); bolt, nut = bolt+1, nut+NutsPerBolt {
		state2d[bolt] = inputLine[nut : nut+NutsPerBolt]
	}
	state := State{
		nuts:  state2d,
		depth: 0,
	}
	return state
}

func (s *State) printState() {
	fmt.Printf("Depth: %d\t", s.depth)
	for _, bolt := range s.nuts {
		fmt.Printf("%s\t", string(bolt))
	}
	fmt.Printf("\n")
}

func (s *State) isComplete() bool {
	for i := range len(s.nuts) {
		colour := s.nuts[i][0]
		for j := range len(s.nuts[i]) {
			if s.nuts[i][j] != colour {
				return false
			}
		}
	}
	return true
}

func (s *State) getTop(src int) Nut {
	for i := range len(s.nuts[src]) {
		var value byte = s.nuts[src][i]
		if value != '_' {
			return Nut{
				colour: value,
				pos:    i,
			}
		}
	}
	return Nut{
		colour: '_',
		pos:    4,
	}
}

func (s *State) computeMoves() {
	if bestSolutionDepth != -1 && bestSolutionDepth <= s.depth {
		return
	}
	for i := range len(s.nuts) {
		nut := s.getTop(i)
		if nut.colour == '_' {
			continue
		}
		for j := range len(s.nuts) {
			if i == j {
				continue
			}
			if s.nuts[j][0] != '_' {
				continue
			}
			jTop := s.getTop(j)
			if jTop.colour == '_' || (jTop.colour == nut.colour && jTop.pos > 0) {
				var fits bool = true
				for k := range len(s.nuts[i]) {
					if k+nut.pos >= 4 {
						break
					}
					if s.nuts[i][k+nut.pos] == nut.colour {
						if s.nuts[j][k] != '_' {
							fits = false
							break
						}
					} else {
						break
					}

				}
				if !fits {
					continue
				}
				key := s.stateToStr()()
				if seen.has(key) && seen[key] >= (s.depth + 1){
					continue
				}
				seen[key] = s.depth + 1
				newState := State{
					nuts: make([][]byte, len(s.nuts)),
				}
				for i := range len(s.nuts) {
					newState.nuts[i] = make([]byte, 4)
					copy(newState.nuts[i], s.nuts[i])
				}

				newState.move(i, j, nut.pos, jTop.pos-1)
				newState.depth = s.depth + 1
				if newState.isComplete() {
					bestSolutionDepth = newState.depth
					fmt.Print("Solution: ")
					newState.printState()
				} else {
					queue <- newState
				}
			}
		}
	}
}

func (s *State) move(src int, dst int, srcIndex int, dstIndex int) {
	s.nuts[dst][dstIndex] = s.nuts[src][srcIndex]
	s.nuts[src][srcIndex] = '_'
}
