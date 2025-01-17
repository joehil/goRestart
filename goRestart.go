package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/nxadm/tail"
	"github.com/spf13/viper"
)

var do_trace bool = false
var msg_trace bool = false
var counter uint64 = 0
var chronoCounter = 0
var chronoOld = 0
var logseverity int
var pidfile string
var ownlog string
var logs []string
var topics []string
var timeOld time.Time
var dumpfile string
var dfile *os.File

var genVar Generalvars

//var ptrGenVars *Generalvars

func main() {
	// Set location of config
	viper.SetConfigName("goRestart") // name of config file (without extension)
	viper.AddConfigPath("/etc/")     // path to look for the config file in

	// Read config
	read_config()

	timeOld = time.Now()

	// Get commandline args
	if len(os.Args) > 1 {
		a1 := os.Args[1]
		if a1 == "reload" {
			b, err := os.ReadFile(pidfile)
			if err != nil {
				log.Fatal(err)
			}
			s := string(b)
			fmt.Println("Reload", s)
			cmd := exec.Command("kill", "-hup", s)
			_ = cmd.Start()
			os.Exit(0)
		}
		if a1 == "stop" {
			b, err := os.ReadFile(pidfile)
			if err != nil {
				log.Fatal(err)
			}
			s := string(b)
			fmt.Println("Stop goRestart")
			cmd := exec.Command("kill", "-9", s)
			_ = cmd.Start()
			os.Exit(0)
		}
		if a1 == "run" {
			go repairRoute()
			procRun()
			os.Exit(0)
		}
                if a1 == "route" {
                        repairRoute()
                        os.Exit(0)
                }
		fmt.Println("parameter invalid")
		os.Exit(-1)
	}
	if len(os.Args) == 1 {
		myUsage()
	}
}

