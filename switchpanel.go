package fpanels

import (
	"github.com/google/gousb"
	"time"
)

const (
	BAT Switch = iota
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
	GREEN_N byte = 1 << iota
	GREEN_L
	GREEN_R
	RED_N
	RED_L
	RED_R
	YELLOW_N   = GREEN_N | RED_N
	YELLOW_L   = GREEN_L | RED_L
	YELLOW_R   = GREEN_R | RED_R
	GREEN_ALL  = GREEN_N | GREEN_L | GREEN_R
	RED_ALL    = RED_N | RED_L | RED_R
	YELLOW_ALL = YELLOW_N | YELLOW_L | YELLOW_R
	GEAR_N     = YELLOW_N
	GEAR_L     = YELLOW_L
	GEAR_R     = YELLOW_R
	GEAR_ALL   = YELLOW_ALL
)

type SwitchPanel struct {
	Panel
	displayState [1]byte
}

func NewSwitchPanel() (*SwitchPanel, error) {
	var err error
	panel := SwitchPanel{}
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

func (panel *SwitchPanel) SetGear(s byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = s
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

func (panel *SwitchPanel) SetGearOn(s byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = panel.displayState[0] | s
	panel.displayDirty = true
	panel.displayMutex.Unlock()
}

func (panel *SwitchPanel) SetGearOff(s byte) {
	panel.displayMutex.Lock()
	panel.displayState[0] = panel.displayState[0] & ^s
	panel.displayDirty = true
	panel.displayMutex.Unlock()
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

func (panel *SwitchPanel) noZeroSwitch(s Switch) bool {
	if s >= ENG_OFF && s <= ENG_START {
		return true
	}
	if s == GEAR_UP || s == GEAR_DOWN {
		return true
	}
	return false
}
