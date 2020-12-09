package targetfile

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/fujimaru-lab/transcribe_go/internal/constant"
)

type config struct {
	Resource        resource `json:"resource"`
	HistoryFilepath string   `json:"history_filepath"`
}

type resource struct {
	InputDirPath  string `json:"input_dir_path"`
	OutputDirPath string `json:"output_dir_path"`
}

type history struct {
	Filepath    string `json:"filepath"`
	ProcessedAt string `json:"processed_at"`
	IsSuccess   bool   `json:"status"`
}

// ListTargetFilePath resourceディレクトリに保存されているファイルのうち
// 処理済みでないファイルの一覧を返す
func ListTargetFilePath(all, treated []string) (trgtFilePath []string, err error) {
	if err != nil {
		return nil, err
	}
	for _, trgt := range all {
		isContain := false
		for _, fltr := range treated {
			isContain = trgt == fltr
		}
		if isContain {
			trgtFilePath = append(trgtFilePath, trgt)
		}
	}
	return
}

// ListAllTargetFilePath inputディレクトリに存在する全てのファイルパス配列を取得する
func ListAllTargetFilePath() (filePaths []string, err error) {
	config, err := loadConfig()
	if err != nil {
		return
	}
	inputPath := config.Resource.InputDirPath
	err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Failed to access the filepath ->", path, ":", err)
			return err
		} else if info.IsDir() || info.Name() == ".gitkeep" {
			return nil
		} else {
			filePaths = append(filePaths, path)
			return nil
		}
	})
	return
}

// ListTranscribedFilePath 処理結果履歴ファイルから処理済みのファイルパス配列を取得する
func ListTranscribedFilePath() (filePaths []string, err error) {
	historyFilePath := "history file path"
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

func loadConfig() (config config, err error) {
	var configFilePath string
	err = filepath.Walk(constant.RootDir, func(path string, info os.FileInfo, err error) error {
		base := filepath.Base(path)
		if isConfig, _ := filepath.Match("config.json", base); isConfig {
			configFilePath, _ = filepath.Abs(path)
		}
		return nil
	})
	if configFilePath == "" {
		err = errors.New("config.json is not exists")
		return
	}
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &config)
	return
}

func loadHstry() (hstry []history, err error) {
	var tmp []history
	config, _ := loadConfig()
	data, _ := ioutil.ReadFile(config.HistoryFilepath)
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return
	}

}
