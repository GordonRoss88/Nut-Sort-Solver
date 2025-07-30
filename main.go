package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const NutsPerBolt = 4
const EmptyStartingBolts = 2
const EmptyNut = 0

type Swap struct {
	boltStart int
	boltEnd   int
}
type State struct {
	nuts  [][NutsPerBolt]byte
	swaps []Swap
}

func main() {
	initialState := loadFile("test.nuts")
	initialState.printState()
	start := time.Now()
	fmt.Printf("%d\n", time.Since(start).Milliseconds())
	numWins := 0
	minSteps := 50
	play(initialState, start, &numWins, &minSteps)
	fmt.Printf("done %d %d %d\n", numWins, minSteps, time.Since(start).Milliseconds())
}
func play(state State, start time.Time, pNumWins *int, pMinSteps *int) {
	var waitGroup sync.WaitGroup
	if len(state.swaps) < *pMinSteps {
		didSwap := false
		for boltStart := 0; boltStart < len(state.nuts); boltStart++ {
			for boltEnd := 0; boltEnd < len(state.nuts); boltEnd++ {
				if boltStart != boltEnd {
					fmt.Printf("%d %d\n", boltStart, boltEnd)
					swapCount := getSwapCount(&state, boltStart, boltEnd)
					if swapCount > 0 {
						didSwap = true
						newState := State{
							nuts:  make([][NutsPerBolt]byte, len(state.nuts)),
							swaps: make([]Swap, len(state.swaps)),
						}
						copy(newState.nuts, state.nuts)
						copy(newState.swaps, state.swaps)
						newState.swaps = append(newState.swaps, createSwap(boltStart, boltEnd))

						swap(&newState, boltStart, boltEnd, swapCount)
						waitGroup.Add(1)
						go play2(newState, start, pNumWins, pMinSteps, &waitGroup)
					}
				}
			}
		}

		if didSwap == false {
			if gameWon(&state) {
				(*pNumWins)++
				if len(state.swaps) < *pMinSteps {
					*pMinSteps = len(state.swaps)
					fmt.Printf("%d %d %d\n", time.Since(start).Milliseconds(), *pNumWins, len(state.swaps))
				}
			}
		}
	}
	waitGroup.Wait()
}
func play2(state State, start time.Time, pNumWins *int, pMinSteps *int, pWaitGroup *sync.WaitGroup) {
	play3(state, start, pNumWins, pMinSteps)
	pWaitGroup.Done()
	println("done")
}
func play3(state State, start time.Time, pNumWins *int, pMinSteps *int) {
	if len(state.swaps) < *pMinSteps {
		didSwap := false
		for boltStart := 0; boltStart < len(state.nuts); boltStart++ {
			for boltEnd := 0; boltEnd < len(state.nuts); boltEnd++ {
				if boltStart != boltEnd {
					swapCount := getSwapCount(&state, boltStart, boltEnd)
					if swapCount > 0 {
						didSwap = true
						newState := State{
							nuts:  make([][NutsPerBolt]byte, len(state.nuts)),
							swaps: make([]Swap, len(state.swaps)),
						}
						copy(newState.nuts, state.nuts)
						copy(newState.swaps, state.swaps)
						newState.swaps = append(newState.swaps, createSwap(boltStart, boltEnd))

						swap(&newState, boltStart, boltEnd, swapCount)
						play3(newState, start, pNumWins, pMinSteps)
					}
				}
			}
		}

		if didSwap == false {
			if gameWon(&state) {
				(*pNumWins)++
				if len(state.swaps) < *pMinSteps {
					*pMinSteps = len(state.swaps)
					fmt.Printf("%d %d %d\n", time.Since(start).Milliseconds(), *pNumWins, len(state.swaps))
					printMoveList(state.swaps)
				}
			}
		}
	}
}
func createSwap(boltStart int, boltEnd int) Swap {
	return Swap{
		boltStart: boltStart,
		boltEnd:   boltEnd,
	}
}
func getSwapCount(pState *State, boltStart int, boltEnd int) int {
	countStart, topNutColor, position := getTopNuts(pState, boltStart)
	countEnd := getTopEmpty(pState, boltEnd)
	if countStart > 0 && countEnd > 0 &&
		countEnd >= countStart && countStart != NutsPerBolt &&
		(countEnd == NutsPerBolt || topNutColor == pState.nuts[boltEnd][countEnd]) {
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
		if pState.nuts[bolt][nut] == EmptyNut {
			continue
		} else if topNut == EmptyNut {
			topNut = pState.nuts[bolt][nut]
			count++
		} else if pState.nuts[bolt][nut] == topNut {
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
		if pState.nuts[bolt][nut] != EmptyNut {
			break
		}
	}
	return nut
}

func swap(pState *State, boltStart int, boltEnd int, count int) {
	var color byte
	for nut := 0; nut < NutsPerBolt; nut++ {
		if pState.nuts[boltStart][nut] != EmptyNut {
			color = pState.nuts[boltStart][nut]
			for i := 0; i < count; i++ {
				pState.nuts[boltStart][nut+i] = EmptyNut
			}
			break
		}
	}

	for nut := NutsPerBolt - 1; nut >= 0; nut-- {
		if pState.nuts[boltEnd][nut] == EmptyNut {
			for i := 0; i < count; i++ {
				pState.nuts[boltEnd][nut-i] = color
			}
			break
		}
	}
	//pState.printState()
}

func gameWon(pState *State) bool {
	win := true
	for bolt := 0; bolt < len(pState.nuts); bolt++ {
		for nut := 0; nut < NutsPerBolt-1; nut++ {
			if pState.nuts[bolt][nut] != pState.nuts[bolt][NutsPerBolt-1] {
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
	state2d := make([][NutsPerBolt]byte, len(inputLine)/NutsPerBolt)
	for bolt, nut := 0, 0; bolt < len(state2d); bolt, nut = bolt+1, nut+NutsPerBolt {
		state2d[bolt] = [NutsPerBolt]byte(inputLine[nut : nut+NutsPerBolt])
	}
	state := State{
		nuts:  state2d,
		swaps: make([]Swap, 0),
	}
	return state
}

func (s *State) printState() {
	fmt.Printf("%d ", len(s.swaps))
	for _, bolt := range s.nuts {
		zero := [...]byte{0}
		stringBolt := string(bytes.ReplaceAll(bolt[:], zero[:], []byte("_")))
		fmt.Printf("%s\t", stringBolt)
	}
	fmt.Println("")
}
func printMoveList(swaps []Swap) {
	for moveIdx, swap := range swaps {
		fmt.Printf("%d: %d -> %d\n", moveIdx+1, swap.boltStart, swap.boltEnd)
	}
}
