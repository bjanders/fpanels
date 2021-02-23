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
	USBVendorPanel   = 0x06a3
	USBProductRadio  = 0x0d05
	USBProductMulti  = 0x0d06
	USBProductSwitch = 0x0d67
)

// SwitchID identifies a switch on a panel
type SwitchID uint

// PanelID identifies the panel type
type PanelID int

// PanelIDs
const (
	Radio PanelID = iota
	Multi
	Switch
)

// Panel is the base struct for all panels
type panel struct {
	ctx          *gousb.Context
	device       *gousb.Device
	intf         *gousb.Interface
	inEndpoint   *gousb.InEndpoint
	displayMutex sync.Mutex
	displayCond  *sync.Cond
	id           PanelID
	switches     PanelSwitches
	displayDirty bool
	intfDone     func()
	connected    bool
}

// SwitchState contains the state of a switch on a panel
type SwitchState struct {
	Panel  PanelID
	Switch SwitchID
	On     bool
}

// PanelSwitches is the state of all switches on a panel, one bit per switch
type PanelSwitches uint32

// DisplayID identifies a display on a panel
type DisplayID uint

// SwitchingPanel provides an interface to panels with switches
type SwitchingPanel interface {
	setSwitches(s PanelSwitches)
	noZeroSwitch(i SwitchID) bool
	ID() PanelID
	IsSwitchSet(i SwitchID) bool
}

// StringDisplayer provides an interface to panels that can display strings
type StringDisplayer interface {
	DisplayString(display DisplayID, s string)
}

// LEDDisplayer priovides an interface to panels that has LEDs
type LEDDisplayer interface {
	LEDs(leds byte)
	LEDsOn(leds byte)
	LEDsOff(leds byte)
	LEDsOnOff(leds byte, val float64)
}

// PanelIDMap maps a panel Id string to a PanelID
var PanelIDMap = map[string]PanelID{
	"RADIO":  Radio,
	"MULTI":  Multi,
	"SWITCH": Switch,
}

// PanelIDString maps a panel string to a PanelID. The string s is case insensitive.
func PanelIDString(s string) (PanelID, error) {
	s = strings.ToUpper(s)
	p, ok := PanelIDMap[s]
	if !ok {
		return 0, errors.New("Unknown panel type")
	}
	return p, nil
}

// SwitchIDMap maps a switch ID string to a SwitchID
var SwitchIDMap = map[string]SwitchID{
	// radio
	"COM1_1":     Rot1COM1,
	"COM2_1":     Rot1COM2,
	"NAV1_1":     Rot1NAV1,
	"NAV2_1":     Rot1NAV2,
	"ADF_1":      Rot1ADF,
	"DME_1":      Rot1DME,
	"XPDR_1":     Rot1XPDR,
	"COM1_2":     Rot2Com1,
	"COM2_2":     Rot2Com2,
	"NAV1_2":     Rot2NAV1,
	"NAV2_2":     Rot2NAV2,
	"ADF_2":      Rot2ADF,
	"DME_2":      Rot2DME,
	"XPDR_2":     Rot2XPDR,
	"ACT_1":      SwAct1,
	"ACT_2":      SwAct2,
	"ENC1_CW_1":  Enc1CW1,
	"ENC1_CCW_1": Enc1CCW1,
	"ENC2_CW_1":  Enc2CW1,
	"ENC2_CCW_1": Enc2CCW1,
	"ENC1_CW_2":  Enc1CW2,
	"ENC1_CCW_2": Enc1CCW2,
	"ENC2_CW_2":  Enc2CW2,
	"ENC2_CCW_2": Enc2CCW2,
	// multi
	"ALT":           RotALT,
	"VS":            RotVS,
	"IAS":           RotIAS,
	"HDG":           RotHDG,
	"CRS":           RotCRS,
	"ENC_CW":        EncCW,
	"ENC_CCW":       EncCCW,
	"BTN_AP":        BtnAP,
	"BTN_HDG":       BtnHDG,
	"BTN_NAV":       BtnNAV,
	"BTN_IAS":       BtnIAS,
	"BTN_ALT":       BtnALT,
	"BTN_VS":        BtnVS,
	"BTN_APR":       BtnAPR,
	"BTN_REV":       BtnREV,
	"AUTO_THROTTLE": AutoThrottle,
	"FLAPS_UP":      FlapsUp,
	"FLAPS_DOWN":    FlapsDown,
	"TRIM_DOWN":     TrimDown,
	"TRIM_UP":       TrimUp,
	// switch
	"BAT":        SwBat,
	"ALTERNATOR": SwAlternator,
	"AVIONICS":   SwAvionics,
	"FUEL":       SwFuel,
	"DEICE":      SwDeice,
	"PITOT":      SwPitot,
	"COWL":       SwCowl,
	"PANEL":      SwPanel,
	"BEACON":     SwBeacon,
	"NAV":        SwNav,
	"STROBE":     SwStrobe,
	"TAXI":       SwTaxi,
	"LANDING":    SwLanding,
	"ENG_OFF":    RotOff,
	"ALT_R":      RotR,
	"ALT_L":      RotL,
	"ALT_BOTH":   RotBoth,
	"ENG_START":  RotStart,
	"GEAR_UP":    GearUp,
	"GEAR_DOWN":  GearDown,
}

