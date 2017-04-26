package checkers

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const procDir = "/proc"

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

			exe := filepath.Join(procDir, fi.Name(), "exe")
			realExe, err := filepath.EvalSymlinks(exe)
			if err != nil {
				// Process dead?
				continue
			}

			if checkWithPath != "" {
				if checkWithPath == realExe {
					return true, nil
				}
			}
			if checkBasename != "" {
				processName := filepath.Base(realExe)
				if processName == checkBasename {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
