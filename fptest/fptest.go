package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bjanders/fpanels"
)

func main() {
	radioPanel, err := fpanels.NewRadioPanel()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer radioPanel.Close()
	multiPanel, err := fpanels.NewMultiPanel()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer multiPanel.Close()
	switchPanel, err := fpanels.NewSwitchPanel()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer switchPanel.Close()
	for i := -1000; i < 1000; i++ {
		time.Sleep(1000 * time.Microsecond)
		radioPanel.DisplayInt(fpanels.Display1Active, i)
		s := fmt.Sprintf("%d", i)
		radioPanel.DisplayString(fpanels.Display2Active, s)
		multiPanel.DisplayInt(fpanels.Row1, i)
		multiPanel.DisplayInt(fpanels.Row2, i)

	}
	for i := byte(1); i != 0; i = i << 1 {
		multiPanel.LEDsOn(i)
		time.Sleep(100 * time.Millisecond)
	}
	switchPanel.LEDs(fpanels.LEDNRed | fpanels.LEDLRed | fpanels.LEDRGreen)
	radioPanel.DisplayOff()
	time.Sleep(500 * time.Millisecond)
	switchPanel.LEDsOff(fpanels.LEDNRed)
	time.Sleep(500 * time.Millisecond)
	switchPanel.LEDsOn(fpanels.LEDNYellow)
	time.Sleep(500 * time.Millisecond)
	switchPanel.LEDsOff(fpanels.LEDNAll)
	multiSwitches := multiPanel.WatchSwitches()
	radioSwitches := radioPanel.WatchSwitches()
	switchSwitches := switchPanel.WatchSwitches()
	var switchState fpanels.SwitchState
	var panelName string
	for {
		select {
		case switchState = <-multiSwitches:
			panelName = "multi"
		case switchState = <-radioSwitches:
			panelName = "radio"
		case switchState = <-switchSwitches:
			panelName = "switch"
		}
		var state int
		if switchState.On {
			state = 1
		}
		log.Printf("%s: %d: %d", panelName, switchState.Switch, state)
		radioPanel.DisplayInt(fpanels.Display1Active, int(switchState.Switch))
		radioPanel.DisplayInt(fpanels.Display1Standby, state)

	}
	//      time.Sleep(10 * time.Millisecond)
	//      radioPanel.DisplayFloat(0, 0.1, 2)
	//      radioPanel.UpdateDisplay()
	//      for {
	//              t := time.Now()
	//              radioPanel.DisplayInteger(0, t.Hour())
	//              radioPanel.DisplayInteger(1, t.Minute())
	//              radioPanel.DisplayInteger(3, t.Second())
	//              radioPanel.UpdateDisplay()
	//              time.Sleep(1 * time.Second)
	//      }
}
