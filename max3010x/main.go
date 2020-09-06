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

	// Check which sensor is connected using the PartID.
	switch sensor.PartID {
	case 0x11: // TODO: test with MAX30100
		fmt.Printf("MAX30100 rev.%d detected\n", sensor.RevID)
	case max30102.PartID:
		fmt.Printf("MAX30102 rev.%d detected\n", sensor.RevID)
	}
	fmt.Println("---------------------------")

	// Read the heart rate every 200ms.
	hrCh := make(chan float64)
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

	// Read the SpO2 every 200ms.
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

	// Read the sensor's temperature every second.
	tempCh := make(chan float64)
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

	// Access the underlying device for low level functions.
	// Read raw LED values as fast as possible.
	rawCh := make(chan []float64)
	go func() {
		device, err := sensor.ToMax30102()
		if errors.Is(err, max3010x.ErrWrongDevice) {
			fmt.Println("device is not MAX30102")
			return
		} else if err != nil {
			log.Fatal(err)
		}
		for {
			ir, red, err := device.IRRed()
			if err != nil {
				log.Fatal(err)
			}

			// Adjusting raw value for visualization
			ir -= 0.30
			ir *= 300
			if ir < 0 {
				ir = 0
			}

			// Adjusting raw value for visualization
			red -= 0.30
			red *= 300
			if red < 0 {
				red = 0
			}
			select {
			case rawCh <- []float64{red, ir}:
			}
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
