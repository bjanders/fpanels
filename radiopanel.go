package fpanels

import (
	"fmt"
	"sync"

	"github.com/google/gousb"
)

// Radio panel switches
const (
	Rot1COM1 SwitchID = iota
	Rot1COM2
	Rot1NAV1
	Rot1NAV2
	Rot1ADF
	Rot1DME
	Rot1XPDR
	Rot2Com1
	Rot2Com2
	Rot2NAV1
	Rot2NAV2
	Rot2ADF
	Rot2DME
	Rot2XPDR
	SwAct1
	SwAct2
	Enc1CW1
	Enc1CCW1
	Enc2CW1
	Enc2CCW1
	Enc1CW2
	Enc1CCW2
	Enc2CW2
	Enc2CCW2
)

const (
	blank = 0x0f
	dot   = 0xd0
	dash  = 0xef
)

// Radio panel displays
const (
	Display1Active DisplayID = iota
	Display1Standby
	Display2Active
	Display2Standby
)

// RadioPanel represents a Saitek/Logitech radio panel. The panel has:
//
// - Two seven position function switches
//
// - Two dual rotary encoders
//
// - Two momentary push buttons
//
// - Four five number segment displays
type RadioPanel struct {
	panel
}

// NewRadioPanel creats a new instance of the radio panel
func NewRadioPanel() (*RadioPanel, error) {
	var err error
	panel := RadioPanel{}
	panel.id = Radio
	panel.displayState = make([]byte, 22)
	for i := range panel.displayState {
		panel.displayState[i] = blank
	}
	panel.displayDirty = true
	panel.displayCond = sync.NewCond(&panel.displayMutex)
	panel.ctx = gousb.NewContext()
	panel.device, err = panel.ctx.OpenDeviceWithVIDPID(USBVendorPanel, USBProductRadio)
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

	panel.switchCh = make(chan SwitchState)
	go panel.readSwitches()

	// FIX: Add WaitGroup
	go panel.refreshDisplay()
	panel.connected = true
	return &panel, nil
}

// DisplayString displays the string s on the given display. The string is limited
// to the numbers 0-9, a dot '.', dash/minus '-' and space, with a length of max
// five characters.  If any other character is
// used then the underlying previous character is left intact. This allows you to
// update different areas of the dislay in separate calls. For example:
//   panel.DisplayString(Display1Active, "12   ")
//   panel.DisplayString(Display1Active, "** 34")
//   panel.DisplayString(Display1Active, "** 56")
// will display the the following sequence on the upper left display:
//   12
//   12 34
//   12 56
func (panel *RadioPanel) DisplayString(display DisplayID, s string) {
	if display < Display1Active || display > Display2Standby {
		return
	}
	var d [5]byte
	displayStart := int(display) * 5
	disp := panel.displayState[displayStart : displayStart+5]
	dIdx := 0
	if len(s) > 0 && s[0] == '.' {
		s = " " + s
	}
	for _, c := range s {
		if dIdx > 4 && c != '.' {
			break
		}
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			d[dIdx] = byte(c - '0')
			dIdx++
		case ' ':
			d[dIdx] = blank
			dIdx++
		case '.':
			d[dIdx-1] |= dot
		case '-':
			d[dIdx] = dash
			dIdx++
		default:
			// leave current char as is
			d[dIdx] = disp[dIdx]
			dIdx++
		}
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

// DisplayInt displays the integer n on the given display
func (panel *RadioPanel) DisplayInt(display DisplayID, n int) {
	panel.DisplayString(display, fmt.Sprintf("%d", n))
}

// DisplayFloat displays the floating point number n with the given number of
// decimals on the given display
func (panel *RadioPanel) DisplayFloat(display DisplayID, n float64, decimals int) {
	panel.DisplayString(display, fmt.Sprintf("%.*f", decimals, n))
}

// DisplayOff turns the display off
func (panel *RadioPanel) DisplayOff() {
	panel.displayMutex.Lock()
	for i := 0; i < len(panel.displayState); i++ {
		panel.displayState[i] = 0xff
	}
	panel.displayDirty = true
	panel.displayCond.Signal()
	panel.displayMutex.Unlock()

}
func (panel *RadioPanel) refreshDisplay() {
	for {
		panel.displayMutex.Lock()
		for !panel.displayDirty {
			panel.displayCond.Wait()
		}
		// 0x09 is REQUEST_SET_CONFIGURATION
		// 0x0300 is:
		// 	 0x03 HID_REPORT_TYPE_FEATURE
		//   0x00 Report ID 0
		panel.device.Control(gousb.ControlOut|gousb.ControlClass|gousb.ControlInterface, 0x09,
			0x0300, 0x00, panel.displayState[:])
		// FIX: Check if Control() returns an error and return it somehow or exit
		panel.displayDirty = false
		panel.displayMutex.Unlock()
	}
}

func (panel *RadioPanel) noZeroSwitch(s SwitchID) bool {
	if s == SwAct1 || s == SwAct2 {
		return false
	}
	return true
}
