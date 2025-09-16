package tek

// TODO move to package tek

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"slices"
	"strings"
)

// Unpack unpacks tektronics gpib "pack" data
//
// 3 bytes: hdr, count hi, count low
// data bytes: hi,low
// checksum,semicolon
func Unpack(pack []byte) ([]uint16, error) {
	var ints []uint16
	if len(pack) < 5 {
		return nil, io.EOF
	}
	if pack[0] != '%' {
		return nil, fmt.Errorf("invalid header: want %% got %q", pack[0])
	}
	count := int(pack[1])*256 + int(pack[2])
	if len(pack) != count+4 {
		// return nil, fmt.Errorf
		log.Printf("invalid length: expect %d, got %d", count+4, len(pack))
	}
	end := pack[len(pack)-1]
	if end != ';' {
		return nil, fmt.Errorf("invalid trailer: expect ; got %q", end)
	}
	// csum := pack[len(pack)-2]
	// TODO how to verify checksum byte?
	dataEnd := len(pack) - 2
	if err := checksum(pack[1:dataEnd], pack[dataEnd]); err != nil {
		// log.Print(err)
		return nil, err
	}
	data := pack[3:dataEnd]

	for len(data) > 1 {
		i := int(data[0])*256 + int(data[1])
		// // assumes max is always 512, maybe not?
		// if i > 512 {
		// 	log.Printf("value overflows at %d: %d %d", count-len(data)-1, data[0], data[1])
		// }
		ints = append(ints, uint16(i))
		data = data[2:]
	}
	if len(data) != 0 {
		log.Printf("byte remains: %x", data)
	}
	return ints, nil
}

func checksum(data []byte, expect byte) error {
	// 8-bit, 2's complement number that is modulo-256 sum of preceding bytes
	var s = int(expect)
	for _, c := range data {
		s += int(c)
	}
	if s&0xff != 0 {
		return fmt.Errorf("bad checksum: got 0x%2x, expect 0", s&0xff)
	}
	return nil
}

type Point struct{ X, Y float32 }

func (p Point) String() string { return fmt.Sprintf("{%03.1f,%03.1f}", p.X, p.Y) }

type Points []Point

func (ps Points) String() string {
	s := make([]string, 0, len(ps))
	for _, p := range ps {
		s = append(s, p.String())
	}
	return strings.Join(s, ",")
}

func PtrVerToATC(ptr, ver []uint16) Points {
	var coords []Point
	discards := 0
	for i := range ptr {
		if ptr[i] == math.MaxUint16 {
			// -1, i.e. no values - so discard
			// does this mean intensity was too low?
			discards++
			continue
		}
		// WRONG data points ptr[i]-1 through ptr[i+1]-1
		// first := ptr[i] - 1
		// var last uint16
		// if i == len(ptr)-1 {
		// 	last = uint16(len(ver) - 1)
		// } else {
		// 	last = ptr[i+1] - 1
		// }

		// data points ptr[i-1]+1 through ptr[i]
		var first uint16
		if i > 0 {
			first = ptr[i-1] + 1
		} //else 0
		last := ptr[i]

		// if first == last {
		// 	// or should we interpret this as the top and bottom edge are the same??
		// 	log.Printf("ptr: skip 0-len range %d %d", i, first)
		// 	continue
		// }
		var points []uint16
		for _, p := range ver[first : last+1] {
			if p <= 512 {
				points = append(points, p)
			}
			// TODO instead of ignoring, draw defects a different color?
		}
		if len(points) == 0 {
			continue
		}
		if len(points) != 2 {
			log.Printf("warn: %d edges at %d", len(points), i)
		}
		//FIXME how to handle multiple points in column?
		// just do max()-min()?
		a := slices.Max(points)
		b := slices.Min(points)
		c := float32(a+b) / 2.0
		// if i > 15 {
		fmt.Fprintf(os.Stderr, "% 3d: f=%3d l=%3d pts=%v a=%3d b=%3d c=%04.1f {X=%d, Y=%04.1f}\n", i, first, last, points, a, b, c, i+1, c)
		// }
		coords = append(coords, Point{X: float32(i + 1), Y: c})
	}
	if discards > 0 {
		log.Printf("ptr: discarded %d empty entries", discards)
	}
	return coords
}

// TODO add split func and test on either session's "READ PTR,VER"
// also rewrite unpack to operate on stream?
