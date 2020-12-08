package targetfile

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

type history struct {
	Filepath    string `json:"filepath"`
	ProcessedAt string `json:"processed_at"`
	IsSuccess   bool   `json:"status"`
}

// ListTargetFilePath resourceディレクトリに保存されているファイルのうち
// 処理済みでないファイルの一覧を返す
func ListTargetFilePath() (filePaths []string, err error) {
	all, err := listAllTargetFilePath()
	untreated, err := listTranscribedFilePath()
	if err != nil {
		return nil, err
	}
	for _, filepath := range all {
		if sort.SearchStrings(untreated, filepath) > 0 {
			filePaths = append(filePaths, filepath)
		}
	}
	return
}

func listAllTargetFilePath() (filePaths []string, err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	inputPath := filepath.Join(pwd, "resource", "input")
	err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Failed to access the filepath ->", path, ":", err)
			return err
		}
		if info.IsDir() && info.Name() == ".gitkeep" {
			return filepath.SkipDir
		}
		filePaths = append(filePaths, path)
		return nil
	})
	log.Print("File:", filePaths)
	return
}

func listTranscribedFilePath() (filePaths []string, err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	historyFilePath := filepath.Join(pwd, "resource", "history", "history.json")
	data, err := ioutil.ReadFile(historyFilePath)
	if err != nil {
		return nil, err
	}
	var history []history
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}
	for _, h := range history {
		if h.IsSuccess {
			filePaths = append(filePaths, h.Filepath)
		}
	}
	sort.StringSlice(filePaths).Sort()
	return
}
