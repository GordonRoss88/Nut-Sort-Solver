package main

/*
robotgo Windows setup:
winget install MartinStorsjo.LLVM-MinGW.UCRT
Add to PATH variable: %USERPROFILE%\AppData\Local\Microsoft\WinGet\Packages\MartinStorsjo.LLVM-MinGW.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\llvm-mingw-20250709-ucrt-x86_64\bin
Close and reopen vsCode
Import "github.com/go-vgo/robotgo"

notes, multi-monitor support with robotgo.GetPixelColor() seems to be broken on the 2nd monitor.
*/

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
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
	/*for {
		robotgo.MilliSleep(200)
		x, y := robotgo.Location()
		color := robotgo.GetPixelColor(x, y)
		fmt.Println("x,y = color ", x, y, color)
	}*/
	startTime = time.Now()

	fmt.Printf("%d\n", time.Since(startTime).Milliseconds())
	numBolts := GetNumBolts()
	fmt.Printf("%d bolt count: %d\n", time.Since(startTime).Milliseconds(), numBolts)

	initialState := LoadGameState(numBolts)
	//initialState := loadFile("test.nuts")
	fmt.Printf("%d Game state: ", time.Since(startTime).Milliseconds())
	initialState.printState(numBolts)

	stateChannel = make(chan State, 10000)
	stateToMovesMap = make(StateToMovesMap, 12000)
	stateToMovesMap.addToMap(&initialState, &initialState)

	var state State
	var win bool

	for {
		state = <-stateChannel
		win = state.play(numBolts)
		if win || len(stateChannel) == 0 {
			break
		}
	}

	fmt.Printf("won: %t in %dms in %d moves\n", win, time.Since(startTime).Milliseconds(), len(state.moves))
	//state.printMoveList()
	state.winGame()

}

func (pState *State) winGame() {
	robotgo.MouseSleep = 25

	for moveIdx, move := range pState.moves {
		fmt.Printf("%d: %d -> %d\n", moveIdx+1, move.boltStart+1, move.boltEnd+1)
		robotgo.Move(BoltLocations[move.boltStart][0], BoltLocations[move.boltStart][1])
		robotgo.Click("left", true)
		robotgo.Move(BoltLocations[move.boltEnd][0], BoltLocations[move.boltEnd][1])
		robotgo.Click("left", true)
	}
}

func GetNumBolts() int {
	if BoltCountIs14() {
		return 14
	} else if BoltCountIs11() {
		return 11
	} else {
		return -1
	}
}

const BoltColor = "ffffff"
const Red = "ad010b"
const Orange = "c06c05"
const Yellow = "f59e00"
const Green = "028b0b"
const Sky = "008fc6"
const Blue = "373cc0"
const Fushia = "c30066"
const Purple = "863f95"
const Gray = "759bad"
const Pink = "fa4a43"
const Brown = "804731"
const Metal = "3d3d3d"

var colorList = [13]string{"000000", Red, Orange, Yellow, Green, Sky, Blue, Fushia, Purple, Gray, Pink, Brown, Metal}
var BoltLocations [14][2]int

func BoltCountIs14() bool {
	const BoltStartX = 413
	const BoltOffsetX = 185
	const BoltStartY = 298
	const BoltOffsetY = 300

	var x int
	var y int
	for boltIdx := 0; boltIdx < 14; boltIdx++ {
		if boltIdx < 7 {
			x = BoltStartX + boltIdx*BoltOffsetX
			y = BoltStartY
		} else {
			x = BoltStartX + (boltIdx-7)*BoltOffsetX
			y = BoltStartY + BoltOffsetY
		}
		BoltLocations[boltIdx][0] = x
		BoltLocations[boltIdx][1] = y
	}

	for boltIdx := 0; boltIdx < 14; boltIdx++ {
		//fmt.Println(x, y, robotgo.GetPixelColor(x, y))
		if robotgo.GetPixelColor(BoltLocations[boltIdx][0], BoltLocations[boltIdx][1]) != BoltColor {
			return false
		}
	}
	return true
}
func BoltCountIs11() bool {
	const BoltStartX = 371
	const BoltOffsetX = 241
	const BoltStartY = 229
	const BoltOffsetY = 389
	const BoltRowOffsetX = 120

	var x int
	var y int
	for boltIdx := 0; boltIdx < 14; boltIdx++ {
		if boltIdx < 6 {
			x = BoltStartX + boltIdx*BoltOffsetX
			y = BoltStartY
		} else {
			x = BoltStartX + BoltRowOffsetX + (boltIdx-6)*BoltOffsetX
			y = BoltStartY + BoltOffsetY
		}
		BoltLocations[boltIdx][0] = x
		BoltLocations[boltIdx][1] = y
	}

	for boltIdx := 0; boltIdx < 11; boltIdx++ {
		//fmt.Println(x, y, robotgo.GetPixelColor(x, y))
		if robotgo.GetPixelColor(BoltLocations[boltIdx][0], BoltLocations[boltIdx][1]) != BoltColor {
			return false
		}
	}
	return true
}

func LoadGameState(numBolts int) State {

	var boltList [NumBolts]Bolt

	var nutOffsetX int
	var nutOffsetY int
	var nutOffsetVert int
	if numBolts == 14 {
		nutOffsetX = -64
		nutOffsetY = 33
		nutOffsetVert = 51
	} else if numBolts == 11 {
		nutOffsetX = -90
		nutOffsetY = 42
		nutOffsetVert = 68
	}

	for boltIdx := 0; boltIdx < numBolts; boltIdx++ {
		boltList[boltIdx].idx = boltIdx
		for nutIdx := 0; nutIdx < NutsPerBolt; nutIdx++ {
			pixelColor := robotgo.GetPixelColor(BoltLocations[boltIdx][0]+nutOffsetX, BoltLocations[boltIdx][1]+nutOffsetY+nutIdx*nutOffsetVert)
			for colorIdx, colorValue := range colorList {
				if pixelColor == colorValue {
					boltList[boltIdx].nuts[nutIdx] = byte(colorIdx) + 0x30
					break
				}
			}
		}
	}

	state := State{
		bolts: boltList,
		moves: make([]UserMove, 0),
	}
	return state
}

func (pState *State) play(numBolts int) bool {
	if pState.gameWon(numBolts) {
		return true
	}

	for boltEndIdx := 0; boltEndIdx < numBolts; boltEndIdx++ {
		boltEnd := pState.bolts[boltEndIdx]
		if boltEnd.nuts[0] == EmptyNut {
			for boltStartIdx := 0; boltStartIdx < numBolts; boltStartIdx++ {
				boltStart := pState.bolts[boltStartIdx]
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

func (pState *State) gameWon(numBolts int) bool {
	win := true
	for boltIdx := 0; boltIdx < numBolts; boltIdx++ {
		bolt := pState.bolts[boltIdx]
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
			return bolt1[nutIdx] > bolt2[nutIdx]
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

func (pState *State) printState(numBolts int) {
	for boltIdx := 0; boltIdx < numBolts; boltIdx++ {
		bolt := pState.bolts[boltIdx]
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
