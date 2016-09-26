package main

import (
	"log"
	"os"
	"os/user"

	"github.com/howeyc/fsnotify"
	"github.com/spf13/viper"
)

const (
	LedOrange = 0
	LedGreen  = 1
	LedBlue   = 2
	LedWhite  = 3
)

var (
	irdevice  string
	mpdHost   string
	mpdPass   string
	hotkeys   map[string]string
	playlists map[string]string
	leds      [4]Led
)

func watchConfig() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsModify() {
					log.Println("reloading config file")
					loadConfig()
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch(viper.ConfigFileUsed())
	if err != nil {
		return err
	}

	return nil
}

func initConfig(fn string) error {
	viper.SetConfigName(fn)

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	viper.AddConfigPath(pwd)

	usr, err := user.Current()
	if err != nil {
		return err
	}
	viper.AddConfigPath(usr.HomeDir)
	viper.AddConfigPath("/etc")

	return loadConfig()
}

func loadConfig() error {
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	irdevice = viper.GetString("irdevice")
	mpdHost = viper.GetString("mpdhost")
	mpdPass = viper.GetString("mpdpass")
	hotkeys = viper.GetStringMapString("hotkeys")
	playlists = viper.GetStringMapString("playlists")

	ledDevices := viper.GetStringMapString("leds")
	leds[LedOrange] = Led{Device: ledDevices["orange"]}
	leds[LedGreen] = Led{Device: ledDevices["green"]}
	leds[LedBlue] = Led{Device: ledDevices["blue"]}
	leds[LedWhite] = Led{Device: ledDevices["white"]}

	return nil
}
