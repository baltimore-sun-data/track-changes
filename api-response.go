package main

import (
	"context"
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
	data   []pageInfo
	cancel context.CancelFunc
	sync.RWMutex
}

type envelope struct {
	Data  *[]pageInfo `json:"data"`
	Meta  interface{} `json:"meta"`
	Error interface{} `json:"error,omitempty"`
}

func (a *apiResponse) updateWeb(pi *pageInfo, url, selector string) {
	now := time.Now()
	txt, err := get(url, selector)

	a.Lock()
	defer a.Unlock()

	pi.LastAccessed = &now
	if err != nil {
		log.Printf("Error for %s: %v", pi.Id, err)
		pi.LastError = &now
		pi.Err = err.Error()
	} else {
		pi.Err = ""
		if pi.Content != txt {
			pi.Content = txt
			pi.LastChange = &now
		}
	}
}

func (a *apiResponse) updateTwitter(pi *pageInfo, twitter string) {
	now := time.Now()
	tweet, err := getTweet(twitter)

	a.Lock()
	defer a.Unlock()

	pi.LastAccessed = &now
	if err != nil {
		log.Printf("Error for %s: %v", pi.Id, err)
		pi.LastError = &now
		pi.TwitterErr = err.Error()
	} else {
		pi.TwitterErr = ""
		if pi.Tweet != tweet {
			pi.Tweet = tweet
			pi.LastChange = &now
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
			jobs.push(job{a, val, val.Url, val.Selector, ""})
		}
		if val.Twitter != "" {
			jobs.push(job{a, val, "", "", val.Twitter})
		}
	}

	return jobs
}
