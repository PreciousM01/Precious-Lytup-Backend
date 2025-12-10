package main

/*
	Name: DaemonCore-go
	Version: 0.0.3
*/

import  "fmt"
import  "os"
import  "os/signal"
import  "runtime/debug"
import  "slices"
import  "strings"
import  "time"

func    main () {
	/***1***/
	Output_Logg ("OUT", "Main", "PROJECT: Starting up")
	if DaemonRegister == nil || len(DaemonRegister) == 0 {
		Output_Logg ("OUT", "Main", "PROJECT: No Daemon(s) to run. Shutting down now")
		return
	}
	/***2***/
	defer   func () {
		xc05 := DaemonRegister
		slices.Reverse(xc05)
		for _ , xd10 := range xc05 {
			if xd10.State != 1 { continue }
			xd15 := map[string]string{}
			xd15["Command"]= "shutdown"
			xd10.clap <-  xd15
			xd17 := fmt.Sprintf (
				`PROJECT: Daemon %s: Shutdown signalledd`, xd10.Name,
			)
			Output_Logg ("OUT", "Main", xd17)
			xd20 := make (chan bool, 1)
			if xd10.ShutdownGrace != 0{
				go func () {
					time.Sleep(xd10.ShutdownGrace)
					xd20 <- true
				}  ( )
			}
			select  {
				case xf05 := <- xd10.flap: {
					if xf05["ExctnOtcmCode"] != "200" {
						xg05 := fmt.Sprintf (
							`PROJECT: Daemon %s: Encountered error [%s]`, xd10.Name, xf05["ExctnOtcmNote"],
						)
						Output_Logg ("ERR", "Main", xg05)
						return
					}
				}
				case _ = <- xd20: {}
			}
			xd25 := fmt.Sprintf (
				`PROJECT: Daemon %s: Shutdown successful`, xd10.Name,
			)
			Output_Logg ("OUT", "Main", xd25)
		}
	} ( )
	/***3***/
	xb05 := make (chan bool , 1)
	for _ , xc10 := range DaemonRegister {
		/***1***/
		if xc10.Program == nil {
			xd05 := fmt.Sprintf (
				`PROJECT: Daemon %s: Skipping (Daemon has no program to run)`,
				xc10.Name ,
			)
			Output_Logg ("OUT", "Main", xd05)
			continue
		}
		/***2***/
		xc10.clap = make (chan map[string]string, 1)
		xc10.flap = make (chan map[string]string, 1)
		xc25 := fmt.Sprintf (
			`PROJECT: Daemon %s: Starting up... Please wait`, xc10.Name,
		)
		Output_Logg ("OUT", "Main", xc25)
		/***3***/
		go func ( ) {
			defer func (  ) {
				xe05 := recover ( )
				if xe05 ==  nil { return }
				xc10.State = 2
				xe10 := fmt.Sprintf (
					`Paniced [%v : %s]`, xe05, debug.Stack (),
				)
				xe15 := map[string]string {}
				xe15 ["ExctnOtcmCode"] = "500"
				xe15 ["ExctnOtcmNote"] = xe10
				xc10.flap <- xe15
				xb05 <- true
			} ( )
			xc10.State = 1
			xd05 := xc10.Program (xc10.clap, xc10.flap)
			xc10.State = 2
			xd10 := map[string]string {  }
			xd10 ["ExctnOtcmCode"] = "200"
			if xd05 != nil {
				xd10 ["ExctnOtcmCode"] = "500"
				xd10 ["ExctnOtcmNote"] = xd05.Error ()
			}
			xc10.flap <- xd10
			xb05 <- true
		}  (    )
		/***4***/
		xc30 := make (chan bool , 1)
		if xc10.StartupGrace != 0  {
			go func (  ) {
				time.Sleep (xc10.StartupGrace)
				xc30 <- true
			}  (  )
		}
		select  {
			case xe05 := <- xc10.flap: {
				if xe05 ["StartupCode"] != "200" {
					xf05 := fmt.Sprintf (
						`PROJECT: Daemon %s: Startup failed [%s]`,
						xc10.Name , xe05 ["StartupNote"],
					)
					Output_Logg ("ERR", "Main", xf05)
					return
				}
				break
			}
			case _= <- xc30:{
				xe10 := fmt.Sprintf (
					`PROJECT: Daemon %s: Startup failed [%s]`,
					xc10.Name , "Startup grace period expired",
				)
				Output_Logg ("ERR", "Main", xe10)
				return
			}
		}
		/***5***/
		xc35 := fmt.Sprintf (`PROJECT: Daemon %s: Up and running`, xc10.Name)
		Output_Logg ("OUT", "Main", xc35)
	}
	/***4***/
	xb10 := make (chan os.Signal, 1 )
	for _ , xc10 := range SupportedShutdownSignal { signal.Notify (xb10 , xc10) }
	for     {
		select  {
			case _= <- xb05:{
				for _ , xf10 := range DaemonRegister {
					select  {
						case xh10 := <- xf10.flap: {
							if xh10["ExctnOtcmCode"] != "200" {
								xi05 := fmt.Sprintf (
									`PROJECT: Daemon %s: Encountered error [%s]`, xf10.Name , xh10["ExctnOtcmNote"],
								)
								Output_Logg (
									"ERR", "Main", xi05,
								)
								return
							}
							xh15 := fmt.Sprintf (
								`PROJECT: Daemon %s: Finished`, xf10.Name,
							)
							Output_Logg ("OUT", "Main", xh15)
						}
						default: {}
					}
				}
			}
			case _= <- xb10:{
				fmt.Println ("")
				return
			}
		}
	}
}
//============================================================================================//
//12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012//
//12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012//
//============================================================================================//
type    Daemon struct  {
	Name   string
	Program  func  (<-  chan map[string]string, chan <- map[string]string) (error)
	State  uint64  // 0 - Initial; 1 - Running; 2 - Done
	StartupGrace     time.Duration
	ShutdownGrace    time.Duration
	// internal use: don't set properties below
	clap   chan map[string]string
	flap   chan map[string]string
}
func    Output_Logg (Type, Source, Output string) {
	Type  = strings.ToLower (Type)
	xb05 := fmt.Sprintf (
		`[%s//%s] %s`, time.Now ().In (time.FixedZone("TTT", TimeZoneSecondOffset)).Format ("2006-01-02 15:04:05.000 -07:00"), Source, Output,
	)
	if Type == "out" {
		os.Stdout.Write ([ ]byte (xb05 + "\n") )
	}  else {
		os.Stderr.Write ([ ]byte (xb05 + "\n") )
	}
}
