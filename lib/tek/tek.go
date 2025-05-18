package tek

// TODO move to package tek

import (
	"fmt"
	"io"
	"log"
	"os"
	"slices"
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
	// TODO
	// 8-bit, 2's complement number that is modulo-256 sum of preceding bytes
	var s = int(expect)
	for _, c := range data {
		s += int(c)
	}
	if s&0xff != 0 {
		return fmt.Errorf("bad checksum %x", s&0xff)
	}
	return nil
}

type xy struct{ X, Y float32 }

func PtrVerToATC(ptr, ver []int) []xy {
	var coords []xy
	for i := range ptr {
		// data points ptr[i]-1 through ptr[i+1]-1
		first := ptr[i] - 1
		var last int
		if i == len(ptr)-1 {
			last = len(ver) - 1
		} else {
			last = ptr[i+1] - 1
		}
		var points []int
		for _, p := range ver[first:last] {
			// defects will be negative, right?
			if p >= 0 {
				points = append(points, p)
			}
		}
		//FIXME how to handle multiple points in column?
		// just do max()-min()?
		a := slices.Max(points)
		b := slices.Min(points)
		c := float32(a+b) / 2.0
		if i > 15 {
			fmt.Fprintf(os.Stderr, "%d: f=%d l=%d pts=%v a=%d b=%d c=%f X=%d Y=%f\n", i, first, last, points, a, b, c, i+1, c)
		}
		coords = append(coords, xy{X: float32(i + 1), Y: c})
	}
	return coords
}

// TODO add split func and test on either session's "READ PTR,VER"
// also rewrite unpack to operate on stream?
