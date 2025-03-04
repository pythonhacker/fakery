// Data helpers module
package gofakelib

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
	"sync"
)

const prefix = "data/locales"

type LocaleData struct {
	once     sync.Once           `json:"-"`
	locale   string              `json:"locale"`
	fileName string              `json:"filename"`
	data     map[string][]string `json:"data",omitempty"`
	loaded   bool                `json:"loaded"`
	cacheKey string              `json:"cache_key,omitempty"`
}

type DataLoader struct {
	localeDataMap map[string]LocaleData
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

func (loader *DataLoader) Init(locale, filePath string) *LocaleData {
	var localeData LocaleData

	if len(loader.localeDataMap) == 0 {
		loader.localeDataMap = make(map[string]LocaleData)
	}

	localeData.fileName = filePath
	localeData.locale = locale
	loader.localeDataMap[locale] = localeData

	return &localeData
}

// Return the specific data pointer for the given locale
func (loader *DataLoader) Get(locale string) *LocaleData {
	var val LocaleData
	var ok bool

	if val, ok = loader.localeDataMap[locale]; ok {
		return &val
	}
	return nil
}

// Fetch value of key
func (l *LocaleData) Get(key string) []string {

	var val []string
	var exists bool

	if val, exists = l.data[key]; !exists {
		fmt.Printf("doesnt exist %s\n", key)
		return nil
	}
	return val
}

// Load the data for the given locale and fileName
func (l *LocaleData) load() error {

	var err error
	var data []byte

	relPath := path.Join(prefix, l.locale, l.fileName)
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

	if err = json.Unmarshal(data, &l.data); err != nil {
		return err
	}

	return nil
}

// Do the loading once only for every locale
func (l *LocaleData) Load() error {

	if l.fileName == "" || l.locale == "" {
		return fmt.Errorf("error - loader not initialized")
	}

	var err error

	l.once.Do(func() {
		err = l.load()
		l.loaded = true
	})

	return err
}

func (loader *DataLoader) EnsureLoaded(locale string) (*LocaleData, error) {
	localeData := loader.Get(locale)
	if localeData == nil {
		return nil, fmt.Errorf("locale data not initialized for locale %s", locale)
	}

	if !localeData.loaded {
		err := localeData.Load()
		return localeData, err
	}

	return localeData, nil
}

// Convert a string to a weighted array
func (l *LocaleData) GetWeightedArray(key, sep string) (*WeightedArray, error) {

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
