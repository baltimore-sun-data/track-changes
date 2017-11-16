package main

import (
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
)

var twitterApi *anaconda.TwitterApi

func init() {
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	twitterApi = anaconda.NewTwitterApi(accessToken, accessTokenSecret)
}

func getTweet(screenname string) (string, error) {
	v := url.Values{}
	v.Set("screen_name", screenname)
	v.Set("include_rts", "1")
	v.Set("count", "1")
	ts, err := twitterApi.GetUserTimeline(v)
	if len(ts) > 0 {
		return ts[0].FullText, err
	}

	return "", err
}
