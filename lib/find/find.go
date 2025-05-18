package find

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FilterFn func(*Usbtty) bool

func ArduinoFilter(ut *Usbtty) bool {
	return strings.Contains(ut.Mfg, "Arduino")
}

func PiPicoFilter(ut *Usbtty) bool {
	return ut.Mfg == "Raspberry Pi" &&
		ut.Prod == "Pico"
}

func SerialFilter(s string) func(ut *Usbtty) bool {
	return func(ut *Usbtty) bool { return ut.Serial == s }
}

// Find searches for a usb serial device. If filter is not nil,
// it is used to narrow choices down. The first device for which
// it returns true (if any) is chosen.
func Find(filter FilterFn) (string, error) {
	ttys, err := AllUsbTtys()
	if err != nil {
		return "", err
	}
	if filter != nil {
		for i := range ttys {
			if filter(&ttys[i]) {
				ttys = []Usbtty{ttys[i]}
				break
			}
		}
	}

	if len(ttys) == 0 {
		return "", fmt.Errorf("no matching ttys found")
	}
	if len(ttys) == 1 {
		return ttys[0].Dev, nil
	}
	return "", fmt.Errorf("multiple ttys: %#v", ttys)
}

type Usbtty struct {
	Dev, Path string
	IDp, IDv  string
	Mfg, Prod string
	Serial    string
}

func (u Usbtty) String() string {
	return fmt.Sprintf("dev %s path %s pid/vid %s/%s mfg/prod %s/%s serial %s", u.Dev, u.Path, u.IDp, u.IDv, u.Mfg, u.Prod, u.Serial)
}

type Usbttys []Usbtty

func (uts Usbttys) String() string {
	s := make([]string, 0, len(uts))
	for _, ut := range uts {
		s = append(s, ut.String())
	}
	return strings.Join(s, "\n")
}

// find ttys on usb devices, by looking at
// /sys/class/tty and other /sys paths
//
// TODO use fs.FS for testing, though we need the equivalent of filepath.EvalSymlinks
func AllUsbTtys() (Usbttys, error) {
	var devs []Usbtty
	sct := "/sys/class/tty/"
	entries, err := os.ReadDir(sct)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.Type()&fs.ModeSymlink == 0 {
			// just in case there's anything in the dir that isn't a symlink
			continue
		}
		// we have a symlink like
		// /sys/class/tty/ttyACM0 ->
		// /sys/devices/pci0000:00/0000:00:01.3/0000:02:00.0/usb1/1-10/1-10:1.0/tty/ttyACM0
		path := filepath.Join(sct, e.Name())
		abs, err := filepath.EvalSymlinks(path)
		if err != nil {
			log.Printf("error evaluating symlink %s; skipping: %s", path, err)
			continue
		}
		if strings.Contains(abs, "usb") {
			dev, err := filepath.EvalSymlinks(filepath.Join(abs, "device"))
			if err != nil {
				log.Printf("usb but lacking device subdir?! %s %s", abs, err)
			}
			// for usb, device points up two levels. two dirs we
			// don't have to back out of here

			// /sys/devices/pci0000:00/0000:00:01.3/0000:02:00.0/usb1/1-10/1-10:1.0

			// back out one more level - not sure if all devices are arranged
			// like this but mine are...
			idP, idV, mfg, prod, serial, err := readUsbInfo(filepath.Dir(dev))
			if err != nil {
				log.Printf("%s: %s", abs, err)
			}
			devs = append(devs, Usbtty{
				Dev:    e.Name(),
				Path:   abs,
				IDp:    idP,
				IDv:    idV,
				Mfg:    mfg,
				Prod:   prod,
				Serial: serial,
			})
		}
	}
	return devs, nil
}

// realpath /sys/class/tty/ttyA*/device/
// /sys/devices/pci0000:00/0000:00:01.3/0000:02:00.0/usb1/1-7/1-7:1.0
// /sys/devices/pci0000:00/0000:00:01.3/0000:02:00.0/usb1/1-10/1-10:1.0

// reads prod and vendor ids, and mfg/product/serial strings
//
// returns last error encountered, ignoring os.ErrNotExist.
// errors do not prevent reading additional files or returning data collected.
func readUsbInfo(dev string) (idp, idv, mfg, prod, serial string, err error) {
	b, rerr := os.ReadFile(filepath.Join(dev, "idProduct"))
	if rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
		err = rerr
	}
	idp = strings.TrimSpace(string(b))
	b, rerr = os.ReadFile(filepath.Join(dev, "idVendor"))
	if rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
		err = rerr
	}
	idv = strings.TrimSpace(string(b))
	b, rerr = os.ReadFile(filepath.Join(dev, "manufacturer"))
	if rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
		err = rerr
	}
	mfg = strings.TrimSpace(string(b))
	b, rerr = os.ReadFile(filepath.Join(dev, "product"))
	if rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
		err = rerr
	}
	prod = strings.TrimSpace(string(b))
	b, rerr = os.ReadFile(filepath.Join(dev, "serial"))
	if rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
		err = rerr
	}
	serial = strings.TrimSpace(string(b))
	return idp, idv, mfg, prod, serial, err
}
