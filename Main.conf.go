package main

import  "syscall"

var     DaemonRegister []*Daemon = [ ]*Daemon {
	&Daemon { Name: "DHI0", Program: DHI1 },
}
var     SupportedShutdownSignal []syscall.Signal = []syscall.Signal {
	syscall.SIGHUP ,
	syscall.SIGINT ,
	syscall.SIGTERM,
}
var     TimeZoneSecondOffset int = 0
