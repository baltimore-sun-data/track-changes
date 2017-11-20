package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type jsonData struct {
	Id           string     `json:"id"`
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
	data []jsonData
	sync.RWMutex
}

type envelope struct {
	Data  *[]jsonData `json:"data"`
	Meta  interface{} `json:"meta"`
	Error interface{} `json:"error,omitempty"`
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
		log.Printf("Error for %s: %v", j.data.Id, err)
		j.data.LastError = &now
		// Prefer to record web errors over others
		if err != nil {
			j.data.Err = err.Error()
		} else {
			j.data.Err = err2.Error()
		}
	} else {
		j.data.LastError = nil
		j.data.Err = ""
	}

	// Keep old content from being overwritten with err text
	if err == nil && j.data.Content != txt {
		j.data.Content = txt
		j.data.LastChange = now
	}

	if err2 == nil && j.data.Tweet != tweet {
		j.data.Tweet = tweet
		j.data.LastChange = now
	}

	j.data.LastAccessed = now
}

func (a *apiResponse) MarshalJSON() ([]byte, error) {
	a.RLock()
	defer a.RUnlock()
	return json.Marshal(envelope{
		Data: &a.data,
		Meta: struct {
			PollInterval time.Duration `json:"poll_interval"`
		}{dSleep / time.Millisecond},
	})
}

func (a *apiResponse) UnmarshalJSON(b []byte) error {
	a.Lock()
	defer a.Unlock()

	env := envelope{Data: &a.data}
	return json.Unmarshal(b, &env)
}

func (a *apiResponse) jobs() *jobQueue {
	a.RLock()
	defer a.RUnlock()

	jobs := NewJobQueue(len(a.data))
	for i := range a.data {
		val := &a.data[i]
		jobs.push(job{val, val.Url, val.Selector, val.Twitter})
	}

	return jobs
}
