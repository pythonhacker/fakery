// Data helpers module
package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const prefix = "locales"

type JSONDataLoader struct {
	Locale   string              `json:"locale"`
	Filename string              `json:"filename"`
	Data     map[string][]string `json:"data",omitempty"`
	CacheKey string              `json:"cache_key,omitempty"`
}

type WeightedItem struct {
	Item   string  `json:"item"`
	Weight float64 `json:"weight"`
}

type WeightedArray struct {
	Items []WeightedItem
}

func (w WeightedArray) Validate() (bool, float64) {
	// weights should add to 1.0
	var cumWeight float64 = 0.0

	for _, item := range w.Items {
		cumWeight += item.Weight
	}

	cumWeight = math.Round(cumWeight*100) / 100

	if cumWeight == 1.0 {
		return true, cumWeight
	}
	return false, cumWeight
}

// GetModuleFile returns the full path to a file relative to the module directory
func getModuleFile(relativePath string) (string, error) {
	// Get the current file's directory
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrNotExist
	}

	// Get the directory of the current file (helpers.go)
	dir := filepath.Dir(currentFile)

	// Construct the full path to the target file
	fullPath := filepath.Join(dir, relativePath)

	// Verify the file exists (optional)
	_, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}

	return fullPath, nil
}

// Fetch value of key
func (l *JSONDataLoader) Get(key string) []string {

	var val []string
	var exists bool

	if val, exists = l.Data[key]; !exists {
		fmt.Printf("doesnt exist %s\n", key)
		return nil
	}
	return val
}

func (l *JSONDataLoader) Load(locale, fileName string) error {

	var err error
	var data []byte

	relPath := path.Join(prefix, locale, fileName)
	filePath, err := getModuleFile(relPath)
	if err != nil {
		return err
	}

	if _, err = os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return err
	}

	data, err = os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, &l.Data); err != nil {
		return err
	}

	l.Filename = filePath
	l.Locale = locale

	return nil
}

// Convert a string to a weighted array
func (l *JSONDataLoader) GetWeightedArray(key, sep string) (*WeightedArray, error) {

	var dataArray WeightedArray

	dataItems := l.Get(key)

	for _, dataItem := range dataItems {
		items := strings.Split(dataItem, sep)
		weight, err := strconv.ParseFloat(items[1], 64)
		if err != nil {
			return nil, err
		}
		dataArray.Items = append(dataArray.Items,
			WeightedItem{Item: items[0], Weight: weight})

	}

	if ok, val := dataArray.Validate(); !ok {
		return nil, fmt.Errorf("weighted array[key: %s] didn't validate - weight %.2f", key, val)
	}

	return &dataArray, nil
}
