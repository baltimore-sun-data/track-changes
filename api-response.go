package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type jsonData struct {
	HomePageUrl  string     `json:"homepage_url"`
	Twitter      string     `json:"twitter_screenname"`
	Tweet        string     `json:"last_tweet"`
	DisplayName  string     `json:"display_name"`
	Url          string     `json:"url"`
	Selector     string     `json:"selector"`
	Content      string     `json:"content"`
	LastAccessed time.Time  `json:"last_accessed"`
	LastChange   time.Time  `json:"last_change"`
	Err          string     `json:"error,omitempty"`
	LastError    *time.Time `json:"last_error,omitempty"`
}

type apiResponse struct {
	m map[string]*jsonData
	sync.RWMutex
}

type envelope struct {
	Data  map[string]*jsonData `json:"data"`
	Error interface{}          `json:"error,omitempty"`
}

func (a *apiResponse) Update(j job) {
	log.Printf("Updating %#v", j)
	now := time.Now()

	txt, err := get(j.url, j.selector)
	// This could be done in parallel, but it's not worth the effort
	tweet, err2 := getTweet(j.twitter)

	a.Lock()
	defer a.Unlock()

	if err != nil || err2 != nil {
		log.Printf("Error for %s: %v", j.id, err)
		a.m[j.id].LastError = &now
		// Prefer to record web errors over others
		if err != nil {
			a.m[j.id].Err = err.Error()
		} else {
			a.m[j.id].Err = err2.Error()
		}
	} else {
		a.m[j.id].LastError = nil
		a.m[j.id].Err = ""
	}

	// Keep old content from being overwritten with err text
	if err == nil && a.m[j.id].Content != txt {
		a.m[j.id].Content = txt
		a.m[j.id].LastChange = now
	}

	if err2 == nil && a.m[j.id].Tweet != tweet {
		a.m[j.id].Tweet = tweet
		a.m[j.id].LastChange = now
	}

	a.m[j.id].LastAccessed = now
}

func (a *apiResponse) MarshalJSON() ([]byte, error) {
	a.RLock()
	defer a.RUnlock()
	return json.Marshal(envelope{Data: a.m})
}

func (a *apiResponse) UnmarshalJSON(b []byte) error {
	a.Lock()
	defer a.Unlock()

	env := envelope{Data: a.m}
	return json.Unmarshal(b, &env)
}

func (a *apiResponse) Jobs() []job {
	a.RLock()
	defer a.RUnlock()

	jobs := make([]job, 0, len(a.m))
	for id, val := range a.m {
		jobs = append(jobs, job{id, val.Url, val.Selector, val.Twitter})
	}

	return jobs
}
