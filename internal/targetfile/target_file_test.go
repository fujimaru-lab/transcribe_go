package targetfile

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAllTargetFilePath(t *testing.T) {
	t.Run("normal case - list all files in target folder", func(t *testing.T) {
		assert := assert.New(t)

		config, _ := loadConfig()
		inputFilepath := config.Resource.InputDirPath
		file01 := filepath.Join(inputFilepath, "testfile01.txt")
		file02 := filepath.Join(inputFilepath, "testfile02.txt")
		file03 := filepath.Join(inputFilepath, "testfile03.txt")
		os.Create(file01)
		os.Create(file02)
		os.Create(file03)
		defer os.Remove(file01)
		defer os.Remove(file02)
		defer os.Remove(file03)

		actual, _ := ListAllTargetFilePath()
		sort.Sort(sort.StringSlice(actual))
		assert.Equal(file01, actual[0])
		assert.Equal(file02, actual[1])
		assert.Equal(file03, actual[2])
	})

}

func TestLoadConfig(t *testing.T) {
	t.Run("normal case - get config", func(t *testing.T) {
		assert := assert.New(t)

		config, _ := loadConfig()
		assert.Equal("C:/Users/yoshi/go/src/github.com/fujimaru-lab/transcribe_go/resource/history/history.json",
			config.HistoryFilepath,
			"should get history.json path")
		assert.Equal("C:/Users/yoshi/go/src/github.com/fujimaru-lab/transcribe_go/resource/input/",
			config.Resource.InputDirPath,
			"should get input_file_path")
		assert.Equal("C:/Users/yoshi/go/src/github.com/fujimaru-lab/transcribe_go/resource/output/",
			config.Resource.OutputDirPath,
			"should get output_file_path")
	})
}

func TestLoadHstry(t *testing.T) {
	t.Run("normal case - get history.json", func(t *testing.T) {
		assert := assert.New(t)
		config, _ := loadConfig()
		befData, _ := ioutil.ReadFile(config.HistoryFilepath)

		h1 := history{
			Filepath:    "test/code/file01.txt",
			ProcessedAt: "2020-12-09T12:30:555",
			IsSuccess:   true,
		}
		h2 := history{
			Filepath:    "test/code/file02.txt",
			ProcessedAt: "2021-01-09T12:30:555",
			IsSuccess:   false,
		}
		h3 := history{
			Filepath:    "test/code/file03.txt",
			ProcessedAt: "2020-12-10T12:30:555",
			IsSuccess:   true,
		}
		history := []history{h1, h2, h3}
		data, _ := json.Marshal(history)
		os.Remove(config.HistoryFilepath)
		os.Create(config.HistoryFilepath)
		ioutil.WriteFile(config.HistoryFilepath, data, os.ModeDevice)
		defer ioutil.WriteFile(config.HistoryFilepath, befData, os.ModeDevice)
		defer os.Create(config.HistoryFilepath)

		actual, _ := loadHstry()
		assert.Contains(history, actual[0])
		assert.NotContains(history, actual[1])
		assert.Contains(history, actual[1])
	})
}
