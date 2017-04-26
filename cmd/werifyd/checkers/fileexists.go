package checkers

import (
	"os"
)

func DoesFileExist(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	} else {
		if !os.IsNotExist(err) {
			return false, err
		}

		return false, nil
	}
}
