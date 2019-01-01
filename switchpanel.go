package fpanels

import (
	"github.com/google/gousb"
	"time"
)

const (
	BAT SwitchId = iota
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

type SwitchPanel struct {
	Panel
	displayState [1]byte
}

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
	panel.Connected = true
	return &panel, nil
}

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

func (panel *SwitchPanel) Id() PanelId {
	return panel.id
}

func (panel *SwitchPanel) setSwitches(s PanelSwitches) {
	panel.Switches = s
}

func (panel *SwitchPanel) IsSwitchSet(id SwitchId) bool {
	return panel.Switches.IsSet(id)
}

func (panel *SwitchPanel) LEDs(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

func (panel *SwitchPanel) LEDsOn(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = panel.displayState[0] | leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

func (panel *SwitchPanel) LEDsOff(leds byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = panel.displayState[0] & ^leds
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

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
			panel.device.Control(0x21, 0x09, 0x03, 0x00, panel.displayState[:])
			panel.displayDirty = false
		}
		panel.displayMutex.Unlock()
	}
}

func (panel *SwitchPanel) WatchSwitches() chan SwitchState {
	c := make(chan SwitchState)
	go readSwitches(panel, panel.inEndpoint, c)
	return c
}

func (panel *SwitchPanel) noZeroSwitch(s SwitchId) bool {
	if s >= ENG_OFF && s <= ENG_START {
		return true
	}
	if s == GEAR_UP || s == GEAR_DOWN {
		return true
	}
	return false
}
