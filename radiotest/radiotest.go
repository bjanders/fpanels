package main

import (
	"fmt"
	"github.com/bjanders/fpanels"
	"log"
	"time"
)

func sleep() {
	time.Sleep(500 * time.Millisecond)
}

func main() {
	radioPanel, err := fpanels.NewRadioPanel()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer radioPanel.Close()
	radioSwitches := radioPanel.WatchSwitches()
	for i := -500; i < 500; i++ {
		time.Sleep(1000 * time.Microsecond)
		radioPanel.DisplayInt(fpanels.ACTIVE_1, i)
		s := fmt.Sprintf("%d.0", i)
		radioPanel.DisplayString(fpanels.ACTIVE_2, s)
		radioPanel.DisplayFloat(fpanels.STANDBY_1, float64(i), 2)
		radioPanel.DisplayInt(fpanels.STANDBY_2, i * 1000)
	}
	radioPanel.DisplayString(fpanels.ACTIVE_2, "")
	sleep()
	radioPanel.DisplayString(fpanels.ACTIVE_2, "1234567890")
	sleep()
	radioPanel.DisplayString(fpanels.ACTIVE_2, ".")
	sleep()
	radioPanel.DisplayString(fpanels.ACTIVE_2, "...")
	sleep()
	radioPanel.DisplayString(fpanels.ACTIVE_2, ".42")
	sleep()
	radioPanel.DisplayString(fpanels.ACTIVE_2, "88.")
	sleep()
	radioPanel.DisplayString(fpanels.ACTIVE_2, "-----")
	sleep()
	radioPanel.DisplayString(fpanels.ACTIVE_2, "##0##")
	sleep()
	s := "-----"
	for i := 0; i < 5; i++ {
		s = s[:5-i]
		radioPanel.DisplayString(fpanels.ACTIVE_2, s)
		time.Sleep(100 * time.Millisecond)
	}
	radioPanel.DisplayString(fpanels.ACTIVE_2, " . . . . .")
	var switchState fpanels.SwitchState
	for {
		switchState = <-radioSwitches
		log.Printf("%d: %d",  switchState.Switch, switchState.Value)
		radioPanel.DisplayInt(fpanels.ACTIVE_1, int(switchState.Switch))
		radioPanel.DisplayInt(fpanels.STANDBY_1, int(switchState.Value))

	}
}
