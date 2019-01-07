package fpanels

import (
	"fmt"
	"github.com/google/gousb"
	"time"
)

// Buttons
const (
	ALT SwitchId = iota
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
	BTN_VS
	BTN_APR
	BTN_REV
	AUTO_THROTTLE
	FLAPS_UP
	FLAPS_DOWN
	TRIM_DOWN
	TRIM_UP
)

// LED lights
const (
	LED_AP byte = 1 << iota
	LED_HDG
	LED_NAV
	LED_IAS
	LED_ALT
	LED_VS
	LED_APR
	LED_REV
)

const multi_dash = 0xde

const (
	ROW_1 DisplayId = iota
	ROW_2
)

type MultiPanel struct {
	Panel
	displayState [11]byte
}

func NewMultiPanel() (*MultiPanel, error) {
	var err error
	panel := MultiPanel{}
	panel.id = MULTI
	for i := range panel.displayState {
		if i == 10 {
			// 11th byte is the LEDs
			panel.displayState[i] = 0x00
		} else {
			panel.displayState[i] = blank
		}
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
	panel.Connected = true
	return &panel, nil
}

func (panel *MultiPanel) Close() {
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

func (panel *MultiPanel) Id() PanelId {
	return panel.id
}

func (panel *MultiPanel) setSwitches(s PanelSwitches) {
	panel.Switches = s
}

func (panel *MultiPanel) IsSwitchSet(id SwitchId) bool {
	return panel.Switches.IsSet(id)
}

func (panel *MultiPanel) LEDs(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[10] = leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

func (panel *MultiPanel) LEDsOn(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[10] = panel.displayState[10] | leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

func (panel *MultiPanel) LEDsOff(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[10] = panel.displayState[10] & ^leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

func (panel *MultiPanel) LEDsOnOff(leds byte, val float64) {
	if val > 0 {
		panel.LEDsOn(leds)
	} else {
		panel.LEDsOff(leds)
	}
}

func (panel *MultiPanel) DisplayString(display DisplayId, s string) {
	if display != ROW_1 && display != ROW_2 {
		return
	}

	var d [5]byte
	displayStart := int(display) * 5
	disp := panel.displayState[displayStart : displayStart+5]
	dIdx := 0
	for _, c := range s {
		if dIdx > 4 {
			break
		}
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			d[dIdx] = byte(c - '0')
		case ' ':
			d[dIdx] = blank
		case '-':
			d[dIdx] = multi_dash
		default:
			// leave current char as is
			d[dIdx] = disp[dIdx]
		}
		dIdx++
	}

	panel.displayMutex.Lock()
	defer panel.displayMutex.Unlock()
	panel.displayDirty = true
	dIdx--
	// align right and fill with blanks
	for i := 4; i >= 0; i-- {
		if dIdx < 0 {
			disp[i] = blank
		} else {
			disp[i] = d[dIdx]
		}
		dIdx--
	}
}

func (panel *MultiPanel) DisplayInt(display DisplayId, n int) {
	s := fmt.Sprintf("%d", n)
	panel.DisplayString(display, s)
}

func (panel *MultiPanel) refreshDisplay() {
	for {
		// refresh rate 20 Hz
		time.Sleep(50 * time.Millisecond)
		panel.displayMutex.Lock()
		if panel.displayDirty {
			panel.device.Control(0x21, 0x09, 0x03, 0x00, panel.displayState[:])
			panel.displayDirty = false
		}
		panel.displayMutex.Unlock()
	}
}

func (panel *MultiPanel) WatchSwitches() chan SwitchState {
	c := make(chan SwitchState)
	go readSwitches(panel, panel.inEndpoint, c)
	return c
}

func (panel *MultiPanel) noZeroSwitch(s SwitchId) bool {
	if s >= ALT && s <= ENC_CCW {
		return true
	}
	if s == TRIM_DOWN || s == TRIM_UP {
		return true
	}
	return false
}
