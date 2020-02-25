// Package fpanels provides an interface to Logitech/Saitek flight panels.
//
// Use the New*Panel() functions to create an instance of the specific panel
// type. When you are done, call the panel's Close() function.
package fpanels

import (
	"errors"
	"strings"
	"sync"

	"github.com/google/gousb"
)

// USB vendor and product IDs
const (
	USB_VENDOR_PANEL   = 0x06a3
	USB_PRODUCT_RADIO  = 0x0d05
	USB_PRODUCT_MULTI  = 0x0d06
	USB_PRODUCT_SWITCH = 0x0d67
)

// SwitchId identifies a switch on a panel
type SwitchId uint

// PanelId identifies the panel type
type PanelId int

// PanelIds
const (
	RADIO PanelId = iota
	MULTI
	SWITCH
)

// Panel is the base struct for all panels
type panel struct {
	ctx          *gousb.Context
	device       *gousb.Device
	intf         *gousb.Interface
	inEndpoint   *gousb.InEndpoint
	displayMutex sync.Mutex
	id           PanelId
	switches     PanelSwitches
	displayDirty bool
	intfDone     func()
	connected    bool
}

// SwitchState contains the state of a switch on a panel
type SwitchState struct {
	Panel  PanelId
	Switch SwitchId
	On     bool
}

// PanelSwitches is the state of all switches on a panel, one bit per switch
type PanelSwitches uint32

// DisplayId identifies a display on a panel
type DisplayId uint

// SwitchingPanel provides an interface to panels with switches
type SwitchingPanel interface {
	setSwitches(s PanelSwitches)
	noZeroSwitch(i SwitchId) bool
	Id() PanelId
	IsSwitchSet(i SwitchId) bool
}

// StringDisplayer provides an interface to panels that can display strings
type StringDisplayer interface {
	DisplayString(display DisplayId, s string)
}

// LEDDisplayer priovides an interface to panels that has LEDs
type LEDDisplayer interface {
	LEDs(leds byte)
	LEDsOn(leds byte)
	LEDsOff(leds byte)
	LEDsOnOff(leds byte, val float64)
}

// PanelIdMap maps a panel Id string to a PanelId
var PanelIdMap = map[string]PanelId{
	"RADIO":  RADIO,
	"MULTI":  MULTI,
	"SWITCH": SWITCH,
}

// PanelIdString maps a panel string to a PanelId. The string s is case insensitive.
func PanelIdString(s string) (PanelId, error) {
	s = strings.ToUpper(s)
	p, ok := PanelIdMap[s]
	if !ok {
		return 0, errors.New("Unknown panel type")
	}
	return p, nil
}

// SwitchIdMap maps a switch ID string to a SwitchId
var SwitchIdMap = map[string]SwitchId{
	// radio
	"COM1_1":     COM1_1,
	"COM2_1":     COM2_1,
	"NAV1_1":     NAV1_1,
	"NAV2_1":     NAV2_1,
	"ADF_1":      ADF_1,
	"DME_1":      DME_1,
	"XPDR_1":     XPDR_1,
	"COM1_2":     COM1_2,
	"COM2_2":     COM2_2,
	"NAV1_2":     NAV1_2,
	"NAV2_2":     NAV2_2,
	"ADF_2":      ADF_2,
	"DME_2":      DME_2,
	"XPDR_2":     XPDR_2,
	"ACT_1":      ACT_1,
	"ACT_2":      ACT_2,
	"ENC1_CW_1":  ENC1_CW_1,
	"ENC1_CCW_1": ENC1_CCW_1,
	"ENC2_CW_1":  ENC2_CW_1,
	"ENC2_CCW_1": ENC2_CCW_1,
	"ENC1_CW_2":  ENC1_CW_2,
	"ENC1_CCW_2": ENC1_CCW_2,
	"ENC2_CW_2":  ENC2_CW_2,
	"ENC2_CCW_2": ENC2_CCW_2,
	// multi
	"ALT":           ALT,
	"VS":            VS,
	"IAS":           IAS,
	"HDG":           HDG,
	"CRS":           CRS,
	"ENC_CW":        ENC_CW,
	"ENC_CCW":       ENC_CCW,
	"BTN_AP":        BTN_AP,
	"BTN_HDG":       BTN_HDG,
	"BTN_NAV":       BTN_NAV,
	"BTN_IAS":       BTN_IAS,
	"BTN_ALT":       BTN_ALT,
	"BTN_VS":        BTN_VS,
	"BTN_APR":       BTN_APR,
	"BTN_REV":       BTN_REV,
	"AUTO_THROTTLE": AUTO_THROTTLE,
	"FLAPS_UP":      FLAPS_UP,
	"FLAPS_DOWN":    FLAPS_DOWN,
	"TRIM_DOWN":     TRIM_DOWN,
	"TRIM_UP":       TRIM_UP,
	// switch
	"BAT":        BAT,
	"ALTERNATOR": ALTERNATOR,
	"AVIONICS":   AVIONICS,
	"FUEL":       FUEL,
	"DEICE":      DEICE,
	"PITOT":      PITOT,
	"COWL":       COWL,
	"PANEL":      PANEL,
	"BEACON":     BEACON,
	"NAV":        NAV,
	"STROBE":     STROBE,
	"TAXI":       TAXI,
	"LANDING":    LANDING,
	"ENG_OFF":    ENG_OFF,
	"ALT_R":      ALT_R,
	"ALT_L":      ALT_L,
	"ALT_BOTH":   ALT_BOTH,
	"ENG_START":  ENG_START,
	"GEAR_UP":    GEAR_UP,
	"GEAR_DOWN":  GEAR_DOWN,
}

