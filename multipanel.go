package fpanels

import (
	"fmt"
	"sync"

	"github.com/google/gousb"
)

// Multi panel switches and buttons
const (
	RotALT SwitchID = iota
	RotVS
	RotIAS
	RotHDG
	RotCRS
	EncCW
	EncCCW
	BtnAP
	BtnHDG
	BtnNAV
	BtnIAS
	BtnALT
	BtnVS
	BtnAPR
	BtnREV
	AutoThrottle
	FlapsUp
	FlapsDown
	TrimDown
	TrimUp
)

// Multi panel button LED lights
const (
	LEDAP byte = 1 << iota
	LEDHDG
	LEDNAV
	LEDIAS
	LEDALT
	LEDVS
	LEDAPR
	LEDREV
)

const multiDash = 0xde

// Multi panel displays
const (
	Row1 DisplayID = iota
	Row2
)

// MultiPanel represents a Saitek/Logitech multi panel. The panel has:
//
// - A five position switch
//
// - Eight push buttons with individually controlable backlight.
//
// - A rotary encoder
//
// - A two position switch
//
// - A two position momentary switch
//
// - A pitch trim rotary encoder
//
// - A two row segment display with five numbers on each row. Use DisplayString or DisplayInt to display
// text on the panels. The displays are identified by the Row1 and Row2 constants.
type MultiPanel struct {
	panel
}

// NewMultiPanel creates a new instances of the Logitech/Saitek multipanel
func NewMultiPanel() (*MultiPanel, error) {
	var err error
	panel := MultiPanel{}
	panel.id = Multi
	panel.displayState = make([]byte, 12)
	for i := range panel.displayState {
		panel.displayState[i] = blank
	}
	panel.displayState[10] = 0x00
	panel.displayState[11] = 0xff
	panel.displayDirty = true
	panel.displayCond = sync.NewCond(&panel.displayMutex)
	panel.ctx = gousb.NewContext()
	panel.device, err = panel.ctx.OpenDeviceWithVIDPID(USBVendorPanel, USBProductMulti)
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
	panel.switchCh = make(chan SwitchState)
	go panel.readSwitches()
	go panel.refreshDisplay()
	panel.connected = true
	return &panel, nil
}

// LEDs turns on/off the LEDs given by leds. See the LED* constants.
// For example calling
//   panel.LEDs(LEDAP | LEDVS)
// will turn on the AP and VS LEDs and turn off all other LEDs.
func (panel *MultiPanel) LEDs(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[10] = leds
	panel.displayDirty = true
	panel.displayCond.Signal()
	panel.displayMutex.Unlock()
}

// LEDsOn turns on the LEDs given by leds and leaves all other LED states
// intact. See the LED* constants. Multiple LEDs can be ORed together,
// for example
//   panel.LEDsOn(LEDAP | LEDVS)
func (panel *MultiPanel) LEDsOn(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[10] = panel.displayState[10] | leds
	panel.displayDirty = true
	panel.displayCond.Signal()
	panel.displayMutex.Unlock()
}

// LEDsOff turns off the LEDs given by leds and leaves all other LED states
// intact. See the LED* constants. Multiple LEDs can be ORed together.
// For example
//   panel.LEDsOff(LEDAP | LEDVS)
func (panel *MultiPanel) LEDsOff(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[10] = panel.displayState[10] & ^leds
	panel.displayDirty = true
	panel.displayCond.Signal()
	panel.displayMutex.Unlock()
}

// LEDsOnOff turns on or off the LEDs given by leds. If val is 0 then
// the LEDs will be turned offm else they will be turned on. All
// other LEDs are left intact. See the LED* constants.
// Multiple LEDs can be ORed together, for example
//   panel.LEDsOnOff(LEDAP | LEDVS, 1)
func (panel *MultiPanel) LEDsOnOff(leds byte, val float64) {
	if val > 0 {
		panel.LEDsOn(leds)
	} else {
		panel.LEDsOff(leds)
	}
}

// DisplayString displays the string given by s on the display given by
// display. The string is limited to the numbers 0-9 and spaces. Row2 can
// additionally show a dash/minus '-'. If any other char is used the
// underlying previous character is left intact. This allows you to update
// different areas of the dislay in sepeate calls. For example:
//   panel.DisplayString(Row1, "12   ")
//   panel.DisplayString(Row1, "** 34")
//   panel.DisplayString(Row1, "** 56")
// will display the the following sequence on the upper display:
//   12
//   12 34
//   12 56
func (panel *MultiPanel) DisplayString(display DisplayID, s string) {
	if display != Row1 && display != Row2 {
		return
	}

	var d [5]byte
	charmap := map[rune]byte{
		'0': 0,
		'1': 1,
		'2': 2,
		'3': 3,
		'4': 4,
		'5': 5,
		'6': 6,
		'7': 7,
		'8': 8,
		'9': 9,
		' ': blank,
		'-': multiDash,
	}
	displayStart := int(display) * 5
	disp := panel.displayState[displayStart : displayStart+5]
	dIdx := 0
	for _, c := range s {
		if dIdx > 4 {
			break
		}
		char, ok := charmap[c]
		if ok {
			d[dIdx] = char
		} else {
			// leave current char as is
			d[dIdx] = disp[dIdx]
		}
		dIdx++
	}

	panel.displayMutex.Lock()
	defer panel.displayMutex.Unlock()
	panel.displayDirty = true
	panel.displayCond.Signal()
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

// DisplayInt will display the integer n on the given display
func (panel *MultiPanel) DisplayInt(display DisplayID, n int) {
	s := fmt.Sprintf("%d", n)
	panel.DisplayString(display, s)
}

func (panel *MultiPanel) noZeroSwitch(s SwitchID) bool {
	if s >= RotALT && s <= EncCCW {
		return true
	}
	if s == TrimDown || s == TrimUp {
		return true
	}
	return false
}