func procRun() {
	// Write pidfile
	err := writePidFile(pidfile)
	if err != nil {
		log.Fatalf("Pidfile could not be written: %v", err)
	}
	defer os.Remove(pidfile)

	genVar.Telegram = make(chan string)

	go sendTelegram(genVar.Telegram)
	traceLog("Telegram interface was initialized")

	go checkNetwork()
	traceLog("network checking was initialized")

	go timeTrigger()
	traceLog("chrono server was initialized")

	// Open log file
	ownlogger := &lumberjack.Logger{
		Filename:   ownlog,
		MaxSize:    1, // megabytes
		MaxBackups: 7,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	defer ownlogger.Close()
	log.SetOutput(ownlogger)

	// Inform about trace
	log.Println("Trace set to: ", do_trace)

	// Catch signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGTERM)
	go catch_signals(signals)

	// Open logs to read
	if do_trace {
		log.Println(logs)
	}

	for _, rlog := range logs {
		traceLog("Task started for " + rlog)
		go tailLog(rlog)
	}

	traceLog("goRestart is up and running")

	for {
		time.Sleep(60 * time.Second)
		if chronoCounter == chronoOld {
//			var mInfo Msginfo
//			mInfo.Msgevent = "watchdog.event"
//			mInfo.Msgobject = "Watchdog"
//			go processRulesInfo(mInfo)
		}
		chronoOld = chronoCounter
	}
}

func procLine(msg string) {
	var mInfo Msginfo
	var mWarn Msgwarn
	if len(msg) > 75 {
		msgType := msg[25:29]
		if msgType == "INFO" {
			mInfo.Msgdate = msg[0:10]
			mInfo.Msgtime = msg[11:23]
			mInfo.Msgevent = strings.Trim(msg[33:69], " ")
			rest := msg[73:]
			mes := strings.Split(rest, " ")
			if mInfo.Msgevent == "openhab.event.ItemStateChangedEvent" {
				if len(mes) == 7 {
					mInfo.Msgobjtype = mes[0]
					mInfo.Msgobject = strings.Trim(mes[1], "' ")
					mInfo.Msgoldstate = mes[4]
					mInfo.Msgnewstate = mes[6]
				}
				if len(mes) == 9 {
					mInfo.Msgobjtype = mes[0]
					mInfo.Msgobject = strings.Trim(mes[1], "' ")
					mInfo.Msgoldstate = strings.Join(mes[4:5], " ")
					mInfo.Msgnewstate = strings.Join(mes[7:8], " ")
				}
			}
			if mInfo.Msgevent == "openhab.event.ChannelTriggeredEvent" {
				if len(mes) >= 3 {
					mInfo.Msgobject = strings.Trim(mes[0], "' ")
					mInfo.Msgnewstate = mes[2]
				}
			}
//			processRulesInfo(mInfo)
			counter++
		}
		if msgType == "WARN" {
			mWarn.Msgdate = msg[0:10]
			mWarn.Msgtime = msg[11:23]
			mWarn.Msgevent = msg[33:69]
			mWarn.Msgtext = msg[73:]
		}
	}
}

// Write a pid file, but first make sure it doesn't exist with a running pid.
func writePidFile(pidFile string) error {
	// Read in the pid file as a slice of bytes.
	if piddata, err := os.ReadFile(pidFile); err == nil {
		// Convert the file contents to an integer.
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			// Look for the pid in the process list.
			if process, err := os.FindProcess(pid); err == nil {
				// Send the process a signal zero kill.
				if err := process.Signal(syscall.Signal(0)); err == nil {
					// We only get an error if the pid isn't running, or it's not ours.
					return fmt.Errorf("pid already running: %d", pid)
				}
			}
		}
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}

func catch_signals(c <-chan os.Signal) {
	for {
		s := <-c
		log.Println("Got signal:", s)
		if s == syscall.SIGHUP {
			read_config()
		}
		if s == syscall.SIGUSR1 {
			var err error
			msg_trace = true
			log.Println("msg_trace switched on")
			dfile, err = os.Create(dumpfile)
			if err != nil {
				traceLog(fmt.Sprintf("failed creating dumpfile: %s", err))
			}
		}
		if s == syscall.SIGUSR2 {
			msg_trace = false
			log.Println("msg_trace switched off")
			dfile.Close()
		}
		if s == syscall.SIGTERM {
			fmt.Println("goRestart stopped")
			os.Exit(0)
		}
	}
}

func read_config() {
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatalf("Config file not found: %v", err)
	}

	pidfile = viper.GetString("pid_file")
	if pidfile == "" { // Handle errors reading the config file
		log.Fatalf("Filename for pidfile unknown: %v", err)
	}
	ownlog = viper.GetString("own_log")
	if ownlog == "" { // Handle errors reading the config file
		log.Fatalf("Filename for ownlog unknown: %v", err)
	}
	logs = viper.GetStringSlice("logs")
	do_trace = viper.GetBool("do_trace")
	genVar.Tbtoken = viper.GetString("tbtoken")
	genVar.Chatid = int64(viper.GetInt("chatid"))
	logseverity = viper.GetInt("log_severity")

	genVar.machineNet = viper.GetString("machine_net")
	genVar.localNet = viper.GetString("local_net")
	genVar.interNet = viper.GetString("inter_net")

	if do_trace {
		log.Println("do_trace: ", do_trace)
		log.Println("own_log; ", ownlog)
		log.Println("pid_file: ", pidfile)
		log.Println("Dumpfile: ", dumpfile)
		log.Println("Logseverity: ", logseverity)

		for i, v := range logs {
			log.Printf("Index: %d, Value: %v\n", i, v)
		}
		for i, v := range topics {
			log.Printf("Index: %d, Value: %v\n", i, v)
		}
	}
}

func tailLog(logFile string) {
	t, err := tail.TailFile(logFile, tail.Config{Follow: true})
	if err != nil {
		panic(err)
	}
	for line := range t.Lines {
		tNow := time.Now()
		if tNow.Sub(timeOld) > 10*time.Second {
			go procLine(line.Text)
		}
	}
}

func myUsage() {
	fmt.Printf("Usage: %s argument\n", os.Args[0])
	fmt.Println("Arguments:")
	fmt.Println("run           Run progam as daemon")
	fmt.Println("reload        Make running daemon reload it's configuration")
	fmt.Println("stop          Stop daemon")
	fmt.Println("sim           Test rules by reading the dump file")
}
