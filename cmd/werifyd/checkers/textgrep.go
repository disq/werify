package checkers

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

// DoesFileHasWord reads a file line by line and checks if the line contains the word
func DoesFileHasWord(filename, word string) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if len(word) == 0 {
		return false, errors.New("Check pattern is empty")
	}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, word) {
			return true, nil
		}
	}

	return false, s.Err()
}
