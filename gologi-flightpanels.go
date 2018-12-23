package main

import (
	"fmt"
        "log"
	"time"

        "github.com/google/gousb"
//        "github.com/google/gousb/usbid"
)

const (
	COM1_1 = 0x01
	COM2_1 = 0x02
	NAV1_1 = 0x04
	NAV2_1 = 0x08
	ADF_1 = 0x10
	DME_1 = 0x20
	XPDR_1 = 0x40
	COM1_2 = 0x80
)

type RadioPanel struct {
	Device *gousb.Device
	SwitchState [3]byte
	DisplayState [20]byte
}

//func (self *RadioPanel) SetDisplay(display int, n float32, decimals int) {
//}

func (self *RadioPanel) WriteDisplay(data []byte) {
	self.Device.Control(0x21, 0x09, 0x03, 0x00, data[0:20])
}

func NewRadioPanel(dev *gousb.Device) *RadioPanel {
	var data [20]byte
	panel := RadioPanel{}
	panel.Device = dev
	panel.WriteDisplay(data[:])
	return &panel
}

func readStream(ep *gousb.InEndpoint, c chan int) {
	var data [3]byte
	var buf []byte = data[0:3]

	stream, err := ep.NewStream(3, 1)
	if err != nil {
		log.Fatalf("Could not create read stream: %v", err)
	}
	defer stream.Close()

	for {
		_, err := stream.Read(buf)
		if err != nil {
			log.Fatalf("Read error: %v", err)
		}
		fmt.Printf("%x\n", buf)
		c <- 1
	}
}

func read(ep *gousb.InEndpoint) {
	var data [3]byte
	var buf []byte = data[0:3]
	for {
		_, err := ep.Read(buf)
		if err != nil {
			log.Fatalf("Read error: %v", err)
		}
		fmt.Printf("%x\n", buf)
	}
}

func write(dev *gousb.Device, c chan int) {
	data :=  []byte{8, 8, 1, 8, 8, 0xff, 0xff, 0xff, 0xff , 0xff , 0xff, 0xff, 0xff, 0xff, 0xff ,0xff, 0xff, 0xff, 0xff, 0xff  }
	for {
	for i := 0; i < 255; i++ {
		chandata := <-c
		data[0] = byte(i)
		fmt.Printf("%d\n", chandata)
		dev.Control(0x21, 0x09, 0x03, 0x00, data)
		time.Sleep(100* time.Millisecond)
	}
	}

}

func main() {
	ctx := gousb.NewContext()
	defer ctx.Close()

	dev, err := ctx.OpenDeviceWithVIDPID(0x06a3, 0x0d05)
	if err != nil {
		log.Fatalf("Could not open a device: %v", err)
	}
	dev.SetAutoDetach(true)
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer done()

	ep, err := intf.InEndpoint(1)
	if err != nil {
		log.Fatalf("%s.inEndpoint(): %v", intf, err)
	}

	radioPanel := NewRadioPanel(dev)
	//c1 := make(chan int)
	//go write(dev, c1)
	//go readStream(ep, c1)
	//c2 := make(chan int)
	//d := <-c2
	//fmt.Println(d)
	//time.Sleep(10 * time.Second)

}