// LEDMap maps a LED Id string to the corresponding LED bits
var LEDMap = map[string]byte{
	// switch
	"N_GREEN":  LEDNGreen,
	"L_GREEN":  LEDLGreen,
	"R_GREEN":  LEDRGreen,
	"N_RED":    LEDNRed,
	"L_RED":    LEDLRed,
	"R_RED":    LEDRRed,
	"N_YELLOW": LEDNGreen | LEDNRed,
	"L_YELLOW": LEDLGreen | LEDLRed,
	"R_YELLOW": LEDRGreen | LEDRRed,
	// multi
	"LED_AP":  LEDAP,
	"LED_HDG": LEDHDG,
	"LED_NAV": LEDNAV,
	"LED_IAS": LEDIAS,
	"LED_ALT": LEDALT,
	"LED_VS":  LEDVS,
	"LED_APR": LEDAPR,
	"LED_REV": LEDREV,
}

// DisplayMap maps the display names to a DisplayID
var DisplayMap = map[string]DisplayID{
	// radio
	"ACTIVE_1":  Display1Active,
	"STANDBY_1": Display1Standby,
	"ACTIVE_2":  Display2Active,
	"STANDBY_2": Display2Standby,
	// multi
	"ROW_1": Row1,
	"ROW_2": Row2,
}

// SwitchIDString maps a Switch ID string to a SwitchID. The ID string s
// is case insensitive.
func SwitchIDString(s string) (SwitchID, error) {
	s = strings.ToUpper(s)
	p, ok := SwitchIDMap[s]
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

// DisplayIDString maps a Display name to the DisplayID. The string s
// is case insesitive.
func DisplayIDString(s string) (DisplayID, error) {
	s = strings.ToUpper(s)
	d, ok := DisplayMap[s]
	if !ok {
		return 0, errors.New("Unknown display")
	}
	return d, nil
}

// IsSet returns true if the switch id is set.
func (switches PanelSwitches) IsSet(id SwitchID) bool {
	return uint32(switches)&1<<uint32(id) != 0
}

// SwitchState returns the statee of the switch with ID id, 0 or 1
func (switches PanelSwitches) SwitchState(id SwitchID) uint {
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
		for i := SwitchID(0); i < 24; i++ {
			if (changed>>i)&1 == 1 {
				val := uint(state >> i & 1)
				//if val == 0 && panel.noZeroSwitch(i) {
				//	continue
				//}
				c <- SwitchState{panel.ID(), i, val == 1}
			}
		}
	}
}
