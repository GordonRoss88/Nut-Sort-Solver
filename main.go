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

type State struct {
	nuts    [][4]byte
	swaps   []SwapAction
	history string
}

type SwapAction struct {
	src       int
	dst       int
	count     int
	destStart int
}

var solved chan *State

var wg = new(sync.WaitGroup)
var solutions = 0

func main() {
	initialState := loadFile("test.nuts")
	initialState.printState()
	solved = make(chan *State)

	const maxConcurrentGoroutines = 34

	var limiter = make(chan struct{}, maxConcurrentGoroutines)

	go func() {
		initialState.solve(limiter)
	}()
	//wg.Wait()
	for {
		time.Sleep(time.Second)
		log.Printf("found: %d solutions", solutions)
	}
	//finalState := <-solved
	//log.Println("Solved!")
	//finalState.printState()
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
	twoBolts := [NutsPerBolt * EmptyStartingBolts]byte{0}
	inputLine = append(inputLine, twoBolts[:]...)
	if len(inputLine)%4 != 0 {
		log.Printf("State length not divisible by nuts per bolt!\n%s", string(inputLine))
	}
	state2d := make([][4]byte, len(inputLine)/NutsPerBolt)
	for bolt, nut := 0, 0; bolt < len(state2d); bolt, nut = bolt+1, nut+NutsPerBolt {
		state2d[bolt] = [4]byte(inputLine[nut : nut+NutsPerBolt])
	}
	state := State{
		nuts:  state2d,
		swaps: make([]SwapAction, 0),
	}
	return state
}

func (s *State) printState() {
	for _, bolt := range s.nuts {
		zero := [...]byte{0}
		stringBolt := string(bytes.ReplaceAll(bolt[:], zero[:], []byte("_")))
		fmt.Printf("%s\t", stringBolt)
	}
	fmt.Printf("\n")
	for _, swap := range s.swaps {
		fmt.Printf("%d>%d  ", swap.src, swap.dst)
	}
}

func (s *State) solve(limiter chan struct{}) {

	limiter <- struct{}{}
	go func() {
		//defer wg.Done()

		s.doAllSwaps(limiter)
	}()
}

func (s *State) doSwap(swap SwapAction) *State {
	newState := State{
		nuts:  make([][4]byte, len(s.nuts)),
		swaps: append(s.swaps, swap),
	}

	for i := 0; i < len(newState.nuts); i++ {
		newState.nuts[i] = s.nuts[i]
	}

	for numSwapped := 0; numSwapped < swap.count; numSwapped++ {
		srcNut := (swap.count - 1) - numSwapped
		dstNut := swap.destStart - numSwapped
		if dstNut < 0 {
			if newState.nuts[swap.src][srcNut] != 0 {
				log.Printf("ran out of dest spots")
			}
			break
		}
		newState.nuts[swap.dst][dstNut] = newState.nuts[swap.src][srcNut]
		newState.nuts[swap.src][srcNut] = 0
	}
	return &newState
}

func (s *State) doAllSwaps(limiter chan struct{}) {

	for srcIndex, srcBolt := range s.nuts {

		srcCount, numBlanks, willLeaveBlank, srcNutType := s.topNuts(srcBolt)
		if srcCount == 4 && numBlanks == 0 {
			continue
		}

		for destIndex, destBolt := range s.nuts {
			if srcIndex == destIndex || destBolt[0] != 0 {
				continue
			}

			lastBlank := 0
			var destNutType byte = 0
			for nutIndex, nut := range destBolt {
				if nut != 0 {
					lastBlank = nutIndex - 1
					destNutType = nut
					break
				} else if nutIndex == 3 {
					lastBlank = 3
				}
			}
			if destNutType == 0 && willLeaveBlank {
				continue
			}
			if (destNutType != 0 && destNutType != srcNutType) || lastBlank+1 < srcCount-numBlanks {
				continue
			}

			swap := SwapAction{
				src:       srcIndex,
				dst:       destIndex,
				count:     srcCount,
				destStart: lastBlank,
			}
			newState := s.doSwap(swap)
			if newState.isSolved() {
				//solved <- newState
				solutions++
				//log.Printf("SOLVED")
				//newState.printState()
				break
			}
			<-limiter

			newState.solve(limiter)
		}
	}
	<-limiter
}

func (s *State) topNuts(srcBolt [4]byte) (int, int, bool, byte) {
	srcCount := 0
	numBlanks := 0
	willLeaveBlank := false
	var srcNutType byte = 0
	for nutIndex, nut := range srcBolt {
		if srcNutType == 0 && nut != 0 {
			srcNutType = nut
			if nutIndex != 0 {
				numBlanks = nutIndex
			}
		}
		if srcNutType != 0 && nut != srcNutType {
			srcCount = nutIndex
			break
		}
		if nutIndex == 3 {
			srcCount = 4
			willLeaveBlank = true
		}
	}
	return srcCount, numBlanks, willLeaveBlank, srcNutType
}

func (s *State) isSolved() bool {
	for _, bolt := range s.nuts {
		for nut := 0; nut < 4; nut++ {
			if bolt[0] != bolt[nut] {
				return false
			}
		}
	}
	return true
}
