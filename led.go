package main

import (
	"os"
	"sync"
	"time"
)

const (
	LED_STATE_ON  = "1"
	LED_STATE_OFF = "0"
)

type Led struct {
	sync.Mutex
	Device string
}

func (l *Led) Set(state string) error {
	l.Lock()
	defer l.Unlock()

	file, err := os.OpenFile(l.Device, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(state)
	if err != nil {
		return err
	}

	return nil
}

func (l *Led) Blink(count uint8, delay uint) {
	if count == 0 {
		return
	}

	go func() {
		delayMs := time.Duration(delay) * time.Millisecond

		for i := uint8(1); i <= count; i++ {
			l.Set(LED_STATE_ON)
			time.Sleep(delayMs)
			l.Set(LED_STATE_OFF)

			if i < count {
				time.Sleep(delayMs)
			}
		}
	}()
}

func LedsSet(leds []*Led, state string) {
	for _, led := range leds {
		led.Set(state)
	}
}

func LedsBlink(leds []*Led, count uint8, delay uint) {
	go func() {
		delayMs := time.Duration(delay) * time.Millisecond

		for i := uint8(1); i < count; i++ {
			for _, led := range leds {
				led.Set(LED_STATE_ON)
			}
			time.Sleep(delayMs)
			for _, led := range leds {
				led.Set(LED_STATE_OFF)
			}
			if i < count {
				time.Sleep(delayMs)
			}
		}
	}()
}
