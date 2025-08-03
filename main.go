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

const NumBolts = 14
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
type StateToMovesMap map[[14][4]byte][]UserMove
type Bolt struct {
	nuts [NutsPerBolt]byte
	idx  int
}
type State struct {
	bolts [NumBolts]Bolt
	moves []UserMove
}

var stateChannel chan State
var startTime time.Time
var stateToMovesMap StateToMovesMap

func main() {
	initialState := loadFile("test.nuts")
	initialState.printState()
	startTime = time.Now()
	stateChannel = make(chan State, 5000)
	stateToMovesMap = make(StateToMovesMap, 10000)
	stateToMovesMap.addToMap(&initialState, &initialState)
	fmt.Printf("%d\n", time.Since(startTime).Milliseconds())

	var state State
	var win bool
	for {
		state = <-stateChannel
		win = state.play()
		if win || len(stateChannel) == 0 {
			break
		}
	}

	fmt.Printf("won: %t in %dms\n", win, time.Since(startTime).Milliseconds())
	//state.printMoveList()
}

func (pState *State) play() bool {
	if pState.gameWon() {
		return true
	}

	for boltEndIdx, boltEnd := range pState.bolts {
		if boltEnd.nuts[0] == EmptyNut {
			for boltStartIdx, boltStart := range pState.bolts {
				if boltEndIdx != boltStartIdx && boltStart.nuts[NutsPerBolt-1] != EmptyNut {
					pDetailedMove := getMove(&boltStart, &boltEnd)
					if pDetailedMove != nil {
						pDetailedMove.boltStart = boltStartIdx
						pDetailedMove.boltEnd = boltEndIdx

						newState := pState.deepCloneState()
						newState.moves = append(newState.moves, createMove(boltStart.idx, boltEnd.idx))

						newState.doMove(pDetailedMove)
						newState.sortBolts()
						stateToMovesMap.addToMap(&newState, pState)
					}
				}
			}
		}
	}
	return false
}

func (pState *State) gameWon() bool {
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

func createMove(boltStart int, boltEnd int) UserMove {
	return UserMove{
		boltStart: boltStart,
		boltEnd:   boltEnd,
	}
}

func getMove(pBoltStart *Bolt, pBoltEnd *Bolt) *DetailedMove {
	nutCount, nutColor, nutPosition := pBoltStart.getTopNuts()
	if nutCount == NutsPerBolt {
		//Don't move a completed bolt
		return nil
	}
	blankCount := pBoltEnd.getTopEmpty()
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

func (pBolt *Bolt) getTopNuts() (int, byte, int) {
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

func (pBolt *Bolt) getTopEmpty() int {
	for nutIdx, nut := range pBolt.nuts {
		if nut != EmptyNut {
			return nutIdx
		}
	}
	return NutsPerBolt
}

func (pState *State) deepCloneState() State {
	newState := State{
		bolts: pState.bolts,
		moves: make([]UserMove, len(pState.moves)),
	}
	copy(newState.moves, pState.moves)
	return newState
}

func (pState *State) doMove(pDetailedMove *DetailedMove) {
	for i := 0; i < pDetailedMove.count; i++ {
		pState.bolts[pDetailedMove.boltStart].nuts[pDetailedMove.positionStart+i] = EmptyNut
		pState.bolts[pDetailedMove.boltEnd].nuts[pDetailedMove.positionEnd-i] = pDetailedMove.color
	}
	//pState.printState()
}

func (pState *State) sortBolts() {
	boltsAsSlice := pState.bolts[:]
	sort.Slice(boltsAsSlice, func(i, j int) bool {
		return compareNuts(boltsAsSlice[i].nuts, boltsAsSlice[j].nuts)
	})
}
func compareNuts(bolt1 [NutsPerBolt]byte, bolt2 [NutsPerBolt]byte) bool {
	for nutIdx := 0; nutIdx < NutsPerBolt; nutIdx++ {
		if bolt1[nutIdx] != bolt2[nutIdx] {
			return bolt1[nutIdx] < bolt2[nutIdx]
		}
	}
	return false
}

func (pMap StateToMovesMap) addToMap(pNewState *State, pOldState *State) {
	key := pNewState.makeMapKey()
	_, preexists := pMap[key]
	if !preexists {
		pMap[key] = pNewState.moves
		stateChannel <- *pNewState
	} else {
		if len(pMap[pOldState.makeMapKey()])+1 < len(pMap[key]) {
			pMap[key] = pNewState.moves
		}
	}
}

func (pState *State) makeMapKeyOld() string {
	var sb strings.Builder
	zero := [...]byte{0}
	for _, bolt := range pState.bolts {
		sb.WriteString(string(bytes.ReplaceAll(bolt.nuts[:], zero[:], []byte("_"))))
	}
	return sb.String()
}

func (pState *State) makeMapKey() [NumBolts][NutsPerBolt]byte {
	key := [NumBolts][NutsPerBolt]byte{}
	for boltIdx, bolt := range pState.bolts {
		key[boltIdx] = bolt.nuts
	}
	return key
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
	if len(inputLine)%NumBolts != 0 {
		log.Printf("State length not divisible by num bolts!\n%s", string(inputLine))
	}
	var boltList [NumBolts]Bolt
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

func (pState *State) printState() {
	fmt.Printf("%d ", len(pState.moves))
	for _, bolt := range pState.bolts {
		zero := [...]byte{0}
		stringBolt := string(bytes.ReplaceAll(bolt.nuts[:], zero[:], []byte("_")))
		fmt.Printf("%s\t", stringBolt)
	}
	fmt.Println("")
}

func (pState *State) printMoveList() {
	for moveIdx, move := range pState.moves {
		fmt.Printf("%d: %d -> %d\n", moveIdx+1, move.boltStart+1, move.boltEnd+1)
	}
}
