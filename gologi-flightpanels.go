package main

import (
//	"fmt"
        "log"
//	"time"
	"math"
	"errors"

        "github.com/google/gousb"
//        "github.com/google/gousb/usbid"
)

type Switch uint
const (
	COM1_1 Switch = iota
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

type Display int
const (
	ACTIVE_1 Display = iota
	STANDBY_1
	ACTIVE_2
	STANDBY_2
)


type RadioPanel struct {
	ctx *gousb.Context
	device *gousb.Device
	intf *gousb.Interface
	intfDone func()
	inEndpoint  *gousb.InEndpoint
	displayState [20]byte
	autoUpdate bool
}

func NewRadioPanel() (*RadioPanel, error) {
	var err error
	panel := RadioPanel{}
	for i := 0; i < 20; i++ {
		panel.displayState[i] = 0x0f
	}

	panel.ctx = gousb.NewContext()
	panel.device, err = panel.ctx.OpenDeviceWithVIDPID(0x06a3, 0x0d05)
	if err != nil {
		panel.Close()
		return nil, err
	}
	panel.device.SetAutoDetach(true)

	// FIX: implement automatic periodic update
	panel.autoUpdate = false

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

	panel.UpdateDisplay()
	return &panel, nil
}

func (self *RadioPanel) Close() {
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

func (self *RadioPanel) DisplayInt(display Display, n int) error {
	return self.DisplayFloat(display, float32(n), 0)
}

func (self *RadioPanel) DisplayFloat(display Display, n float32, decimals int) error {
	neg := false

	if decimals < 0 || decimals > 5 {
		return errors.New("decimals out of range")
	}
	// Get an integer number that contains all digits
	// we want to display
	tempN := int(n * float32(math.Pow10(decimals)))
	if tempN < 0 {
		tempN  = -tempN
		neg = true
	}
	if display < 0 || display > 3 {
		return errors.New("display number out of range")
	}
	if tempN < -9999 ||  tempN > 99999 {
		return errors.New("value to be displayed out of range")
	}

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
			v = (tempN /pow) % 10
			if decimals != 0 && digit == decimals {
				v |= 0xd0
			}
		}
		i := int(display) * 5 + 4 -  digit
		self.displayState[i] = byte(v)
	}
	if self.autoUpdate {
		self.UpdateDisplay()
	}
	return nil
}

func (self *RadioPanel) UpdateDisplay() {
	n, err := self.device.Control(0x21, 0x09, 0x03, 0x00, self.displayState[0:20])
	log.Printf("Wrote %d bytes", n)
	if err != nil {
		log.Printf("%v", err)
	}
}
type SwitchState struct {
	Switch Switch
	Value uint
}

func (self *RadioPanel) WatchSwitches() chan SwitchState {
	c := make(chan SwitchState)
	go readSwitches(self.inEndpoint, c)
	return c
}

func readSwitches(ep *gousb.InEndpoint, c chan SwitchState) {
	var data [3]byte
	var state uint64
	var newState uint64

	stream, err := ep.NewStream(3, 1)
	if err != nil {
		log.Fatalf("Could not create read stream: %v", err)
	}
	defer stream.Close()

	for {
		_, err := stream.Read(data[:])
		if err != nil {
			log.Fatalf("Read error: %v", err)
		}
		newState = uint64(data[0]) | uint64(data[1]) << 8 |  uint64(data[2]) << 16
		changed := state  ^ newState
		state = newState
		for i := COM1_1; i <= ENC2_CCW_2; i++ {
			if (changed >> i) & 1 == 1 {
				val := uint(state >> i & 1)
				if val == 0 && i != ACT_1 && i != ACT_2 {
					continue
				}
				c <- SwitchState{i, val}
			}
		}
	}
}

func main() {
	radioPanel, err := NewRadioPanel()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer radioPanel.Close()
	c := radioPanel.WatchSwitches()
	for {
		switchState := <-c
		log.Printf("%d: %d", switchState.Switch, switchState.Value)
		radioPanel.DisplayInt(ACTIVE_1, int(switchState.Switch))
		radioPanel.DisplayInt(STANDBY_1, int(switchState.Value))
		radioPanel.UpdateDisplay()
	}
//	time.Sleep(10 * time.Millisecond)
//	radioPanel.DisplayFloat(0, 0.1, 2)
//	radioPanel.UpdateDisplay()
//	for {
//		t := time.Now()
//		radioPanel.DisplayInteger(0, t.Hour())
//		radioPanel.DisplayInteger(1, t.Minute())
//		radioPanel.DisplayInteger(3, t.Second())
//		radioPanel.UpdateDisplay()
//		time.Sleep(1 * time.Second)
//	}
}