// LEDMap maps a LED Id string to the corresponding LED bits
var LEDMap = map[string]byte{
	// switch
	"N_GREEN":  N_GREEN,
	"L_GREEN":  L_GREEN,
	"R_GREEN":  R_GREEN,
	"N_RED":    N_RED,
	"L_RED":    L_RED,
	"R_RED":    R_RED,
	"N_YELLOW": N_GREEN | N_RED,
	"L_YELLOW": L_GREEN | L_RED,
	"R_YELLOW": R_GREEN | R_RED,
	// multi
	"LED_AP":  LED_AP,
	"LED_HDG": LED_HDG,
	"LED_NAV": LED_NAV,
	"LED_IAS": LED_IAS,
	"LED_ALT": LED_ALT,
	"LED_VS":  LED_VS,
	"LED_APR": LED_APR,
	"LED_REV": LED_REV,
}

// DisplayMap maps the display names to a DisplayId
var DisplayMap = map[string]DisplayId{
	// radio
	"ACTIVE_1":  ACTIVE_1,
	"STANDBY_1": STANDBY_1,
	"ACTIVE_2":  ACTIVE_2,
	"STANDBY_2": STANDBY_2,
	// multi
	"ROW_1": ROW_1,
	"ROW_2": ROW_2,
}

// SwitchIdString maps a Switch ID string to a SwitchId. The ID string s
// is case insensitive.
func SwitchIdString(s string) (SwitchId, error) {
	s = strings.ToUpper(s)
	p, ok := SwitchIdMap[s]
	if !ok {
		return 0, errors.New("Unknown switch")
	}
	return p, nil
}

// LEDString maps a LED name to the corresponding LED bits. The string s
// is case insensitive.
func LEDString(s string) (byte, error) {
	s = strings.ToUpper(s)
	l, ok := LEDMap[s]
	if !ok {
		return 0, errors.New("Unknown LED")
	}
	return l, nil
}

// DisplayIdString maps a Display name to the DisplayId. The string s
// is case insesitive.
func DisplayIdString(s string) (DisplayId, error) {
	s = strings.ToUpper(s)
	d, ok := DisplayMap[s]
	if !ok {
		return 0, errors.New("Unknown display")
	}
	return d, nil
}

// IsSet returns true if the switch id is set.
func (switches PanelSwitches) IsSet(id SwitchId) bool {
	return uint32(switches)&1<<uint32(id) != 0
}

// SwitchState returns the statee of the switch with ID id, 0 or 1
func (switches PanelSwitches) SwitchState(id SwitchId) uint {
	return uint((uint32(switches) >> uint32(id)) & 1)
}

func readSwitches(panel SwitchingPanel, inEndpoint *gousb.InEndpoint, c chan SwitchState) error {
	var data [3]byte
	var state uint32
	var newState uint32

	stream, err := inEndpoint.NewStream(3, 1)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		_, err := stream.Read(data[:])
		if err != nil {
			return err
		}
		newState = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16
		changed := state ^ newState
		state = newState
		panel.setSwitches(PanelSwitches(state))
		for i := SwitchId(0); i < 24; i++ {
			if (changed>>i)&1 == 1 {
				val := uint(state >> i & 1)
				//if val == 0 && panel.noZeroSwitch(i) {
				//	continue
				//}
				c <- SwitchState{panel.Id(), i, val == 1}
			}
		}
	}
}
