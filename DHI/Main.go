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
import  "syscall"


/* Initialize Daemoncore and start the application
 * Creates a DHI instance, assigns it to be the program of the daemon
 * Initialize daemon manager, starts execution and blocks unti an OS shutdown signal is received which shuts down the application.
*/
func    main () {
	
	/***1***/
	Output_Logg ("OUT", "Main", "PROJECT: Starting up")

	// If there are no daemons running, shut down.
	if DaemonRegister == nil{
		Output_Logg ("OUT", "Main", "PROJECT: No Daemon(s) to run. Shutting down now")
		return
	}

	// Load persistent cache on startup
	if err := GlobalWeatherCache.Load(); err != nil {
		Output_Logg("ERR", "Main", fmt.Sprintf("Failed to load cache: %s", err.Error()))
	}

	//Creating a new DHI object
	d := NewDHI()

	// Running the daemon with DHI
	DaemonRegister[0].Program = d.DHIStart

	manager := &DaemonManager{
		Daemons: 		DaemonRegister,
		SignalCh: 		make(chan os.Signal),
		StatusCh: 		make(chan bool),
		ShutdownSignal: SupportedShutdownSignal,
	}

	/***2***/
	defer func() {
		Output_Logg("OUT", "Main", "PROJECT: Initiating graceful shutdown")
		
		// Save cache before shutting down daemons
		if err := GlobalWeatherCache.Save(); err != nil {
			Output_Logg("ERR", "Main", fmt.Sprintf("Failed to save cache: %s", err.Error()))
		}
		
		// Shutdown daemons
		manager.DaemonShutDown()
		
		Output_Logg("OUT", "Main", "PROJECT: Shutdown complete")
	}()

	/***4***/
	go manager.DaemonStartUp()
	
	/***5***/
	go manager.Supervise(manager.SignalCh, manager.StatusCh)

	// Wait until all goroutines have shutdown
	<- manager.SignalCh
}

/* Gracefully stops all running daemons. 
 * Ensures reseources are released in a dependency-safe order by shutting down in a reverse order
*/
func (m *DaemonManager) DaemonShutDown() {
		xc05 := m.Daemons
		slices.Reverse(xc05)
		for _ , xd10 := range xc05 {
			if xd10.State != 1 { continue }
			xd15 := map[string]string{}
			xd15["Command"]= "shutdown"
			xd10.clap <-  xd15

			xd20 := fmt.Sprintf (
				`PROJECT: Daemon %s: Shutdown signalledd`, xd10.Name,
			)
			Output_Logg ("OUT", "Main", xd20)

			xd25 := make (chan bool, 1)
			if xd10.ShutdownGrace != 0{
				go func () {
					time.Sleep(xd10.ShutdownGrace)
					xd25 <- true
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
				case _ = <- xd25: {}
			}
			xd30 := fmt.Sprintf (
				`PROJECT: Daemon %s: Shutdown successful`, xd10.Name,
			)
			Output_Logg ("OUT", "Main", xd30)
		}
	} ;

/* Starts all registered daemons independently. Initializes communication (flap, clap) for each daemon.
 * Launches daemon execution  and monitors for startup success or failure
*/
func (m *DaemonManager) DaemonStartUp() {
		for _ , xc10 := range m.Daemons {
		/***1***/ // Daemon has no program running
		if xc10.Program == nil {
			xd05 := fmt.Sprintf (
				`PROJECT: Daemon %s: Skipping (Daemon has no program to run)`,
				xc10.Name ,
			)
			Output_Logg ("OUT", "Main", xd05)
			continue
		}
		/***2***/ // Starting up daemon
		xc10.clap = make (chan map[string]string, 1)
		xc10.flap = make (chan map[string]string, 1)
		xc15 := fmt.Sprintf (
			`PROJECT: Daemon %s: Starting up... Please wait`, xc10.Name,
		)
		Output_Logg ("OUT", "Main", xc15)

		/***3***/ // 
		xc20 := make (chan bool, 1)
		go m.DaemonRun(xc10, xc20) 

		/***4***/
		xc30 := make (chan bool , 1)
		go m.DaemonShutDownSignal(xc10, xc30)
		
	}
	}

/* Executes a single daemon program. 
 * Takes a daemon instance and a status channel as arguments.
 * Runs the daemon program, captures panics safely, updates state, and sends the execution status to the status channel.
 * The status channel is used to signal the completion of the daemon execution (true-> done, false -> not done)
*/

func (m *DaemonManager ) DaemonRun (daemon *Daemon, status chan bool) {

	defer func (  ) {
		xc05 := recover ( )
		if xc05 ==  nil { return }

		daemon.State = 2
		xc10 := fmt.Sprintf (
			`Paniced [%v : %s]`, xc05, debug.Stack (),
		)
		xc15 := map[string]string {}
		xc15 ["ExctnOtcmCode"] = "500"
		xc15 ["ExctnOtcmNote"] = xc10
		daemon.flap <- xc15
		status <- true
	} ( )

	daemon.State = 1
	xb05 := daemon.Program (daemon.clap, daemon.flap)
	daemon.State = 2
	xb10 := map[string]string {  }
	xb10 ["ExctnOtcmCode"] = "200"
	if xb05 != nil {
		xb10 ["ExctnOtcmCode"] = "500"
		xb10 ["ExctnOtcmNote"] = xb05.Error ()
	}
	daemon.flap <- xb10
	status <- true
} 

/* Helper function for DaemonStartUp. Handles Errors and signals during startup that indicate success or failure.
*/
func (m *DaemonManager) DaemonShutDownSignal(daemon *Daemon, status chan bool) {
	if daemon.StartupGrace != 0  {
			go func (  ) {
				time.Sleep (daemon.StartupGrace)
				status <- true
			}  (  )
		}
		select  {
			case xe05 := <- daemon.flap: {
				if xe05 ["StartupCode"] != "200" {
					xf05 := fmt.Sprintf (
						`PROJECT: Daemon %s: Startup failed [%s]`,
						daemon.Name , xe05 ["StartupNote"],
					)
					Output_Logg ("ERR", "Main", xf05)
					return
				}
				break
			}
			case _= <- status:{
				xe10 := fmt.Sprintf (
					`PROJECT: Daemon %s: Startup failed [%s]`,
					daemon.Name , "Startup grace period expired",
				)
				Output_Logg ("ERR", "Main", xe10)
				return
			}
		}
		
	}

/* Monitors all daemons and  listens for OS shutdown signals.
 * Takes a signal channel and a status channel as arguments.
 * Waits for shutdown signals and initiates the shutdown process for all running daemons.
 * The status channel is used to signal the completion of the shutdown process
*/
func (m *DaemonManager) Supervise(SigChannel chan os.Signal, status chan bool) { 
	for _ , xc10 := range SupportedShutdownSignal { signal.Notify (SigChannel , xc10) }
	for     {
		select  {
			case _= <-  status:{
				for _ , xf10 := range m.Daemons {
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
			case <- SigChannel:{
				Output_Logg("OUT", "Manager", "Shutdown signal received")
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
	Program  func  (<-  chan map[string]string, chan <- map[string]string) (error) // this function is DHI
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

/* Coordinates the lifecycle of all daemons. References shared  communication channels and supported OS shutdown signals
 * Controls startup, supervision, error handling, and graceful shutdown
*/
type DaemonManager struct {
	Daemons 		[]*Daemon
	SignalCh 		chan os.Signal
	StatusCh 		chan bool
	ShutdownSignal	[]syscall.Signal
}
