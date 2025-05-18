package analyzer

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gotmc/prologix/lib/find"
)

// type CaptureConfig struct {
// 	Rate, Samples uint64
// 	Channels      Channels
// 	AddTimestamp  bool
// }

type GusmanB struct {
	// var (
	//  sigrokFlag string
	// rate, samples uint64
	// ts            bool
	CaptureConfig
	// )
}

func (gb *GusmanB) AddFlags() {
	flag.Uint64Var(&gb.Rate, "la.rate", 0, "sample rate")
	flag.Uint64Var(&gb.Samples, "la.samp", 0, "number of samples")
	flag.BoolVar(&gb.AddTimestamp, "la.ts", true, "add timestamp to filename to avoid overwrites")
}
func (gb *GusmanB) CheckFlags() (start bool) {
	n := 0
	if gb.Rate > 0 {
		n++
	}
	if gb.Samples > 0 {
		n++
	}
	if n == 1 {
		log.Fatalf("only one of -la.samp, -la.rate provided; both required")
	}
	if n == 0 {
		return
	}
	// if sigrokFlag == "0,0" {
	// 	return
	// }
	// elems := strings.Split(sigrokFlag, ",")
	// if len(elems) != 2 {
	// 	log.Fatalf("-sigrok: need two comma-separated ints, got %s", sigrokFlag)
	// }
	// rate, err := strconv.ParseUint(elems[0], 10, 64)
	// if err != nil {
	// 	log.Fatalf("error parsing -sigrok: %s", err)
	// }
	// samples, err := strconv.ParseUint(elems[1], 10, 64)
	// if err != nil {
	// 	log.Fatalf("error parsing -sigrok: %s", err)
	// }
	// gb.Rate = rate
	// gb.Samples = samples
	// gb.AddTimestamp = ts
	return true
}

// maps from channel name to signal name
// type Channels []Channel //map[string]string

// type Channel [2]string // channel name, signal name

func (c Channels) GBArgs() string {
	if c == nil {
		return ""
	}
	cs := make([]string, 0, len(c))
	for _, ch := range c {
		if len(ch[1]) == 0 {
			cs = append(cs, ch[0])
		} else {
			cs = append(cs, fmt.Sprintf("%s:%s", ch[0], ch[1]))
		}
	}
	return strings.Join(cs, ",")
}

// default channels (and names)
var DefaultGBChannels = Channels{
	{"1", "DIO1"},
	{"2", "DIO2"},
	{"3", "DIO3"},
	{"4", "DIO4"},
	{"5", "DIO5"},
	{"6", "DIO6"},
	{"7", "DIO7"},
	{"8", "DIO8"},
	{"9", "REN"},
	{"10", "EOI"},
	{"11", "DAV"},
	{"12", "NRFD"},
	{"13", "NDAC"},
	{"14", "IFC"},
	{"15", "SRQ"},
	{"16", "ATN"},
}

func (gb *GusmanB) Capture() error {
	if gb.Channels == nil {
		gb.Channels = DefaultGBChannels
	}
	pico, err := find.Find(find.PiPicoFilter)
	if err != nil {
		// log.Printf("sigrok: %s", err)
		return err
	}
	log.Printf("gusmanb port %q", pico)
	if len(pico) == 0 {
		log.Fatal("no pico??")
	}
	cli := exec.Command("/home/mark/experiments/elec/equip/gusman_logic_probe/code/logicanalyzer/Software/LogicAnalyzer/CLCapture/bin/Release/net8.0/CLCapture")
	dev := fmt.Sprintf("/dev/%s", pico)
	cli.Args = append(cli.Args,
		"capture",
		dev,                                  // device
		strconv.FormatUint(gb.Rate, 10),      // rate
		gb.Channels.GBArgs(),                 // channels
		"2",                                  // pre-trigger samples
		strconv.FormatUint(gb.Samples, 10),   // post-trigger samples
		`TriggerType:Edge,Channel:1,Value:1`, // trigger condition
	)
	// cli.Args = append(cli.Args, gb.Channels.GBArgs()...)
	fname := "gb_out"
	if gb.AddTimestamp {
		fname += "_" + time.Now().Format("02Jan_15_04_05.000")
	}
	fname += ".lac"
	cli.Args = append(cli.Args, fname)

	cli.Stderr, cli.Stdout, cli.Stdin = os.Stderr, os.Stdout, os.Stdin
	log.Printf("running\n%v", cli.Args)
	if err := cli.Run(); err != nil {
		// log.Printf("gusmanb error: %s", err)
		return err
	}

	fi, err := os.Stat(fname)
	if err != nil {
		return err
		// log.Printf("stat'ing gusmanb output: %s", err)
	} else {
		log.Printf("gusmanb output file %s size %d", fname, fi.Size())
	}
	return nil
}
