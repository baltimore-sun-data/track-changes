package main

import (
	"net/url"

	"github.com/ChimeraCoder/anaconda"
)

var twitterApi *anaconda.TwitterApi

func init() {
	consumerKey := MustGetEnv("TWITTER_CONSUMER_KEY")
	consumerSecret := MustGetEnv("TWITTER_CONSUMER_SECRET")
	accessToken := MustGetEnv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret := MustGetEnv("TWITTER_ACCESS_TOKEN_SECRET")

	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	twitterApi = anaconda.NewTwitterApi(accessToken, accessTokenSecret)
}

func getTweet(screenname string) (string, error) {
	if screenname == "" {
		return "", nil
	}

	v := url.Values{}
	v.Set("screen_name", screenname)
	v.Set("include_rts", "true")
	v.Set("exclude_replies", "false")
	v.Set("count", "1")
	ts, err := twitterApi.GetUserTimeline(v)
	if len(ts) > 0 {
		return ts[0].FullText, err
	}

	return "", err
}
