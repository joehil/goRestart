package main

import (
	"fmt"
	"net/http"
	"time"
)

func isNetworkAvailable(url string) int {
	resp, err := http.Get(url)
	if err != nil {
		return int(999)
	}
	defer resp.Body.Close()

	return http.StatusOK
}

func checkNetwork() {
	var avalLocal int
	var avalMachine int
	var avalInternet int
	var cntMachine int = 0
	var cntLocal int = 0
	var cntInternet int = 0
	for {
		avalMachine = isNetworkAvailable(genVar.machineNet)
		time.Sleep(10 * time.Second)
		avalLocal = isNetworkAvailable(genVar.localNet)
		time.Sleep(10 * time.Second)
		avalInternet = isNetworkAvailable(genVar.interNet)
		if avalMachine != 200 {
			traceLog(fmt.Sprintf("Network availability error: %s %d", genVar.machineNet, avalMachine))
			cntMachine++
			if cntMachine > 3 {
				fmt.Printf("%s code: %d\n", genVar.machineNet, avalMachine)
				restartNetwork()
			}
		} else {
			cntMachine = 0
		}
		if avalLocal != 200 {
			traceLog(fmt.Sprintf("Network availability error: %s %d", genVar.localNet, avalLocal))
			cntLocal++
			if cntLocal > 3 {
				fmt.Printf("%s code: %d\n", genVar.localNet, avalLocal)
				restartNetwork()
			}
		} else {
			cntLocal = 0
		}
		if avalInternet != 200 {
			traceLog(fmt.Sprintf("Network availability error: %s %d", genVar.interNet, avalInternet))
			cntInternet++
			if cntInternet > 3 {
				fmt.Printf("%s code: %d\n", genVar.interNet, avalInternet)
				restartNetwork()
			}
		} else {
			cntInternet = 0
		}
		if avalMachine == 200 && avalLocal == 200 && avalInternet == 200 {
			traceLog("Network availability ok")
		}
		time.Sleep(2 * time.Minute)
	}
}
