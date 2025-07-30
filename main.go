package main

import (
	"bytes"
	"encoding/binary"
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
const NumBolts = 14

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

type Bolt struct {
	nuts          [4]byte
	startingIndex int
}

type NutOrdering [NumBolts]Bolt

type MapValue struct {
	previous  NutOrdering
	swapCount int
}

// const map
var stateMap = make(map[NutOrdering]MapValue, 10000)

var pendingOrders chan NutOrdering
var solved chan NutOrdering
var solutions = 0

func main() {
	initialState := loadFile("test.nuts")
	//initialState.printState()
	solved = make(chan NutOrdering, 1)
	pendingOrders = make(chan NutOrdering, 5000)

	ord := NutOrdering{}
	for i := 0; i < NumBolts; i++ {
		ord[i] = Bolt{
			nuts:          initialState.nuts[i],
			startingIndex: i,
		}
	}

	//NutOrdering(initialState.nuts)
	ord = sortNutOrdering(ord)
	stateMap[ord] = MapValue{
		swapCount: 0,
	}
	//log.Printf(stateMap)
	test := stateMap[ord].previous
	log.Printf("%s", test[0])
	pendingOrders <- ord

	work()
	solutionPrinter()
	for {
		time.Sleep(time.Second)
		//log.Printf("found: %d solutions", solutions)
	}
}

func printSolution(order NutOrdering) {
	fmt.Printf("%d\t", stateMap[order].swapCount)
	printState(order)
	if stateMap[order].previous != order {
		printSolution(stateMap[order].previous)
	}
}
func solutionPrinter() {
	go func() {
		for {
			newSolution := <-solved
			printSolution(newSolution)
		}
	}()
}
func work() {
	go func() {
		for {
			newOrder := <-pendingOrders
			log.Printf("newOrder: %d", len(pendingOrders))
			doAllSwaps(newOrder)
		}
	}()
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

func printState(order NutOrdering) {
	for _, bolt := range order {
		zero := [...]byte{0}
		stringBolt := string(bytes.ReplaceAll(bolt.nuts[:], zero[:], []byte("_")))
		fmt.Printf("%s\t", stringBolt)
	}
	fmt.Printf("\n")
	//for _, swap := range s.swaps {
	//	fmt.Printf("%d>%d  ", swap.src, swap.dst)
	//}
}

func doSwap(nutOrdering NutOrdering, swap SwapAction) NutOrdering {
	//nutOrdering :=
	//newState := State{
	//	nuts:  make([][4]byte, len(nutOrdering)),
	//	//swaps: append(s.swaps, swap),
	//}

	//for i := 0; i < len(nutOrdering); i++ {
	//	nutOrdering[i] = nutOrdering[i]
	//}

	for numSwapped := 0; numSwapped < swap.count; numSwapped++ {
		srcNut := (swap.count - 1) - numSwapped
		dstNut := swap.destStart - numSwapped
		if dstNut < 0 {
			if nutOrdering[swap.src].nuts[srcNut] != 0 {
				log.Printf("ran out of dest spots")
			}
			break
		}
		nutOrdering[swap.dst].nuts[dstNut] = nutOrdering[swap.src].nuts[srcNut]
		nutOrdering[swap.src].nuts[srcNut] = 0
	}
	return nutOrdering
}

func doAllSwaps(nutOrdering NutOrdering) {

	if isSolved(nutOrdering) {
		solved <- nutOrdering
		solutions++
		return
	}
	for srcIndex, srcBolt := range nutOrdering {

		srcCount, numBlanks, willLeaveBlank, srcNutType := topNuts(srcBolt.nuts)
		if srcCount == 4 && numBlanks == 0 {
			continue
		}

		for destIndex, destBolt := range nutOrdering {
			if srcIndex == destIndex || destBolt.nuts[0] != 0 {
				continue
			}

			lastBlank := 0
			var destNutType byte = 0
			for nutIndex, nut := range destBolt.nuts {
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
			newNutOrdering := doSwap(nutOrdering, swap)

			newNutOrdering = sortNutOrdering(newNutOrdering)
			addNewOrder(newNutOrdering, nutOrdering)
			//newState.solve(limiter)
		}
	}
}

func addNewOrder(nutOrdering NutOrdering, previous NutOrdering) {
	_, preexists := stateMap[nutOrdering]
	if !preexists {
		stateMap[nutOrdering] = MapValue{
			previous:  previous,
			swapCount: stateMap[previous].swapCount + 1,
		}
		pendingOrders <- nutOrdering
	} else {
		if stateMap[previous].swapCount+1 < stateMap[nutOrdering].swapCount {
			stateMap[nutOrdering] = MapValue{
				previous:  previous,
				swapCount: stateMap[previous].swapCount + 1,
			}
		}
	}
}

func compareNuts(nut1 [4]byte, nut2 [4]byte) bool {
	return binary.BigEndian.Uint32(nut1[:]) < binary.BigEndian.Uint32(nut2[:])
}

func sortNutOrdering(nutOrdering NutOrdering) NutOrdering {
	asSlice := nutOrdering[:]
	sort.Slice(asSlice, func(i, j int) bool { return compareNuts(asSlice[i].nuts, asSlice[j].nuts) })
	return NutOrdering(asSlice)
}

func topNuts(srcBolt [4]byte) (int, int, bool, byte) {
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

func isSolved(order NutOrdering) bool {
	for _, bolt := range order {
		for nut := 0; nut < 4; nut++ {
			if bolt.nuts[0] != bolt.nuts[nut] {
				return false
			}
		}
	}
	return true
}
