package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bjanders/fpanels"
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
	radioSwitches := radioPanel.SwitchCh()
	for i := -500; i < 500; i++ {
		time.Sleep(1000 * time.Microsecond)
		radioPanel.DisplayInt(fpanels.Display1Active, i)
		s := fmt.Sprintf("%d.0", i)
		radioPanel.DisplayString(fpanels.Display2Active, s)
		radioPanel.DisplayFloat(fpanels.Display1Standby, float64(i), 2)
		radioPanel.DisplayInt(fpanels.Display2Standby, i*1000)
	}
	radioPanel.DisplayString(fpanels.Display2Active, "")
	sleep()
	radioPanel.DisplayString(fpanels.Display2Active, "1234567890")
	sleep()
	radioPanel.DisplayString(fpanels.Display2Active, ".")
	sleep()
	radioPanel.DisplayString(fpanels.Display2Active, "...")
	sleep()
	radioPanel.DisplayString(fpanels.Display2Active, ".42")
	sleep()
	radioPanel.DisplayString(fpanels.Display2Active, "88.")
	sleep()
	radioPanel.DisplayString(fpanels.Display2Active, "-----")
	sleep()
	radioPanel.DisplayString(fpanels.Display2Active, "##0##")
	sleep()
	s := "-----"
	for i := 0; i < 5; i++ {
		s = s[:5-i]
		radioPanel.DisplayString(fpanels.Display2Active, s)
		time.Sleep(100 * time.Millisecond)
	}
	radioPanel.DisplayString(fpanels.Display2Active, " . . . . .")
	var switchState fpanels.SwitchState
	for {
		switchState = <-radioSwitches
		var state int
		if switchState.On {
			state = 1
		}

		log.Printf("%d: %d", switchState.Switch, state)
		radioPanel.DisplayInt(fpanels.Display1Active, int(switchState.Switch))
		radioPanel.DisplayInt(fpanels.Display1Standby, state)

	}
}
