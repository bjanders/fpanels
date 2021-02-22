package fpanels

import (
	"log"
	"time"

	"github.com/google/gousb"
)

// Switch panel switches
const (
	SwBat SwitchID = iota
	SwAlternator
	SwAvionics
	SwFuel
	SwDeice
	SwPitot
	SwCowl
	SwPanel
	SwBeacon
	SwNav
	SwStrobe
	SwTaxi
	SwLanding
	RotOff
	RotR
	RotL
	RotBoth
	RotStart
	GearUp
	GearDown
)

// Switch panel landing gear lights
const (
	LEDNGreen byte = 1 << iota
	LEDLGreen
	LEDRGreen
	LEDNRed
	LEDLRed
	LEDRRed
	LEDNYellow   = LEDNGreen | LEDNRed
	LEDLYellow   = LEDLGreen | LEDLRed
	LEDRYellow   = LEDRGreen | LEDRRed
	LEDAllGreen  = LEDNGreen | LEDLGreen | LEDRGreen
	LEDAllRed    = LEDNRed | LEDLRed | LEDRRed
	LEDAllYellow = LEDNYellow | LEDLYellow | LEDRYellow
	LEDNAll      = LEDNYellow
	LEDLAll      = LEDLYellow
	LEDRAll      = LEDRYellow
	LEDAll       = LEDAllYellow
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
	panel.id = Switch
	panel.displayState[0] = 0
	panel.displayDirty = true
	panel.ctx = gousb.NewContext()
	panel.device, err = panel.ctx.OpenDeviceWithVIDPID(USBVendorPanel, USBProductSwitch)
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

// ID returns Switch
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

// LEDs turns on/off the LEDs given by leds. See the LED* constants.
// For example calling
//   panel.LEDs(LEDLGreen | LEDRGreen)
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
//  panel.LEDsOn(LEDLGreen | LEDRGreen)
func (panel *SwitchPanel) LEDsOn(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = panel.displayState[0] | leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

// LEDsOff turn off the LEDs given by leds and leaves all other LED states
// instact. See the switch panel LED constants. Multiple LEDs can be ORed
// together. For example:
//   panel.LEDsOff(LEDLGreen | LEDRGreen)
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
//   panel.LEDsOnOff(LEDLGreen | LEDRGreen, 1)
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
			_, err := panel.device.Control(gousb.ControlOut|gousb.ControlClass|gousb.ControlInterface, 0x09,
				0x0300, 0x01, panel.displayState[:])
			if err != nil {
				log.Print(err)
			}
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
	if s >= RotOff && s <= RotStart {
		return true
	}
	if s == GearUp || s == GearDown {
		return true
	}
	return false
}
