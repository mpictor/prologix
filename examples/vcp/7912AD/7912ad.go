// Copyright (c) 2020–2024 The prologix developers. All rights reserved.
// Project site: https://github.com/gotmc/prologix
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package main

// TODO https://github.com/charmbracelet/log

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gotmc/prologix"
	"github.com/gotmc/prologix/lib/analyzer"
	"github.com/gotmc/prologix/lib/cmdlog"
	"github.com/gotmc/prologix/lib/connutil"
	"github.com/gotmc/prologix/lib/tek"
)

func main() {
	log.SetFlags(log.Lmicroseconds)

	var maxTime time.Duration

	c := connutil.Conn{}
	c.AddFlags()
	flag.DurationVar(&maxTime, "max", 30*time.Second, "max runtime before exiting")

	var la analyzer.GusmanB
	la.AddFlags()

	// Parse the flags
	flag.Parse()

	gpib, cleanup, err := c.Setup(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	var wg sync.WaitGroup
	var laErr error
	if la.CheckFlags() {
		// activate
		wg.Add(1)
		go func() {
			defer wg.Done()
			// startSigrokCapture(rate, samples)
			laErr = la.Capture()
		}()
		time.Sleep(time.Millisecond * 60)
	}

	done := make(chan interface{})
	wg.Add(1)
	go func() { run(gpib); wg.Done() }()
	go func() {
		wg.Wait()
		if laErr != nil {
			log.Fatalf("logic analyzer failed: %s", laErr)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(maxTime):
		log.Printf("exiting after %s", maxTime)
		os.Exit(1)
	}
}

func run(gpib *prologix.Controller) {
	defer log.Print("run complete, exiting...")
	// defer gpib.FrontPanel(true)

	// Send the Selected Device Clear (SDC) message
	log.Println("Sending the Selected Device Clear (SDC) message")
	err := gpib.ClearDevice()
	if err != nil {
		log.Printf("error clearing device: %s", err)
	}
	a, b, err := gpib.InstrumentAddress()
	if err != nil {
		log.Printf("error: %s", err)
	}
	log.Printf("addr: %d:%d", a, b)

	query, bquery, cmd := cmdlog.PrettyFuncs(gpib)

	pquery := func(q string) []uint16 {
		r := query(q)
		r = strings.TrimSpace(r)
		p, err := tek.Unpack([]byte(r))
		if err != nil {
			log.Printf("error: %s", err)
			if len(p) == 0 {
				return nil
			}
		}
		l := fmt.Sprintf("%d points, min % 5d max % 5d", len(p), slices.Min(p), slices.Max(p))
		// d:=fmt.Sprintf("")
		var ns = make([]string, 0, len(p))
		for _, n := range p {
			ns = append(ns, fmt.Sprintf("% 5d", n))
		}
		log.Printf("%s\n%s", cmdlog.R1Style.Render(l), cmdlog.R2Style.Render(strings.Join(ns, " ")))
		return p
	}

	// TODO save each run's raw data as timestamped json file?

	// FIXME on some runs we see one command's response returned for the _next_ command
	// will flush fix??

	// cmd("")
	time.Sleep(time.Millisecond * 100)
	// gpib.Flush() // fixme make this impl actual interface reset command

	// NOTE ignores lowercase commands

	bquery("ID?")  // ID TEK/7912AD,V77.1,F3.1;
	bquery("VS1?") // VS1 +500.E-03;
	bquery("HS1?") // HS1 NONE;
	bquery("HS2?") // HS2 +500.E-12;
	bquery("VU1?") // VU1 V;
	bquery("HU1?") // HU1 NONE;
	bquery("HU2?") // HU2 TODO;
	bquery("MAI?")
	bquery("GRI?")
	bquery("FOC?")
	bquery("LIMITS?")
	bquery("ERR?")

	// if true {
	// 	return
	// }

	bquery("MODE DIG")
	// cmd("TEST") // cmd or query?
	log.Printf("sleep 2...")
	time.Sleep(2 * time.Second)
	// pquery("DIG DAT;READ PTR,VER") // NOTE we can't currently handle multiple responses, so do not issue two reads in one command
	ptr := pquery("DIG DAT;READ PTR")
	ver := pquery("READ VER")
	// log.Printf("\nptr=%x\nver=%x", ptr, ver)
	points := tek.PtrVerToATC(ptr, ver)
	log.Printf("points: %s", points)

	bquery("DUMP RAW")
	if true {
		return
	}

	// // cmd("GRI 40") //?
	// cmd("MAI 448")
	// cmd("FOC 32")
	// cmd("GRAT ON")

	// cmd("OPC ON") //TODO check SRQ for done status
	// cmd("DIG GRAT")
	// gpib.Flush()
	log.Printf("sleep 20...") // FIXME query status instead of sleeping?
	time.Sleep(20 * time.Second)
	cmd("ATC")
	bquery("READ ATC")
	cmd("DIG DEF,2")   // digitize defects 2x (and average?)
	pquery("READ PTR") // pointers array
	pquery("READ VER") // vertical array
	bquery("READ DEF") // defects

	bquery("DIG DAT;READ PTR,VER")
	bquery("READ ATC") // ATC=average-to-center
	// bquery("READ DEF")
	// bquery("ID?")
	bquery("DUMP RAW")
	if false {
		pquery("XXX TODO")
	}
	// LOAD <block> - load defects data

}
