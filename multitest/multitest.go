package main

import (
	"log"
	"time"

	"github.com/bjanders/fpanels"
)

func main() {
	multiPanel, err := fpanels.NewMultiPanel()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer multiPanel.Close()

	for i := 0; i < 100; i++ {
		multiPanel.DisplayInt(fpanels.Row1, i)
		time.Sleep(100 * time.Millisecond)
	}
	for i := byte(0x01); i <= 0x80; i = i << 1 {
		multiPanel.LEDsOn(i)
		time.Sleep(500 * time.Millisecond)
	}
}
