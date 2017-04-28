package checkers

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const procDir = "/proc"

// IsProcessRunning checks if the process is running
func IsProcessRunning(checkBasename, checkWithPath string) (bool, error) {
	// Assuming Linux
	dir, err := os.Open(procDir)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	for {
		pids, err := dir.Readdir(100)
		if err == io.EOF {
			break
		}

		for _, fi := range pids {
			if !fi.IsDir() {
				continue
			}
			_, err := strconv.ParseInt(fi.Name(), 10, 64)
			if err != nil {
				// Not a pid-dir
				continue
			}

			// This file is supposed to be readable by all users
			cmdlineFile := filepath.Join(procDir, fi.Name(), "cmdline")

			cmdline, err := ioutil.ReadFile(cmdlineFile)
			if err != nil {
				// Process dead?
				continue
			}
			idx := bytes.Index(cmdline, []byte{0})
			if idx < 1 {
				// No NUL-byte in cmdline... This should not happen
				continue
			}

			command := string(cmdline[:idx])

			if checkWithPath != "" {
				if checkWithPath == command {
					return true, nil
				}
			}
			if checkBasename != "" {
				processName := filepath.Base(command)
				if processName == checkBasename {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
