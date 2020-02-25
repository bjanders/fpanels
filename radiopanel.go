package fpanels

import (
	"fmt"
	"time"

	"github.com/google/gousb"
)

// Radio panel switches
const (
	COM1_1 SwitchID = iota
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
	blank = 0x0f
	dot   = 0xd0
	dash  = 0xef
)

// Radio panel displays
const (
	ACTIVE_1 DisplayID = iota
	STANDBY_1
	ACTIVE_2
	STANDBY_2
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
	displayState [20]byte
}

// NewRadioPanel creats a new instance of the radio panel
func NewRadioPanel() (*RadioPanel, error) {
	var err error
	panel := RadioPanel{}
	panel.id = RADIO
	for i := range panel.displayState {
		panel.displayState[i] = blank
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
	panel.connected = true
	return &panel, nil
}

// Close disconnects the connection to the radio panel and releases all
// related resources
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

// ID returns RADIO
func (panel *RadioPanel) ID() PanelID {
	return panel.id
}

func (panel *RadioPanel) setSwitches(s PanelSwitches) {
	panel.switches = s
}

// IsSwitchSet returns true if the switch is with ID id is set
func (panel *RadioPanel) IsSwitchSet(id SwitchID) bool {
	return panel.switches.IsSet(id)
}

// DisplayString displays the string s on the given display. The string is limited
// to the numbers 0-9, a dot '.', dash/minus '-' and space, with a length of max
// five characters.  If any other character is
// used then the underlying previous character is left intact. This allows you to
// update different areas of the dislay in separate calls. For example:
//   panel.DisplayString(ACTIVE_1, "12   ")
//   panel.DisplayString(ACTIVE_1, "** 34")
//   panel.DisplayString(ACTIVE_1, "** 56")
// will display the the following sequence on the upper left display:
//   12
//   12 34
//   12 56
func (panel *RadioPanel) DisplayString(display DisplayID, s string) {
	if display < ACTIVE_1 || display > STANDBY_2 {
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
	panel.displayMutex.Unlock()

}
func (panel *RadioPanel) refreshDisplay() {
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

// WatchSwitches creates a channel for receiving SwitchState events
// whenever the state of a switch changes
func (panel *RadioPanel) WatchSwitches() chan SwitchState {
	c := make(chan SwitchState)
	go readSwitches(panel, panel.inEndpoint, c)
	return c
}

func (panel *RadioPanel) noZeroSwitch(s SwitchID) bool {
	if s == ACT_1 || s == ACT_2 {
		return false
	}
	return true
}
