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

type Sigrok struct{ CaptureConfig }

type CaptureConfig struct {
	Rate, Samples uint64
	Channels      Channels
	AddTimestamp  bool
}

func (sr *Sigrok) AddFlags() {
	// flag.StringVar(&sigrokFlag, "sigrok", "0,0", "rate,samples for sigrok")
	flag.Uint64Var(&sr.Rate, "la.rate", 0, "sigrok: sample rate")
	flag.Uint64Var(&sr.Samples, "la.samp", 0, "sigrok: number of samples")
	flag.BoolVar(&sr.AddTimestamp, "la.ts", true, "add timestamp to filename to avoid overwrites")
}
func (sr *Sigrok) CheckFlags() (start bool) {
	n := 0
	if sr.Rate > 0 {
		n++
	}
	if sr.Samples > 0 {
		n++
	}
	if n == 1 {
		log.Fatalf("only one of -sigsamp, -sigrate provided; both required")
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
	return true
}

// maps from channel name to signal name
type Channels []Channel //map[string]string

type Channel [2]string // channel name, signal name

func (c Channels) SRArgs() []string {
	if c == nil {
		return nil
	}
	cs := make([]string, 0, len(c))
	for _, ch := range c {
		if len(ch[1]) == 0 {
			cs = append(cs, ch[0])
		} else {
			// NOTE renamed channels must begin with D - see
			// https://github.com/pico-coder/sigrok-pico/issues/41
			// cs = append(cs, fmt.Sprintf("%s=%s%s", ch[0], ch[0], ch[1]))
			cs = append(cs, fmt.Sprintf("%s=D%s", ch[0], ch[1]))
		}
	}
	return []string{
		"--channels",
		strings.Join(cs, ","),
	}
}

// FIXME channels must stay in order for srpico
// default channels (and names)
var DefaultSRChannels = Channels{
	{"D2", "11_ATN"},
	{"D3", "10_SRQ"},
	{"D4", "09_IFC"},
	{"D5", "08_NDAC"},
	{"D6", "07_NRFD"},
	{"D7", "06_DAV"},
	{"D8", "05_EOI"},
	{"D9", "04_DIO4"},
	{"D10", "16_DIO8"},
	{"D11", "17_REN"},
	{"D12", ""},
	{"D13", ""},
	{"D14", ""},
	{"D15", ""},
	{"D16", "03_DIO3"},
	{"D17", "02_DIO2"},
	{"D18", "01_DIO1"},
	{"D19", "13_DIO5"},
	{"D20", "14_DIO6"},
	{"D21", "15_DIO7"},
}

func (sr *Sigrok) StartCapture() {
	if sr.Channels == nil {
		sr.Channels = DefaultSRChannels
	}
	pico, err := find.Find(find.PiPicoFilter)
	if err != nil {
		log.Printf("sigrok: %s", err)
	}
	log.Printf("sigrok port %q", pico)
	if len(pico) == 0 {
		log.Fatal("no pico??")
	}
	cli := exec.Command("sigrok-cli")
	dev := fmt.Sprintf("raspberrypi-pico:conn=/dev/%s:serialcomm=115200/flow=0", pico)
	cli.Args = append(cli.Args,
		"-l", "2",
		"-d", dev,
		"--config", fmt.Sprintf("samplerate=%d", sr.Rate),
		"--samples", strconv.FormatUint(sr.Samples, 10),
		// D11-15 unused, but it fails if they are removed
		// "--channels", "D2,D3,D4,D5,D6,D7,D8,D9,D10,D11,D12,D13,D14,D15,D16,D17,D18,D19,D20,D21",
	)
	cli.Args = append(cli.Args, sr.Channels.SRArgs()...)
	fname := "sigrok_out"
	if sr.AddTimestamp {
		fname += "_" + time.Now().Format("02Jan_15_04_05.000")
	}
	outBinary := false
	if outBinary {
		fname += ".bin"
		cli.Args = append(cli.Args,
			"-o", fname,
			"--output-format", "binary",
		)
	} else {
		fname += ".sr"
		cli.Args = append(cli.Args,
			"-o", fname,
		)
	}
	cli.Stderr, cli.Stdout, cli.Stdin = os.Stderr, os.Stdout, os.Stdin
	log.Printf("running\n%v", cli.Args)
	if err := cli.Run(); err != nil {
		log.Printf("sigrok error: %s", err)
	}

	fi, err := os.Stat(fname)
	if err != nil {
		log.Printf("stat'ing sigrok output: %s", err)
	} else {
		log.Printf("sigrok output file %s size %d", fname, fi.Size())
	}
}
