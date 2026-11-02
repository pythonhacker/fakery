// Decoded data of a single (resource, locale) pair
package fakery

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// LocaleData holds the decoded contents of one resource for one locale.
// Values are created and filled by a Registry and are immutable once
// returned, so they can be shared freely between goroutines. Slices and
// maps handed out by its accessors are shared and must not be modified.
type LocaleData struct {
	locale   string
	fileName string
	// Most data is string array
	data map[string][]string
	// Some data is a map e.g: currency without specific top level keys
	dataMap []map[string]string
	isMap   bool
	// Error encountered while loading, if any
	err error
	// Lazily decoded typed views of the data, keyed by reflect.Type
	views sync.Map
}

// A lazily computed typed view of a LocaleData.
type dataView struct {
	once  sync.Once
	value any
}

// Locale returns the locale the data was read from. When a fallback was
// used this differs from the requested locale.
func (l *LocaleData) Locale() string {
	return l.locale
}

// FileName returns the resource name of the data.
func (l *LocaleData) FileName() string {
	return l.fileName
}

// IsMap reports whether the data is map based (an array of string maps).
func (l *LocaleData) IsMap() bool {
	return l.isMap
}

// Err returns the error encountered while loading the data, if any.
func (l *LocaleData) Err() error {
	return l.err
}

// Keys returns the sorted top level keys of list based data.
func (l *LocaleData) Keys() []string {
	keys := make([]string, 0, len(l.data))
	for key := range l.data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Len returns the number of keys of list based data or the number of
// items of map based data.
func (l *LocaleData) Len() int {
	if l.isMap {
		return len(l.dataMap)
	}
	return len(l.data)
}

// Fetch value of key
func (l *LocaleData) Get(key string) []string {

	var val []string
	var exists bool

	if val, exists = l.data[key]; !exists {
		log.Printf("error - key doesnt exist %s\n", key)
		return nil
	}
	return val
}

// Fetch random data item from map. Returns nil when there are no items.
func (l *LocaleData) RandomWeightedItem(f *Fakery) map[string]string {
	if len(l.dataMap) == 0 {
		return nil
	}
	return l.dataMap[f.IntRange(len(l.dataMap))]
}

// Load is kept for compatibility. Data obtained from a Registry is always
// loaded already, so Load only reports the load error, if any.
func (l *LocaleData) Load() error {
	if l.fileName == "" || l.locale == "" {
		return fmt.Errorf("error - loader not initialized")
	}
	return l.err
}

// Decode converts the data into v (a pointer to a struct, map or slice)
// using its JSON representation.
func (l *LocaleData) Decode(v any) error {
	var src any = l.data
	if l.isMap {
		src = l.dataMap
	}

	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

// Given a data key, fetch its value and parse it into a weighted array
func (l *LocaleData) GetWeightedArray(key, sep string) (*WeightedArray, error) {

	var dataArray WeightedArray

	dataItems := l.Get(key)

	for _, dataItem := range dataItems {
		items := strings.Split(dataItem, sep)
		if len(items) < 2 {
			return nil, fmt.Errorf("weighted item %q[key: %s] has no weight", dataItem, key)
		}
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

// viewOf returns the data decoded into a T. The conversion is done once
// per LocaleData and type, and the result is shared by all callers.
func viewOf[T any](l *LocaleData) *T {
	slot, _ := l.views.LoadOrStore(reflect.TypeFor[T](), &dataView{})
	view := slot.(*dataView)
	view.once.Do(func() {
		value := new(T)
		if err := l.Decode(value); err != nil {
			log.Printf("error - decoding %s/%s into %T: %v\n", l.locale, l.fileName, value, err)
		}
		view.value = value
	})
	return view.value.(*T)
}
