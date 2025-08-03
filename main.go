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
		for _, boltEnd := range state.bolts {
			if boltEnd.nuts[0] == EmptyNut {
				for _, boltStart := range state.bolts {
					if boltStart.idx != boltEnd.idx && boltStart.nuts[NutsPerBolt-1] != EmptyNut {
						pDetailedMove := getMove(&boltStart, &boltEnd)
						if pDetailedMove != nil {
							didMove = true
							newState := State{
								bolts: make([]Bolt, len(state.bolts)),
								moves: make([]UserMove, len(state.moves)),
							}
							copy(newState.bolts, state.bolts)
							copy(newState.moves, state.moves)
							newState.moves = append(newState.moves, createMove(boltStart.idx, boltEnd.idx))

							doMove(&newState, pDetailedMove)
							play(newState, start, pNumWins, pMinMoves)
						}
					}
				}
			}
		}

		if !didMove {
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
			boltStart:     pBoltStart.idx,
			boltEnd:       pBoltEnd.idx,
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
	var nut int
	for nut = 0; nut < NutsPerBolt; nut++ {
		if pBolt.nuts[nut] == EmptyNut {
			continue
		} else if topNut == EmptyNut {
			topNut = pBolt.nuts[nut]
			count++
		} else if pBolt.nuts[nut] == topNut {
			count++
		} else {
			break
		}
	}
	return count, topNut, nut - count
}
func getTopEmpty(pBolt *Bolt) int {
	var nut int
	for nut = 0; nut < NutsPerBolt; nut++ {
		if pBolt.nuts[nut] != EmptyNut {
			break
		}
	}
	return nut
}

func doMove(pState *State, pDetailedMove *DetailedMove) {
	for i := 0; i < pDetailedMove.count; i++ {
		pState.bolts[pDetailedMove.boltStart].nuts[pDetailedMove.positionStart+i] = EmptyNut
		pState.bolts[pDetailedMove.boltEnd].nuts[pDetailedMove.positionEnd-i] = pDetailedMove.color
	}
	//pState.printState()
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
		fmt.Printf("%d: %d -> %d\n", moveIdx+1, move.boltStart, move.boltEnd)
	}
}
