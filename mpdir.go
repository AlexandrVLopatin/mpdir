package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dddpaul/golang-evdev/evdev"
	"github.com/docopt/docopt-go"
)

func main() {
	var err error

	err = initConfig("mpdir")
	if err != nil {
		log.Fatal(err)
	}

	err = watchConfig()
	if err != nil {
		log.Fatal(err)
	}

	usage := `MPD IR Control

Usage:
    mpdir
    mpdir devlist
    mpdir scan
`

	arguments, _ := docopt.Parse(usage, nil, true, "MPD IR Control", false)

	if value, ok := arguments["devlist"]; ok && value == true {
		devlist()
		os.Exit(0)
	}

	if value, ok := arguments["scan"]; ok && value == true {
		scan(irdevice)
		os.Exit(0)
	}

	log.Println("started.")

	err = mpdConnect(mpdHost, mpdPass)
	if err != nil {
		log.Fatal(err)
	}

	LedsBlink([]*Led{&leds[LedBlue], &leds[LedGreen], &leds[LedOrange], &leds[LedWhite]}, 2, 300)

	mpdNoIdle(time.Second * 10)
	listen(irdevice)
}

func devlist() {
	devices, err := evdev.ListInputDevices()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Input devices list:")
	for _, device := range devices {
		fmt.Println("==========================")
		fmt.Printf("Name:    %s\nDevnode: %s\nPhys:    %s\n", device.Name, device.Fn, device.Phys)
	}
}

func scan(device string) {
	var (
		err    error
		dev    *evdev.InputDevice
		events []evdev.InputEvent
	)

	dev, err = evdev.Open(device)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("waiting for events... press ctrl+c to interrupt")

	for {
		events, err = dev.Read()

		for _, ev := range events {
			var codeName string

			code := int(ev.Code)
			evType := int(ev.Type)

			if m, ok := evdev.ByEventType[evType]; ok {
				codeName = m[code]
			}

			if evType == evdev.EV_KEY && ev.Value == 1 {
				fmt.Printf("%d was pressed. codename: %s\n", code, codeName)
			}
		}
	}
}

func volume(c <-chan bool, incVal int) {
	status, err := mpdClient.Status()
	if err != nil {
		return
	}

	vol, err := strconv.Atoi(status["volume"])
	if err != nil {
		return
	}

	for {
		vol += incVal
		if vol < 0 {
			vol = 0
		}
		if vol > 100 {
			vol = 100
		}
		mpdClient.SetVolume(vol)

		select {
		case <-c:
			return
		default:
			leds[LedOrange].Blink(1, 30)
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func listen(device string) {
	var (
		err     error
		dev     *evdev.InputDevice
		events  []evdev.InputEvent
		incCh   chan bool = make(chan bool, 1)
		keylock bool      = true
	)

	defer close(incCh)

	dev, err = evdev.Open(device)
	if err != nil {
		log.Fatal(err)
	}

	for {
		events, err = dev.Read()
		if err != nil {
			log.Fatal(err)
		}

		for _, ev := range events {
			var codeName string

			code := int(ev.Code)
			evType := int(ev.Type)
			if m, ok := evdev.ByEventType[evType]; ok {
				codeName = m[code]
			}

			if code, ok := hotkeys[codeName]; ok {
				if keylock && code != "keylock" {
					leds[LedWhite].Blink(2, 50)
					continue
				}

				if evType == evdev.EV_KEY && ev.Value == 1 {
					leds[LedOrange].Blink(1, 50)

					switch code {
					case "play":
						status, _ := mpdClient.Status()
						pstatus := status["state"]
						switch pstatus {
						case "stop":
							mpdClient.Play(-1)
							leds[LedGreen].Blink(1, 50)
						case "pause":
							mpdClient.Pause(false)
							leds[LedGreen].Blink(1, 50)
						case "play":
							mpdClient.Pause(true)
							leds[LedBlue].Blink(1, 50)
						}
					case "stop":
						mpdClient.Stop()
						leds[LedBlue].Blink(1, 50)
					case "next":
						mpdClient.Next()
					case "previous":
						mpdClient.Previous()
					case "volume_up":
						go volume(incCh, 1)
					case "volume_down":
						go volume(incCh, -1)
					case "random":
						status, _ := mpdClient.Status()
						if status["random"] == "0" {
							mpdClient.Random(true)
							leds[LedGreen].Blink(3, 100)
						} else {
							mpdClient.Random(false)
							leds[LedBlue].Blink(3, 100)
						}
					case "repeat":
						status, _ := mpdClient.Status()
						if status["repeat"] == "0" {
							mpdClient.Repeat(true)
							leds[LedGreen].Blink(3, 100)
						} else {
							mpdClient.Repeat(false)
							leds[LedBlue].Blink(3, 100)
						}
					case "keylock":
						keylock = !keylock
						if keylock {
							leds[LedBlue].Blink(2, 50)
						} else {
							leds[LedGreen].Blink(2, 50)
						}
					case "playlist1", "playlist2", "playlist3", "playlist4", "playlist5",
						"playlist6", "playlist7", "playlist8", "playlist9", "playlist0":
						if playlistName, ok := playlists[code]; ok {
							mpdClient.Clear()
							mpdClient.PlaylistLoad(playlistName, -1, -1)
							mpdClient.Play(-1)
						}
					}
				} else if evType == evdev.EV_KEY && ev.Value == 0 {
					switch code {
					case "volume_up", "volume_down":
						incCh <- true
					}
				}
			}
		}
	}
}
