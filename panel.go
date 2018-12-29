package fpanels

import (
	"github.com/google/gousb"
	"log"
	"sync"
)

type SwitchId uint

const (
	USB_VENDOR_PANEL   = 0x06a3
	USB_PRODUCT_RADIO  = 0x0d05
	USB_PRODUCT_MULTI  = 0x0d06
	USB_PRODUCT_SWITCH = 0x0d67
)

type PanelId int

const (
	RADIO PanelId = iota
	MULTI
	SWITCH
)

type Panel struct {
	ctx          *gousb.Context
	device       *gousb.Device
	intf         *gousb.Interface
	inEndpoint   *gousb.InEndpoint
	displayMutex sync.Mutex
	Switches     PanelSwitches
	displayDirty bool
	intfDone     func()
}

type SwitchState struct {
	Switch SwitchId
	Value  uint
}

type PanelSwitches uint32

type DisplayId uint

type SwitchingPanel interface {
	setSwitches(s PanelSwitches)
	noZeroSwitch(i SwitchId) bool
}

func (switches PanelSwitches) IsSet(id SwitchId) bool {
	return uint32(switches)&1<<uint32(id) != 0
}

func readSwitches(panel SwitchingPanel, inEndpoint *gousb.InEndpoint, c chan SwitchState) {
	var data [3]byte
	var state uint32
	var newState uint32

	stream, err := inEndpoint.NewStream(3, 1)
	if err != nil {
		log.Fatalf("Could not create read stream: %v", err)
	}
	defer stream.Close()

	for {
		_, err := stream.Read(data[:])
		if err != nil {
			log.Fatalf("Read error: %v", err)
		}
		newState = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16
		changed := state ^ newState
		state = newState
		panel.setSwitches(PanelSwitches(state))
		for i := SwitchId(0); i < 24; i++ {
			if (changed>>i)&1 == 1 {
				val := uint(state >> i & 1)
				if val == 0 && panel.noZeroSwitch(i) {
					continue
				}
				c <- SwitchState{i, val}
			}
		}
	}
}
