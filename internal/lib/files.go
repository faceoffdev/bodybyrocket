package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

const gb = 1 << 30

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

func RemoveFiles(paths ...string) {
	for _, path := range paths {
		if path != "" {
			_ = os.Remove(path)
		}
	}
}

func DownloadFile(url, filepath string) (err error) {
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	totalSize, err := strconv.Atoi(response.Header.Get("Content-Length"))
	if err != nil {
		return
	}

	if float64(totalSize) > 1.9*gb {
		return fmt.Errorf("файл слишком большой (%.2f Гб)", float64(totalSize)/gb)
	}

	outFile, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, response.Body)

	return
}
