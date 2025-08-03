package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

const NutsPerBolt = 4
const EmptyStartingBolts = 2
const EmptyNut = 0

type DetailedMove struct {
	boltStart     int
	boltEnd       int
	color         byte
	count         int
	positionStart int
	positionEnd   int
}

type UserMove struct {
	boltStart int
	boltEnd   int
}
type Bolt struct {
	nuts [NutsPerBolt]byte
	idx  int
}
type State struct {
	bolts []Bolt
	moves []UserMove
}

var stateChannel chan State
var startTime time.Time
var stateMap map[string][]UserMove

func main() {
	initialState := loadFile("test.nuts")
	initialState.printState()
	startTime = time.Now()
	stateChannel = make(chan State, 5000)
	stateMap = make(map[string][]UserMove, 10000)
	addToMap(initialState, initialState)
	fmt.Printf("%d\n", time.Since(startTime).Milliseconds())

	var state State
	var win bool
	for {
		state = <-stateChannel
		win = play(state)
		if win || len(stateChannel) == 0 {
			break
		}
	}

	fmt.Printf("%s in %dms\n", winString(win), time.Since(startTime).Milliseconds())
	printMoveList(state.moves)
}
func winString(win bool) string {
	if win {
		return "won"
	} else {
		return "lost"
	}
}
func play(state State) bool {
	if gameWon(&state) {
		return true
	}

	for boltEndIdx, boltEnd := range state.bolts {
		if boltEnd.nuts[0] == EmptyNut {
			for boltStartIdx, boltStart := range state.bolts {
				if boltStart.idx != boltEnd.idx && boltStart.nuts[NutsPerBolt-1] != EmptyNut {
					pDetailedMove := getMove(&boltStart, &boltEnd)
					if pDetailedMove != nil {
						pDetailedMove.boltStart = boltStartIdx
						pDetailedMove.boltEnd = boltEndIdx
						newState := State{
							bolts: make([]Bolt, len(state.bolts)),
							moves: make([]UserMove, len(state.moves)),
						}
						copy(newState.bolts, state.bolts)
						copy(newState.moves, state.moves)
						newState.moves = append(newState.moves, createMove(boltStart.idx, boltEnd.idx))

						doMove(&newState, pDetailedMove)
						sortBolts(newState.bolts)
						addToMap(newState, state)
					}
				}
			}
		}
	}
	return false
}

func boltsToKey(pBolts []Bolt) string {
	var sb strings.Builder
	zero := [...]byte{0}
	for _, bolt := range pBolts {
		sb.WriteString(string(bytes.ReplaceAll(bolt.nuts[:], zero[:], []byte("_"))))
	}
	return sb.String()
}

func addToMap(state State, previous State) {
	key := boltsToKey(state.bolts)
	_, preexists := stateMap[key]
	if !preexists {
		stateMap[key] = state.moves
		stateChannel <- state
	} else {
		if len(stateMap[boltsToKey(previous.bolts)])+1 < len(stateMap[key]) {
			stateMap[key] = state.moves
		}
	}
}
func createMove(boltStart int, boltEnd int) UserMove {
	return UserMove{
		boltStart: boltStart,
		boltEnd:   boltEnd,
	}
}
func getMove(pBoltStart *Bolt, pBoltEnd *Bolt) *DetailedMove {
	nutCount, nutColor, nutPosition := getTopNuts(pBoltStart)
	if nutCount == NutsPerBolt {
		//Don't move a completed bolt
		return nil
	}
	blankCount := getTopEmpty(pBoltEnd)
	if nutPosition+nutCount == NutsPerBolt && blankCount == NutsPerBolt {
		//don't move a single color bolt to an empty bolt
		return nil
	}
	if blankCount >= nutCount &&
		(blankCount == NutsPerBolt || nutColor == pBoltEnd.nuts[blankCount]) {
		return &DetailedMove{
			color:         nutColor,
			count:         nutCount,
			positionStart: nutPosition,
			positionEnd:   blankCount - 1,
		}
	}
	return nil
}
func getTopNuts(pBolt *Bolt) (int, byte, int) {
	var topNut byte = EmptyNut

	count := 0
	for nutIdx, nut := range pBolt.nuts {
		if nut == EmptyNut {
			continue
		} else if topNut == EmptyNut {
			topNut = nut
			count++
		} else if nut == topNut {
			count++
		} else {
			return count, topNut, nutIdx - count
		}
	}
	return count, topNut, NutsPerBolt - count
}
func getTopEmpty(pBolt *Bolt) int {
	for nutIdx, nut := range pBolt.nuts {
		if nut != EmptyNut {
			return nutIdx
		}
	}
	return NutsPerBolt
}

func doMove(pState *State, pDetailedMove *DetailedMove) {
	for i := 0; i < pDetailedMove.count; i++ {
		pState.bolts[pDetailedMove.boltStart].nuts[pDetailedMove.positionStart+i] = EmptyNut
		pState.bolts[pDetailedMove.boltEnd].nuts[pDetailedMove.positionEnd-i] = pDetailedMove.color
	}
	//pState.printState()
}

func sortBolts(pBolts []Bolt) {
	sort.Slice(pBolts, func(i, j int) bool {
		for nutIdx := 0; nutIdx < NutsPerBolt; nutIdx++ {
			if pBolts[i].nuts[nutIdx] != pBolts[j].nuts[nutIdx] {
				return pBolts[i].nuts[nutIdx] < pBolts[j].nuts[nutIdx]
			}
		}
		return false
	})
}

func gameWon(pState *State) bool {
	win := true
	for _, bolt := range pState.bolts {
		for nut := 0; nut < NutsPerBolt-1; nut++ {
			if bolt.nuts[nut] != bolt.nuts[NutsPerBolt-1] {
				win = false
			}
		}
	}
	return win
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
	twoBolts := [NutsPerBolt * EmptyStartingBolts]byte{EmptyNut}
	inputLine = append(inputLine, twoBolts[:]...)
	if len(inputLine)%NutsPerBolt != 0 {
		log.Printf("State length not divisible by nuts per bolt!\n%s", string(inputLine))
	}
	boltList := make([]Bolt, len(inputLine)/NutsPerBolt)
	for boltIdx := range boltList {
		boltList[boltIdx].idx = boltIdx
		boltList[boltIdx].nuts = [NutsPerBolt]byte(inputLine[boltIdx*NutsPerBolt : boltIdx*NutsPerBolt+NutsPerBolt])
	}
	state := State{
		bolts: boltList,
		moves: make([]UserMove, 0),
	}
	return state
}

func (s *State) printState() {
	fmt.Printf("%d ", len(s.moves))
	for _, bolt := range s.bolts {
		zero := [...]byte{0}
		stringBolt := string(bytes.ReplaceAll(bolt.nuts[:], zero[:], []byte("_")))
		fmt.Printf("%s\t", stringBolt)
	}
	fmt.Println("")
}
func printMoveList(moves []UserMove) {
	for moveIdx, move := range moves {
		fmt.Printf("%d: %d -> %d\n", moveIdx+1, move.boltStart+1, move.boltEnd+1)
	}
}
