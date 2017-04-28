package checkers

import (
	"os"
)

// DoesFileExist checks if the filename or directory exists in the filesystem
func DoesFileExist(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}

	if !os.IsNotExist(err) {
		return false, err
	}

	return false, nil
}
