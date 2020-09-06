package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cgxeiji/max3010x"
	"github.com/cgxeiji/max3010x/max30102"
)

func main() {
	sensor, err := max3010x.New()
	if err != nil {
		log.Fatal(err)
	}
	defer sensor.Close()

	switch sensor.PartID {
	case 0x11:
		fmt.Printf("MAX30100 rev.%d detected\n", sensor.RevID)
	case max30102.PartID:
		fmt.Printf("MAX30102 rev.%d detected\n", sensor.RevID)
	}
	fmt.Println("------------------------------")

	hrCh := make(chan float64)
	tempCh := make(chan float64)

	go func() {
		for {
			t := time.NewTicker(200 * time.Millisecond)
			hr, err := sensor.HeartRate()
			if errors.Is(err, max3010x.ErrNotDetected) {
				hr = 0
			} else if err != nil {
				log.Fatal(err)
			}
			select {
			case hrCh <- hr:
			}
			<-t.C
		}
	}()

	go func() {
		for {
			t := time.NewTicker(1 * time.Second)
			temp, err := sensor.Temperature()
			if err != nil {
				log.Fatal(temp)
			}
			select {
			case tempCh <- temp:
			}
			<-t.C
		}
	}()

	rawCh := make(chan []float64)
	go func() {
		for {
			red, ir, err := sensor.Leds()
			if err != nil {
				log.Fatal(err)
			}
			red -= 0.30
			red *= 300
			ir -= 0.30
			ir *= 300
			if red < 0 {
				red = 0
			}
			if ir < 0 {
				ir = 0
			}
			select {
			case rawCh <- []float64{red, ir}:
			}
		}
	}()

	spO2Ch := make(chan float64)
	go func() {
		t := time.NewTicker(200 * time.Millisecond)
		for {
			spO2, err := sensor.SpO2()
			if errors.Is(err, max3010x.ErrNotDetected) {
				spO2 = 0
			} else if err != nil {
				log.Fatal(err)
			}
			select {
			case spO2Ch <- spO2:
			}
			<-t.C
		}
	}()

	t := time.NewTicker(50 * time.Millisecond)

	fmt.Printf("\n\n\n\n\n")

	temp := 0.0
	hr := 0.0
	spO2 := 0.0
	raw := make([]float64, 2)

	for {
		select {
		case temp = <-tempCh:
		case hr = <-hrCh:
		case spO2 = <-spO2Ch:
		case raw = <-rawCh:
		}
		fmt.Printf("\033[5F")
		fmt.Printf("sensor temp\t: %2.1fC        \n", temp)
		if hr == 0 {
			fmt.Printf("heart rate\t: --             \n")
		} else {
			fmt.Printf("heart rate\t: %3.2fbpm       \n", hr)
		}
		if spO2 == 0 {
			fmt.Printf("SpO2\t\t: --                   \n")
		} else {
			fmt.Printf("SpO2\t\t: %3.2f%%              \n", spO2)
		}
		fmt.Printf("red LED\t\t: %s              \n", float2bar(raw[0]))
		fmt.Printf("IR LED\t\t: %s               \n", float2bar(raw[1]))
		<-t.C
	}
}

func float2bar(n float64) string {
	block := []string{"", "▏", "▏", "▎", "▍", "▌", "▋", "▊", "▊", "▉"}
	t := int(n)
	s := ""
	f := int((n - float64(t)) * 10)

	for i := 0; i < t; i++ {
		s += "█"
	}
	s += block[f]

	return s
}
