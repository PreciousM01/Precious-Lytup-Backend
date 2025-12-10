package main

import "time"

func    init () {
	DaemonRegister = []*Daemon {
		&Daemon {
			Name: "Daemon 01", Program: Daemon01, StartupGrace: time.Second*5,
			ShutdownGrace:time.Second*5,
		},
		&Daemon {
			Name: "Daemon 02", Program: Daemon02, StartupGrace: time.Second*5,
			ShutdownGrace:time.Second*5,
		},
	}
	TimeZoneSecondOffset = 2*60*60
}
func    Daemon01 (Clap <-chan map[string]string, Flap chan <- map[string]string) (E error) {
	xb05 := map[string]string{ }
	xb05["StartupCode"] = "200"
	xb05["StartupNote"] = "We're good here man"
	Flap <- xb05
	_  = <- Clap
return
}
func    Daemon02 (Clap <-chan map[string]string, Flap chan <- map[string]string) (E error) {
	xb05 := map[string]string{ }
	xb05["StartupCode"] = "500"
	xb05["StartupNote"] = "We're good here man"
	Flap <- xb05
	time.Sleep (time.Second*5)
return
}
