package connutil

import (
	"flag"
	"log"
	"time"

	"github.com/gotmc/prologix"
	"github.com/gotmc/prologix/lib/find"
	"github.com/soypat/cereal"
)

type Conn struct {
	SerialPort string
	GpibPAD    int
	GpibSAD    int
	Delay      time.Duration
	Diag       bool
	// MaxTime    time.Duration

	tty     string
	finderr error
}

// AddFlags is to be called before [flag.Parse].
func (c *Conn) AddFlags() {
	c.tty, c.finderr = find.Find(find.SerialFilter("A603UX94")) //find.ArduinoFilter
	if c.finderr != nil {
		log.Printf("locating serial port failed, guessing ttyACM0: %s", c.finderr)
		c.tty = "ttyACM0"
	}

	// Get Virtual COM Port (VCP) serial port for Prologix.
	flag.StringVar(
		&c.SerialPort,
		"port",
		"/dev/"+c.tty,
		"Serial port for Prologix VCP GPIB controller",
	)
	if c.GpibPAD == 0 {
		c.GpibPAD = 4
	}
	if c.GpibSAD == 0 {
		c.GpibSAD = 101
	}
	if c.Delay == 0 {
		c.Delay = 100 * time.Millisecond
	}

	flag.IntVar(&c.GpibPAD, "pad", c.GpibPAD, "GPIB primary address for the device")
	flag.IntVar(&c.GpibSAD, "sad", c.GpibSAD, "GPIB secondary address for the device")
	flag.DurationVar(&c.Delay, "delay", c.Delay, "delay between writes")
	flag.BoolVar(&c.Diag, "diag", c.Diag, "xdiag and exit")
}

// Setup is to be called after variables are initialized, i.e. after both [(Conn).DefineFlags] and [flag.Parse] are called.
func (c *Conn) Setup(opts []prologix.ControllerOption) (gpib *prologix.Controller, cleanup func(), err error) {
	nocleanup := func() {}

	if c.finderr != nil && c.SerialPort == "/dev/ttyACM0" {
		// only print this if the port isn't overridden via flag
		// FIXME not following the logic of ttyacm0 not being overridden...?
		log.Printf("locating serial port failed, guessing: %s", c.finderr)
	}

	log.SetFlags(log.Lmicroseconds)

	log.Printf("Serial port = %s", c.SerialPort)

	cimpl := cereal.Tarm{}
	port, err := cimpl.OpenPort(c.SerialPort, cereal.Mode{
		BaudRate:    115200,
		ReadTimeout: time.Second * 30,
	})
	if err != nil {
		return nil, nocleanup, err
	}

	if c.Delay > 0 {
		opts = append(opts, prologix.WithWriteDelay(c.Delay))
	}
	if c.GpibSAD != 0xff {
		opts = append(opts, prologix.WithSecondaryAddress(c.GpibSAD))
	}

	gpib, err = prologix.NewController(port, c.GpibPAD, false, opts...)
	if err != nil {
		port.Close()
		return nil, nocleanup, err
	}

	cleanup = func() {
		// Return local control to the front panel.
		err = gpib.FrontPanel(true)
		if err != nil {
			log.Fatalf("error setting local control for front panel: %s", err)
		}

		// Discard any unread data on the serial port and then close.
		// TODO unsupported with cereal
		if fl, ok := port.(interface{ Flush() error }); ok {
			if err := fl.Flush(); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Printf("cannot flush %#v", port)
		}
		// err = port.Flush()
		// if err != nil {
		// 	log.Printf("error flushing serial port: %s", err)
		// }

		err = port.Close()
		if err != nil {
			log.Printf("error closing serial port: %s", err)
		}
	}
	if c.Diag {
		log.Printf("diag starting...")
		gpib.CommandController("xdiag 1 255")
		time.Sleep(time.Millisecond)
		gpib.CommandController("xdiag 0 255")
		time.Sleep(time.Millisecond * 100)
		gpib.CommandController("xdiag 0 0")
		gpib.CommandController("xdiag 1 0")
		return
	}

	return gpib, cleanup, nil
}
