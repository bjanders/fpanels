package main

import (
	"fmt"
	"github.com/bjanders/fpanels"
	"log"
	"time"
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
		radioPanel.DisplayInt(fpanels.ACTIVE_1, i)
		s := fmt.Sprintf("%d", i)
		radioPanel.DisplayString(fpanels.ACTIVE_2, s)
		multiPanel.DisplayInt(fpanels.ROW_1, i)
	}
	switchPanel.LEDs(fpanels.N_RED | fpanels.L_RED | fpanels.R_GREEN)
	radioPanel.DisplayOff()
	time.Sleep(500 * time.Millisecond)
	switchPanel.LEDsOff(fpanels.N_RED)
	time.Sleep(500 * time.Millisecond)
	switchPanel.LEDsOn(fpanels.N_YELLOW)
	time.Sleep(500 * time.Millisecond)
	switchPanel.LEDsOff(fpanels.N_ALL)
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
		log.Printf("%s: %d: %d", panelName, switchState.Switch, switchState.Value)
		radioPanel.DisplayInt(fpanels.ACTIVE_1, int(switchState.Switch))
		radioPanel.DisplayInt(fpanels.STANDBY_1, int(switchState.Value))

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
