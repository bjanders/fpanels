package flightpanels

import (
	"time"
	"github.com/google/gousb"
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
	GREEN_N byte = 1<<iota
	GREEN_L
	GREEN_R
	RED_N
	RED_L
	RED_R
	YELLOW_N = GREEN_N | RED_N
	YELLOW_L = GREEN_L | RED_L
	YELLOW_R = GREEN_R | RED_R
	GREEN_ALL = GREEN_N | GREEN_L | GREEN_R
	RED_ALL = RED_N | RED_L | RED_R
	YELLOW_ALL = YELLOW_N | YELLOW_L | YELLOW_R
	GEAR_N = YELLOW_N
	GEAR_L = YELLOW_L
	GEAR_R = YELLOW_R
	GEAR_ALL = YELLOW_ALL
)


type SwitchPanel struct {
	Panel
	displayState  [1]byte
}

func NewSwitchPanel() (*SwitchPanel, error) {
	var err error
	panel := SwitchPanel{}
	panel.displayState[0] = 0
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

func (self *SwitchPanel) Close() {
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

func (self *SwitchPanel) SetGear(s byte) {
	self.displayMutex.Lock()
	self.displayState[0] = s
	self.displayDirty = true
	self.displayMutex.Unlock()
}

func (self *SwitchPanel) SetGearOn(s byte) {
	self.displayMutex.Lock()
	self.displayState[0] = self.displayState[0] | s
	self.displayDirty = true
	self.displayMutex.Unlock()
}

func (self *SwitchPanel) SetGearOff(s byte) {
	self.displayMutex.Lock()
	self.displayState[0] = self.displayState[0] & ^s
	self.displayDirty = true
	self.displayMutex.Unlock()
}

func (self *SwitchPanel) refreshDisplay() {
	for {
		// refresh rate 20 Hz
		time.Sleep(50 * time.Millisecond)
		self.displayMutex.Lock()
		if self.displayDirty {
			self.device.Control(0x21, 0x09, 0x03, 0x00, self.displayState[:])
			self.displayDirty = false
		}
		self.displayMutex.Unlock()
	}
}



func (self *SwitchPanel) WatchSwitches() chan SwitchState {
	c := make(chan SwitchState)
	go readSwitches(self, self.inEndpoint, c)
	return c
}

func (self *SwitchPanel) noZeroSwitch(s Switch) bool {
	if s >= ENG_OFF && s <= ENG_START {
		return true
	}
	if s == GEAR_UP || s == GEAR_DOWN {
		return true
	}
	return false
}
