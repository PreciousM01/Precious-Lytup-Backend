package main

import  "syscall"

var     DaemonRegister []*Daemon = []*Daemon { }
var     SupportedShutdownSignal []syscall.Signal = []syscall.Signal {
	syscall.SIGHUP ,
	syscall.SIGINT ,
	syscall.SIGTERM,
}
var     TimeZoneSecondOffset int = 0
