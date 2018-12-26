package fpanels

import (
	"errors"
	"github.com/google/gousb"
	"math"
	"time"
)

const (
	ALT Switch = iota
	VS
	IAS
	HDG
	CRS
	ENC_CW
	ENC_CCW
	BTN_AP
	BTN_HDG
	BTN_NAV
	BTN_IAS
	BTN_ALT
	BTN_APR
	BTN_REV
	_
	AUTO_THROTTLE
	FLAPS_UP
	FLAPS_DOWN
	TRIM_DOWN
	TRIM_UP
)

const (
	ROW_1 Display = iota
	ROW_2
)

type MultiPanel struct {
	Panel
	displayState [11]byte
}

func NewMultiPanel() (*MultiPanel, error) {
	var err error
	panel := MultiPanel{}
	for i := 0; i < len(panel.displayState); i++ {
		panel.displayState[i] = 0x0f
	}
	panel.displayDirty = true
	panel.ctx = gousb.NewContext()
	panel.device, err = panel.ctx.OpenDeviceWithVIDPID(USB_VENDOR_PANEL, USB_PRODUCT_MULTI)
	if panel.device == nil || err != nil {
		panel.Close()
		return nil, err
	}
	panel.device.SetAutoDetach(true)

	// Initialize switches
	panel.intf, panel.intfDone, err = panel.device.DefaultInterface()
	if err != nil {
		panel.Close()
		return nil, err
	}

	panel.inEndpoint, err = panel.intf.InEndpoint(1)
	if err != nil {
		panel.Close()
		return nil, err
	}
	// FIX: Add WaitGroup
	go panel.refreshDisplay()
	return &panel, nil
}

func (self *MultiPanel) Close() {
	// FIX: Stop threads
	if self.intfDone != nil {
		self.intfDone()
	}
	if self.device != nil {
		self.device.Close()
	}
	if self.ctx != nil {
		self.ctx.Close()
	}
}

// FIX: Add DisplayString() function

func (self *MultiPanel) DisplayInt(display Display, n int) error {
	return self.DisplayFloat(display, float32(n), 0)
}

func (self *MultiPanel) DisplayFloat(display Display, n float32, decimals int) error {
	neg := false

	if decimals < 0 || decimals > 5 {
		return errors.New("decimals out of range")
	}
	// Get an integer number that contains all digits
	// we want to display
	tempN := int(n * float32(math.Pow10(decimals)))
	if tempN < 0 {
		tempN = -tempN
		neg = true
	}
	if display < 0 || display > 1 {
		return errors.New("display number out of range")
	}
	if tempN < -9999 || tempN > 99999 {
		return errors.New("value to be displayed out of range")
	}
	self.displayMutex.Lock()
	defer self.displayMutex.Unlock()
	self.displayDirty = true
	for digit := 0; digit < 5; digit++ {
		var v int
		// Get the number we want to display in the 10s
		pow := int(math.Pow10(digit))
		// FIX: Show leading zero
		if pow > tempN {
			if neg {
				v = 0xde
				neg = false
			} else {
				v = 0xff
			}
		} else {
			v = (tempN / pow) % 10
			if decimals != 0 && digit == decimals {
				v |= 0xd0
			}
		}
		i := int(display)*5 + 4 - digit
		self.displayState[i] = byte(v)
	}
	return nil
}

func (self *MultiPanel) refreshDisplay() {
	for {
		// refresh rate 20 Hz
		time.Sleep(50 * time.Millisecond)
		self.displayMutex.Lock()
		if self.displayDirty {
			self.device.Control(0x21, 0x09, 0x03, 0x00, self.displayState[0:11])
			self.displayDirty = false
		}
		self.displayMutex.Unlock()
	}
}

func (self *MultiPanel) WatchSwitches() chan SwitchState {
	c := make(chan SwitchState)
	go readSwitches(self, self.inEndpoint, c)
	return c
}

func (self *MultiPanel) noZeroSwitch(s Switch) bool {
	if s >= ALT && s <= ENC_CCW {
		return true
	}
	if s == TRIM_DOWN || s == TRIM_UP {
		return true
	}
	return false
}
