package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const NutsPerBolt = 4
const EmptyStartingBolts = 2

type State struct {
	nuts [][]byte
}

func main() {
	initialState := loadFile("test.nuts")
	initialState.printState()
}

func (s *State) printState() {
	for _, bolt := range s.nuts {
		zero := [...]byte{0}
		stringBolt := string(bytes.ReplaceAll(bolt, zero[:], []byte("_")))
		fmt.Printf("%s\t", stringBolt)
	}
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
	state2d := make([][]byte, len(inputLine)/NutsPerBolt)
	for bolt, nut := 0, 0; bolt < len(state2d); bolt, nut = bolt+1, nut+NutsPerBolt {
		state2d[bolt] = inputLine[nut : nut+NutsPerBolt]
	}
	//for i := 0; i*4 < len(inputLine); i++ {
	//	state2d[i] = inputLine[i*4 : i*4+4]
	//}
	state := State{
		nuts: state2d,
	}
	return state
}
