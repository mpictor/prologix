package find

import "testing"

func Test_Find(t *testing.T) {
	// Find(PiPicoFilter)
	t.Error("unimpl")
}

func Test_usbTtys(t *testing.T) {
	ttys, err := AllUsbTtys()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range ttys {
		if PiPicoFilter(&tt) {
			t.Errorf("pico match: %#v", tt)
		}
		if ArduinoFilter(&tt) {
			t.Errorf("arduino match: %#v", tt)
		}
	}
	t.Error(ttys)
}
