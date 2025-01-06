package main

type Msginfo struct {
	Msgdate     string
	Msgtime     string
	Msgevent    string
	Msgobjtype  string
	Msgobject   string
	Msgoldstate string
	Msgnewstate string
}

type Msgwarn struct {
	Msgdate  string
	Msgtime  string
	Msgevent string
	Msgtext  string
}

type Generalvars struct {
	Telegram      chan string
	Tbtoken       string
	Chatid        int64
	machineNet    string
	localNet      string
	interNet      string
}
