package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type jsonData struct {
	Url          string     `json:"url"`
	Selector     string     `json:"selector"`
	Content      string     `json:"content"`
	LastAccessed time.Time  `json:"last_accessed"`
	LastChange   time.Time  `json:"last_change"`
	Err          string     `json:"error,omitempty"`
	LastError    *time.Time `json:"last_error,omitempty"`
}

type apiResponse struct {
	m map[string]jsonData
	sync.RWMutex
}

func (a *apiResponse) Update(j job) error {
	log.Printf("Updating %#v", j)

	txt, err := get(j.url, j.selector)
	now := time.Now()
	var (
		errStr  string
		errTime *time.Time
	)
	if err != nil {
		errStr = err.Error()
		errTime = &now
		log.Printf("Error for %s: %v", j.id, err)
	}

	a.Lock()
	defer a.Unlock()

	lastChange := a.m[j.id].LastChange
	if a.m[j.id].Content != txt {
		lastChange = now
	}

	a.m[j.id] = jsonData{
		Url:          j.url,
		Selector:     j.selector,
		Content:      txt,
		LastAccessed: now,
		LastChange:   lastChange,
		Err:          errStr,
		LastError:    errTime,
	}

	return err
}

func (a *apiResponse) MarshalJSON() ([]byte, error) {
	a.RLock()
	defer a.RUnlock()
	return json.Marshal(&a.m)
}
