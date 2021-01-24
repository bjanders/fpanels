package fpanels

import (
	"time"

	"github.com/google/gousb"
)

// Switch panel switches
const (
	BAT SwitchID = iota
	ALTERNATOR
	AVIONICS
	FUEL
	DEICE
	PITOT
	COWL
	PANEL
	BEACON
	NAV
	STROBE
	TAXI
	LANDING
	ENG_OFF
	ALT_R
	ALT_L
	ALT_BOTH
	ENG_START
	GEAR_UP
	GEAR_DOWN
)

// Switch panel landing gear lights
const (
	N_GREEN byte = 1 << iota
	L_GREEN
	R_GREEN
	N_RED
	L_RED
	R_RED
	N_YELLOW   = N_GREEN | N_RED
	L_YELLOW   = L_GREEN | L_RED
	R_YELLOW   = R_GREEN | R_RED
	ALL_GREEN  = N_GREEN | L_GREEN | R_GREEN
	ALL_RED    = N_RED | L_RED | R_RED
	ALL_YELLOW = N_YELLOW | L_YELLOW | R_YELLOW
	N_ALL      = N_YELLOW
	L_ALL      = L_YELLOW
	R_ALL      = R_YELLOW
	ALL        = ALL_YELLOW
)

// SwitchPanel represents a Saitek/Logitech switch panel. The panel has:
//
// - A five position switch
//
// - Thirteen two position switches
//
// - A two position landing gear lever
//
// - Three red/green landing gear indicator LEDs
type SwitchPanel struct {
	panel
	displayState [1]byte
}

// NewSwitchPanel create a new instance of the Logitech/Saitek switch panel
func NewSwitchPanel() (*SwitchPanel, error) {
	var err error
	panel := SwitchPanel{}
	panel.id = SWITCH
	panel.displayState[0] = 0
	panel.displayDirty = true
	panel.ctx = gousb.NewContext()
	panel.device, err = panel.ctx.OpenDeviceWithVIDPID(USB_VENDOR_PANEL, USB_PRODUCT_SWITCH)
	if panel.device == nil || err != nil {
		panel.Close()
		return nil, err
	}
	panel.device.SetAutoDetach(true)

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

// Close disconnects the connection to the switch panel and releases all
// related resources
func (panel *SwitchPanel) Close() {
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

// ID returns SWITCH
func (panel *SwitchPanel) ID() PanelID {
	return panel.id
}

func (panel *SwitchPanel) setSwitches(s PanelSwitches) {
	panel.switches = s
}

// IsSwitchSet returns true if the switch is set
func (panel *SwitchPanel) IsSwitchSet(id SwitchID) bool {
	return panel.switches.IsSet(id)
}

// LEDs turns on/off the LEDs given by leds. See the LED_* constants.
// For example calling
//   panel.LEDs(L_GREEN | R_GREEN)
// will turn on the left and right green landing gear LEDs and turn off
// all other LEDs
func (panel *SwitchPanel) LEDs(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

// LEDsOn turns on the ELDS given by leds and leaves the other LED states
// intact. See the switch panel LED constants. Multiple LEDs can be ORed
// together. For example:
//  panel.LEDsOn(L_GREEN | R_GREEN)
func (panel *SwitchPanel) LEDsOn(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = panel.displayState[0] | leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

// LEDsOff turn off the LEDs given by leds and leaves all other LED states
// instact. See the switch panel LED constants. Multiple LEDs can be ORed
// together. For example:
//   panel.LEDsOff(L_GREEN | R_GREEN)
func (panel *SwitchPanel) LEDsOff(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = panel.displayState[0] & ^leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

// LEDsOnOff turns on or off the LEDs given by leds. If val is 0 then
// the LEDs will be turned off, else they will be turned on. All other
// LEDs are left intact. See the switch panel LED constants.
// Multiple LEDs can be ORed togethe, for example:
//   panel.LEDsOnOff(L_GREEN | R_GREEN, 1)
func (panel *SwitchPanel) LEDsOnOff(leds byte, val float64) {
	if val > 0 {
		panel.LEDsOn(leds)
	} else {
		panel.LEDsOff(leds)
	}
}

func (panel *SwitchPanel) refreshDisplay() {
	for {
		// refresh rate 20 Hz
		time.Sleep(50 * time.Millisecond)
		panel.displayMutex.Lock()
		if panel.displayDirty {
			// 0x09 is REQUEST_SET_CONFIGURATION
			panel.device.Control(gousb.ControlOut|gousb.ControlClass|gousb.ControlInterface, 0x09,
				0x03, 0x00, panel.displayState[:])
			// FIX: Check if Control() returns an error and return it somehow or exit
			panel.displayDirty = false
		}
		panel.displayMutex.Unlock()
	}
}

// WatchSwitches creates a channel for reeiving SwitchState events
// whenever the state of a swtich changes.
func (panel *SwitchPanel) WatchSwitches() chan SwitchState {
	c := make(chan SwitchState)
	go readSwitches(panel, panel.inEndpoint, c)
	return c
}

func (panel *SwitchPanel) noZeroSwitch(s SwitchID) bool {
	if s >= ENG_OFF && s <= ENG_START {
		return true
	}
	if s == GEAR_UP || s == GEAR_DOWN {
		return true
	}
	return false
}
