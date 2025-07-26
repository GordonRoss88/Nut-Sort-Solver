package Nut_Sort_Solver

import (
	"io"
	"log"
	"nuts/hello"
	"os"
	"strings"
)

func main() {
	initialState := loadFile("test.nuts")
	log.Println(string(initialState))
	hello.Hello()
}

func loadFile(path string) []byte {
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
	state := []byte(inputString)
	twoBolts := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
	state = append(state, twoBolts[:]...)
	return state
}
