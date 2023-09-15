package main

import (
	"fmt"

	"github.com/gianz74/mailconf/internal/service"
)

func main() {

	mbsync := service.NewMbsync(nil)
	status := mbsync.Status()
	switch status {
	case service.DisabledRunning:
		fmt.Printf("enabling\n")
		mbsync.Enable()
	case service.DisabledStopped:
		fmt.Printf("enabling\n")
		mbsync.Enable()
		fmt.Printf("starting\n")
		mbsync.Start()
	case service.EnabledStopped:
		fmt.Printf("starting\n")
		mbsync.Start()
	case service.EnabledRunning:
	}
	fmt.Printf("mbsync: %+v\n", mbsync.Status())
}
