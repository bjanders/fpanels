package main

import (
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
                radioPanel.DisplayInt(fpanels.ACTIVE_1, i)
                multiPanel.DisplayInt(fpanels.ROW_1, i)
        }
	switchPanel.SetGear(fpanels.RED_N | fpanels.RED_L | fpanels.GREEN_R)
	radioPanel.DisplayOff()
	time.Sleep(500 * time.Millisecond)
	switchPanel.SetGearOff(fpanels.RED_N)
	time.Sleep(500 * time.Millisecond)
	switchPanel.SetGearOn(fpanels.YELLOW_N)
	time.Sleep(500 * time.Millisecond)
	switchPanel.SetGearOff(fpanels.GEAR_N)
        multiSwitches := multiPanel.WatchSwitches()
        radioSwitches := radioPanel.WatchSwitches()
        switchSwitches := switchPanel.WatchSwitches()
	var switchState fpanels.SwitchState
	var panelName string
	for {
		select {
		case switchState = <-multiSwitches: panelName = "multi"
		case switchState = <-radioSwitches: panelName = "radio"
		case switchState = <-switchSwitches: panelName = "switch"
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

