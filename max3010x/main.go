package main

import (
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

	t := time.NewTicker(500 * time.Millisecond)

	for {
		temp, err := sensor.Temperature()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\rtemp = %02.2f ", temp)
		<-t.C
	}

	/*

		// reset
		reg := make([]byte, 1)
		if err := d.Tx([]byte{maxModeCfg}, reg); err != nil {
			log.Fatal(fmt.Errorf("max30100: could not read register: %w", err))
		}
		if _, err := d.Write([]byte{maxModeCfg, reg[0] & 0x40}); err != nil {
			log.Fatal(fmt.Errorf("max30100: could not configure mode: %w", err))
		}
		// start
		reg = make([]byte, 1)
		if err := d.Tx([]byte{maxModeCfg}, reg); err != nil {
			log.Fatal(fmt.Errorf("max30100: could not read register: %w", err))
		}
		if _, err := d.Write([]byte{maxModeCfg, reg[0] & 0x7F}); err != nil {
			log.Fatal(fmt.Errorf("max30100: could not configure mode: %w", err))
		}

		// begin
		if _, err := d.Write([]byte{maxModeCfg, 0x02}); err != nil {
			log.Fatal(fmt.Errorf("max30100: could not configure mode: %w", err))
		}
		if _, err := d.Write([]byte{maxLedCfg, mA500}); err != nil {
			log.Fatal(fmt.Errorf("max30100: could not configure LED: %w", err))
		}
		if _, err := d.Write([]byte{maxSpO2Cfg, (sr100 << 2) | pw1600}); err != nil {
			log.Fatal(fmt.Errorf("max30100: could not configure SpO2: %w", err))
		}

		read := func() {
			// readSensor
			write := []byte{maxFifoData}
			read := make([]byte, 4)
			if err := d.Tx(write, read); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("read = %#x\n", read)

			ir := (uint16(read[0]) << 8) | uint16(read[1])
			red := (uint16(read[2]) << 8) | uint16(read[3])

			fmt.Printf("ir = %+v\n", ir)
			fmt.Printf("red = %+v\n", red)
		}

		//t := time.NewTicker(10 * time.Millisecond)
		read()
	*/
}
