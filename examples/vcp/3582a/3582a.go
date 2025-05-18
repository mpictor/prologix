// Copyright (c) 2020–2024 The prologix developers. All rights reserved.
// Project site: https://github.com/gotmc/prologix
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gotmc/prologix"
	"github.com/gotmc/prologix/lib/find"
	"github.com/soypat/cereal"
)

func main() {
	var (
		serialPort string
		delay      time.Duration
		sdelay     time.Duration
		gpibPAD    int
		gpibSAD    int
		verbose    bool
	)
	log.SetFlags(log.Lmicroseconds)
	tty, finderr := find.Find(find.ArduinoFilter)
	if finderr != nil {
		// log.Printf("%s: guessing ttyACM0", finderr)
		tty = "ttyACM0"
	}

	// Get Virtual COM Port (VCP) serial port for Prologix.
	flag.StringVar(
		&serialPort,
		"port",
		"/dev/"+tty,
		"Serial port for Prologix VCP GPIB controller",
	)
	flag.DurationVar(&delay, "delay", time.Second/10, "delay before ++ cmds")
	flag.DurationVar(&sdelay, "sdelay", 30*time.Second, "serial port timeout")

	flag.IntVar(&gpibPAD, "pad", 11, "GPIB primary address for the HP 3582A")
	flag.IntVar(&gpibSAD, "sad", 255, "GPIB secondary address for the HP 3582A")

	flag.BoolVar(&verbose, "v", false, "increase verbosity")

	// Parse the flags
	flag.Parse()

	if finderr != nil && serialPort == "/dev/ttyACM0" {
		// only print this if the port isn't overridden via flag
		log.Printf("%s: guessing %s", finderr, serialPort)
	}

	cimpl := cereal.Tarm{}
	port, err := cimpl.OpenPort(serialPort, cereal.Mode{
		BaudRate:    115200,
		ReadTimeout: sdelay,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new GPIB controller using the aforementioned serial port
	// communicating with the instrument at the given GPIB address.
	opts := []prologix.ControllerOption{
		prologix.WithWriteDelay(delay),
	}
	if gpibSAD != 0xff {
		opts = append(opts, prologix.WithSecondaryAddress(gpibSAD))
	}
	// AR488 does not like CLR so use false
	gpib, err := prologix.NewController(port, gpibPAD, false, opts...)
	if err != nil {
		log.Fatal(err)
	}
	if verbose {
		if gpibSAD == 0xff {
			log.Printf("GPIB address = %d", gpibPAD)
		} else {
			log.Printf("GPIB address = %d:%d", gpibPAD, gpibSAD)
		}
	}
	defer func() {
		// Return local control to the front panel.
		err := gpib.FrontPanel(true)
		if err != nil {
			log.Fatalf("error setting local control for front panel: %s", err)
		}

		// Discard any unread data on the serial port and then close.
		err = port.Close()
		if err != nil {
			log.Printf("error closing serial port: %s", err)
		}
	}()
	exerciseInst(gpib, verbose)
}

func listAlphanumeric(gpib *prologix.Controller) []string {
	// lan returns 128 chars, to be split into 4 lines
	txt := gpib.MustQuery("lan")

	if len(txt) < 96 {
		log.Fatalf("short response %d %s", len(txt), txt)
	}
	return []string{txt[:32], txt[32:64], txt[64:96], txt[96:]}
}

func listDataset(gpib *prologix.Controller) ([]float32, string) {
	// lds returns comma-separated floats
	data := gpib.MustQuery("lds")
	n := strings.Count(data, ",") + 1
	log.Printf("%d data entries", n)
	floats := make([]float32, 0, n)
	remain := data
	found := true
	elem := ""
	elem, remain, found = strings.Cut(remain, ",")
	for found {
		f, err := strconv.ParseFloat(elem, 32)
		if err != nil {
			log.Fatal(err)
		}
		floats = append(floats, float32(f))
		elem, remain, found = strings.Cut(remain, ",")
	}
	return floats, remain
}

func exerciseInst(gpib *prologix.Controller, verbose bool) {
	// q := func(c string) string { return gpib.MustQuery(c) }

	gpib.Debug = true

	lines := listAlphanumeric(gpib)
	log.Printf("text:\n%s\n%s\n%s\n%s", lines[0], lines[1], lines[2], lines[3])

	floats, remain := listDataset(gpib)
	log.Printf("data: %v\n%s", floats, remain)

	// read memory regions
	// NOTE LFM takes longer than I'd expect, 26-27s for each of the following
	// length of result is 2x 2nd arg

	dis := gpib.MustQuery("LFM,74000,512")
	log.Printf("display: %d\n%x", len(dis), dis) // 1024 bc00a0008000b400bc...

	dis = gpib.MustQuery("lfm,70000,1024")
	log.Printf("time record: %d\n%x", len(dis), dis) // 2048 10500f100f9010900fa00f001...

	dis = gpib.MustQuery("lfm,77454,5")
	log.Printf("front panel switches: %d\n%x", len(dis), dis) // 10 2564108100fe010000b8

	log.Print("done")
}

// TODO query funcs that know the amount of data expected and wait for it?
