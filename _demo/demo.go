package main

import (
	"time"

	"github.com/google/chrometracing"
)

const (
	tidMain = iota
	tidNetwork
)

func networkRequest() {
	defer chrometracing.Event("networkRequest", tidNetwork).Done()
	time.Sleep(100 * time.Millisecond)
}

func main() {
	//defer chrometracing.Flush()
	defer chrometracing.Event("main", tidMain).Done()
	networkRequest()
	time.Sleep(500 * time.Millisecond)
}
