package main

import (
	"os/exec"
	"strings"
	"time"
)

func repairRoute() {
	time.Sleep(20 * time.Second) 	

	out, err := exec.Command("route").Output()
	if err != nil {
		traceLog("route command error: ")
		traceLog(err.Error())
	}
	tstr := string(out)

	astr := strings.Split(tstr, "\n")

    	// using for loop 
    	for i := 0; i < len(astr); i++ {
		if strings.HasPrefix(astr[i], "default") { 
        		traceLog(astr[i])
			if strings.HasPrefix(astr[i], "eth0") {
				exec.Command("route del default").Run()
			} 
		}
    	} 
}
