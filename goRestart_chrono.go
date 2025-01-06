package main

import (
	"fmt"
	"time"
)

func timeTrigger() {
	var secs int
	var old uint64
	_, _, secs = time.Now().Clock()
	for secs != 0 {
		time.Sleep(1 * time.Second)
		_, _, secs = time.Now().Clock()
		chronoCounter++
	}
	for {
		time.Sleep(1 * time.Minute)
//		hours, minutes, seconds := time.Now().Clock()

//		currentTime := time.Now()
//		tdat := fmt.Sprintf("%04d-%02d-%02d",
//			currentTime.Year(),
//			currentTime.Month(),
//			currentTime.Day())

		debugLog(5, fmt.Sprintf("Watchdog counter: %d", counter))
		if counter == old {
		}
		old = counter
		chronoCounter++
	}
}
