package fpanels

import (
	"errors"
	"github.com/google/gousb"
	"math"
	"time"
)

const (
	COM1_1 SwitchId = iota
	COM2_1
	NAV1_1
	NAV2_1
	ADF_1
	DME_1
	XPDR_1
	COM1_2
	COM2_2
	NAV1_2
	NAV2_2
	ADF_2
	DME_2
	XPDR_2
	ACT_1
	ACT_2
	ENC1_CW_1
	ENC1_CCW_1
	ENC2_CW_1
	ENC2_CCW_1
	ENC1_CW_2
	ENC1_CCW_2
	ENC2_CW_2
	ENC2_CCW_2
)

const (
	ACTIVE_1 DisplayId = iota
	STANDBY_1
	ACTIVE_2
	STANDBY_2
)

type RadioPanel struct {
	Panel
	displayState [20]byte
}

func NewRadioPanel() (*RadioPanel, error) {
	var err error
	panel := RadioPanel{}
	for i := range panel.displayState {
		panel.displayState[i] = 0x0f
	}
	panel.displayDirty = true
	panel.ctx = gousb.NewContext()
	panel.device, err = panel.ctx.OpenDeviceWithVIDPID(USB_VENDOR_PANEL, USB_PRODUCT_RADIO)
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

func (panel *RadioPanel) Close() {
	// FIX: Stop threads
	if panel.intfDone != nil {
		panel.intfDone()
	}
	if panel.device != nil {
		panel.device.Close()
	}
	if panel.ctx != nil {
		panel.ctx.Close()
	}
}

// FIX: Add DisplayString() function

func (panel *RadioPanel) DisplayInt(display DisplayId, n int) error {
	return panel.DisplayFloat(display, float32(n), 0)
}

func (panel *RadioPanel) DisplayFloat(display DisplayId, n float32, decimals int) error {
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
	if display < 0 || display > 3 {
		return errors.New("display number out of range")
	}
	if tempN < -9999 || tempN > 99999 {
		return errors.New("value to be displayed out of range")
	}
	panel.displayMutex.Lock()
	defer panel.displayMutex.Unlock()
	panel.displayDirty = true
	for digit := 0; digit < 5; digit++ {
		var v int
		// Get the number we want to display in the 10s
		pow := int(math.Pow10(digit))
		// FIX: Show leading zero
		if pow > tempN {
			if neg {
				v = 0xef
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
		panel.displayState[i] = byte(v)
	}
	return nil
}

func (panel *RadioPanel) DisplayOff() {
	panel.displayMutex.Lock()
	for i := 0; i < len(panel.displayState); i++ {
		panel.displayState[i] = 0xff
	}
	panel.displayDirty = true
	panel.displayMutex.Unlock()

}
func (panel *RadioPanel) refreshDisplay() {
	for {
		// refresh rate 20 Hz
		time.Sleep(50 * time.Millisecond)
		panel.displayMutex.Lock()
		if panel.displayDirty {
			panel.device.Control(0x21, 0x09, 0x03, 0x00, panel.displayState[0:20])
			panel.displayDirty = false
		}
		panel.displayMutex.Unlock()
	}
}

func (panel *RadioPanel) WatchSwitches() chan SwitchState {
	c := make(chan SwitchState)
	go readSwitches(panel, panel.inEndpoint, c)
	return c
}

func (panel *RadioPanel) noZeroSwitch(s SwitchId) bool {
	if s == ACT_1 || s == ACT_2 {
		return false
	}
	return true
}
