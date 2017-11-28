package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type pageInfo struct {
	Id           string     `json:"id"`
	HomePageUrl  string     `json:"homepage_url"`
	Twitter      string     `json:"twitter_screenname"`
	Tweet        string     `json:"last_tweet"`
	DisplayName  string     `json:"display_name"`
	Url          string     `json:"url"`
	Selector     string     `json:"selector"`
	Content      string     `json:"content"`
	LastAccessed *time.Time `json:"last_accessed,omitempty"`
	LastChange   *time.Time `json:"last_change,omitempty"`
	Err          string     `json:"error,omitempty"`
	TwitterErr   string     `json:"twitter_error,omitempty"`
	LastError    *time.Time `json:"last_error,omitempty"`
}

type apiResponse struct {
	data []pageInfo
	sync.RWMutex
}

type envelope struct {
	Data  *[]pageInfo `json:"data"`
	Meta  interface{} `json:"meta"`
	Error interface{} `json:"error,omitempty"`
}

func (a *apiResponse) Update(j job) {
	log.Printf("Updating %#v", j)

	if j.url != "" {
		a.updateWeb(j)
	}

	if j.twitter != "" {
		a.updateTwitter(j)
	}
}

func (a *apiResponse) updateWeb(j job) {
	now := time.Now()
	txt, err := get(j.url, j.selector)

	a.Lock()
	defer a.Unlock()

	j.data.LastAccessed = &now
	if err != nil {
		log.Printf("Error for %s: %v", j.data.Id, err)
		j.data.LastError = &now
		j.data.Err = err.Error()
	} else {
		j.data.Err = ""
		if j.data.Content != txt {
			j.data.Content = txt
			j.data.LastChange = &now
		}
	}
}

func (a *apiResponse) updateTwitter(j job) {
	now := time.Now()
	tweet, err := getTweet(j.twitter)

	a.Lock()
	defer a.Unlock()

	j.data.LastAccessed = &now
	if err != nil {
		log.Printf("Error for %s: %v", j.data.Id, err)
		j.data.LastError = &now
		j.data.TwitterErr = err.Error()
	} else {
		j.data.TwitterErr = ""
		if j.data.Tweet != tweet {
			j.data.Tweet = tweet
			j.data.LastChange = &now
		}
	}
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
		if val.Url != "" {
			jobs.push(job{val, val.Url, val.Selector, ""})
		}
		if val.Twitter != "" {
			jobs.push(job{val, "", "", val.Twitter})
		}
	}

	return jobs
}
