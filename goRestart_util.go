package main

import (
//	"encoding/json"
//	"fmt"
	"log"
//	"net/http"
	"os/exec"
//	"strconv"
	"time"
)

func exec_cmd(command string, args ...string) {
	cmd := exec.Command(command, args...)
	err := cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}
}

func traceLog(message string) {
	if do_trace {
		log.Println(message)
	}
}

func debugLog(sev int, message string) {
	if logseverity >= sev {
		log.Println(message)
	}
}

func restartNetwork() {
	cmd1 := exec.Command("/usr/bin/sudo", "/usr/sbin/service", "networking", "stop")
	cmd1.Run()
	time.Sleep(10 * time.Second)
	cmd2 := exec.Command("/usr/bin/sudo", "/usr/sbin/service", "networking", "start")
	cmd2.Run()
}

func reboot() {
	cmd := exec.Command("/usr/bin/sudo", "/usr/sbin/reboot")
	cmd.Run()
}

