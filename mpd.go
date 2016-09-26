package main

import (
	"github.com/fhs/gompd/mpd"
	"time"
)

var mpdClient *mpd.Client

func mpdConnect(host string, pass string) error {
	var err error

	if pass == "" {
		mpdClient, err = mpd.Dial("tcp", host)
	} else {
		mpdClient, err = mpd.DialAuthenticated("tcp", host, pass)
	}

	if err != nil {
		return err
	} else {
		return nil
	}
}

func mpdNoIdle(sleep time.Duration) {
	go func() {
		for {
			mpdClient.Ping()
			time.Sleep(sleep)
		}
	}()
}
