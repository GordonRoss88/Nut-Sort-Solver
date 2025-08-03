package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const NutsPerBolt = 4
const EmptyStartingBolts = 2
const EmptyNut = 0

type Move struct {
	boltStart int
	boltEnd   int
}
type Bolt struct {
	nuts    [NutsPerBolt]byte
	boltIdx int
}
type State struct {
	bolts []Bolt
	moves []Move
}

func main() {
	initialState := loadFile("test.nuts")
	initialState.printState()
	start := time.Now()
	fmt.Printf("%d\n", time.Since(start).Milliseconds())
	numWins := 0
	minMoves := 50
	play(initialState, start, &numWins, &minMoves)
	fmt.Printf("done %d %d %d\n", numWins, minMoves, time.Since(start).Milliseconds())
}
func play(state State, start time.Time, pNumWins *int, pMinMoves *int) {
	if len(state.moves) < *pMinMoves {
		didMove := false
		for boltStart := 0; boltStart < len(state.bolts); boltStart++ {
			for boltEnd := 0; boltEnd < len(state.bolts); boltEnd++ {
				if boltStart != boltEnd {
					moveCount := getMoveCount(&state, boltStart, boltEnd)
					if moveCount > 0 {
						didMove = true
						newState := State{
							bolts: make([]Bolt, len(state.bolts)),
							moves: make([]Move, len(state.moves)),
						}
						copy(newState.bolts, state.bolts)
						copy(newState.moves, state.moves)
						newState.moves = append(newState.moves, createMove(boltStart, boltEnd))

						move(&newState, boltStart, boltEnd, moveCount)
						play(newState, start, pNumWins, pMinMoves)
					}
				}
			}
		}

		if didMove == false {
			if gameWon(&state) {
				(*pNumWins)++
				if len(state.moves) < *pMinMoves {
					*pMinMoves = len(state.moves)
					fmt.Printf("%d %d %d\n", time.Since(start).Milliseconds(), *pNumWins, len(state.moves))
				}
			}
		}
	}
}
func createMove(boltStart int, boltEnd int) Move {
	return Move{
		boltStart: boltStart,
		boltEnd:   boltEnd,
	}
}
func getMoveCount(pState *State, boltStart int, boltEnd int) int {
	countStart, topNutColor, position := getTopNuts(pState, boltStart)
	countEnd := getTopEmpty(pState, boltEnd)
	if countStart > 0 && countEnd > 0 &&
		countEnd >= countStart && countStart != NutsPerBolt &&
		(countEnd == NutsPerBolt || topNutColor == pState.bolts[boltEnd].nuts[countEnd]) {
		if position+countStart == NutsPerBolt && countEnd == NutsPerBolt {
			return 0
		}
		return countStart
	}
	return 0
}
func getTopNuts(pState *State, bolt int) (int, byte, int) {
	var topNut byte = EmptyNut

	count := 0
	var nut int
	for nut = 0; nut < NutsPerBolt; nut++ {
		if pState.bolts[bolt].nuts[nut] == EmptyNut {
			continue
		} else if topNut == EmptyNut {
			topNut = pState.bolts[bolt].nuts[nut]
			count++
		} else if pState.bolts[bolt].nuts[nut] == topNut {
			count++
		} else {
			break
		}
	}
	return count, topNut, nut - count
}
func getTopEmpty(pState *State, bolt int) int {
	var nut int
	for nut = 0; nut < NutsPerBolt; nut++ {
		if pState.bolts[bolt].nuts[nut] != EmptyNut {
			break
		}
	}
	return nut
}

func move(pState *State, boltStart int, boltEnd int, count int) {
	var color byte
	for nut := 0; nut < NutsPerBolt; nut++ {
		if pState.bolts[boltStart].nuts[nut] != EmptyNut {
			color = pState.bolts[boltStart].nuts[nut]
			for i := 0; i < count; i++ {
				pState.bolts[boltStart].nuts[nut+i] = EmptyNut
			}
			break
		}
	}

	for nut := NutsPerBolt - 1; nut >= 0; nut-- {
		if pState.bolts[boltEnd].nuts[nut] == EmptyNut {
			for i := 0; i < count; i++ {
				pState.bolts[boltEnd].nuts[nut-i] = color
			}
			break
		}
	}
	//pState.printState()
}

func gameWon(pState *State) bool {
	win := true
	for bolt := 0; bolt < len(pState.bolts); bolt++ {
		for nut := 0; nut < NutsPerBolt-1; nut++ {
			if pState.bolts[bolt].nuts[nut] != pState.bolts[bolt].nuts[NutsPerBolt-1] {
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
		boltList[boltIdx].boltIdx = boltIdx
		boltList[boltIdx].nuts = [NutsPerBolt]byte(inputLine[boltIdx*NutsPerBolt : boltIdx*NutsPerBolt+NutsPerBolt])
	}
	state := State{
		bolts: boltList,
		moves: make([]Move, 0),
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
func printMoveList(moves []Move) {
	for moveIdx, move := range moves {
		fmt.Printf("%d: %d -> %d\n", moveIdx+1, move.boltStart, move.boltEnd)
	}
}
