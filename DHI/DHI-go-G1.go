package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"runtime/debug"
	"slices"
	"sync"
	"time"
)

// Interface manager structure
type DHI struct {
	// Config attributes
	Addr1                string
	Addr2                string
	RedirectHTTP         bool
	RedirectDestination  string
	MaxHeaderSize        int
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	IdleTimeout          time.Duration
	SPRegister           []*DHI0_SP
	AllowedResponseCode []int
	ResponseHeaders      [][]string
	TLSCert              string
	TLSKey               string

	//Runtime shared state
	Servers      []*http.Server
	ShutdownFlag bool        // shared across goroutines
	Mutex        sync.Mutex // protects ShutdownFlag
}

// Create a new DHI instance, and initialize it.
func NewDHI() *DHI{
	return &DHI{
	Addr1:                DHI0_Addr1,
	Addr2:                DHI0_Addr2,
	RedirectHTTP:         DHI0_RedirectHTTP,
	RedirectDestination:  DHI0_RedirectDestination,
	MaxHeaderSize:        DHI0_MaxHeaderSize,
	ReadTimeout:          DHI0_ReadTimeout,
	WriteTimeout:         DHI0_WrttTimeout,
	IdleTimeout:          DHI0_IdleTimeout,
	SPRegister:           DHI0_SPRegister,
	AllowedResponseCode: DNI0_AllowedResponseCode,
	ResponseHeaders:      DNI0_ResponseHeaders,
	TLSCert:              DHI0_Addr2_Crt, 
	TLSKey:               DHI0_Addr2_Key,

	Servers:      []*http.Server{},
	ShutdownFlag: false,   
	Mutex:        sync.Mutex{},

	}
}

/* Start the DHI interface
 * Takes Clap and Flap channels as input
 * Returns an error if any
*/
func (d *DHI) DHIStart (Clap <-chan map[string]string, Flap chan<- map[string]string) (E error) {
	/***1***/
	if err := d.DHI1ValidateCreateServers(Flap); err != nil {
		return err
	}

	/***2***/
	xb05 := 0
	for _, xc10 := range d.Servers {
		xc10.MaxHeaderBytes = d.MaxHeaderSize
		xc10.ReadHeaderTimeout = d.ReadTimeout
		xc15 := bytes.NewBuffer([]byte{})
		xc10.ErrorLog = log.New(xc15, "", log.Lshortfile)
		if xc10.Addr == DHI0_Addr1 {
			d.DHI1StartServer(xc10, `HTTP`, &xb05, &E)
		} else {
			d.DHI1StartServer(xc10, `HTTPS`, &xb05, &E)
		}
	}

	/***3***/
	return d.DHI1_WaitForShutdown(Clap, &xb05, &E)
}

/* Create and configure servers.
 * Takes Flap channel as input
 * Returns an error if any
*/
func (d *DHI) DHI1ValidateCreateServers(Flap chan<- map[string]string) (E error) {
	xb01 := map[string]string{}

	// Server 1
	if d.Addr1 != "" {
		xc05 := &http.Server{Addr: d.Addr1, Handler: d}
		d.Servers = append(d.Servers, xc05)
	}

	// Server 2
	if d.Addr2 != "" {
		xc05 := &http.Server{Addr: d.Addr2, Handler: d}
		d.Servers = append(d.Servers, xc05)
	}

	// No servers configured
	if len(d.Servers) < 1 {
		xb01["StartupCode"] = "500"
		xb01["StartupNote"] = fmt.Sprintf(`HTTP and HTTPS addresses not configured`)
		Flap <- xb01
		return
	}

	// Redirect Config Check
	if d.RedirectHTTP &&
		regexp.MustCompile(`^https\:\/\/.+$`).MatchString(d.RedirectDestination) ==
			false {
		xb01["StartupCode"] = "500"
		xb01["StartupNote"] = fmt.Sprintf(
			`Conf parameter DHI0_RedirectDestination not valid`,
		)
		Flap <- xb01
		return
	}

	// Successful configuration (*)
	xb01["StartupCode"] = "200"
	xb01["StartupNote"] = fmt.Sprintf(`OK`)
	Flap <- xb01
	return
}

/* Starts servers and establish communication channel
 * Takes server, label, serverCount and E as input
 */
func (d *DHI) DHI1StartServer(srv *http.Server, label string, serverCount *int, E *error) {

	go func() {
		// Update Server count
		defer func() {
			d.Mutex.Lock()
			*serverCount = *serverCount + 1
			d.Mutex.Unlock()
		}()

		time.Sleep(time.Millisecond * 100)

		xb05 := fmt.Sprintf(
			`%s interface listener started on %s`, label, srv.Addr)

		Output_Logg("OUT", "DHI1", xb05)

		// Starting Server
		var xc05 error
		if label == "HTTPS" {
			xc05 = srv.ListenAndServeTLS(d.TLSCert, d.TLSKey)
		} else {
			xc05 = srv.ListenAndServe()
		}

		// Error Handling
		d.Mutex.Lock()
		if xc05 != nil && !(d.ShutdownFlag) {
			*E = errors.New(fmt.Sprintf(
				`%s interface listener unexpectedly shutdown [%s]`, label, xc05.Error(),
			))
		}
		d.Mutex.Unlock()
	}()
}

/* Keeps the interface running until it is told to stop or a failure occurs. Closes servers when shutddown is requested.
 * Takes Clap, serverCount and E as input
 * Returns an error if any, else returns nil
 */
func (d *DHI) DHI1_WaitForShutdown(Clap <-chan map[string]string, serverCount *int, E *error) error {

	for {
		select {

		//Shutdown command received
		case <-Clap:
			d.Mutex.Lock()
			d.ShutdownFlag = true
			d.Mutex.Unlock()

			for _, srv := range d.Servers {
				srv.Close()
			}

		default:

			d.Mutex.Lock()
			err := *E
			done := (*serverCount == len(d.Servers))
			d.Mutex.Unlock()

			if err != nil {
				return err
			}

			// If all servers are finished
			if done {
				return nil
			}
		}

		time.Sleep(time.Millisecond * 100)
	}
}

/* Http request handler for the interface servers (Panic manager). 
 * Takes http.ResponseWriter and *http.Request as input
 */
func (d *DHI) ServeHTTP(R http.ResponseWriter, r *http.Request) {
	/***1***/
	if r.TLS == nil && d.RedirectHTTP {
		http.Redirect(R, r, d.RedirectDestination, http.StatusTemporaryRedirect)
	}
	/***2***/
	xb05 := map[string]any{}
	xb05["ExecutionOutcomeCode"] = 500
	defer func() {
		/***1***/
		xc01 := recover()
		if xc01 != nil {
			xb05["ExecutionOutcomeCode"] = 500
			xb05["ExecutionOutcomeNote"] = fmt.Sprintf(
				`Panic sighted [%v : %s]`, xc01, string(debug.Stack()),
			)
		}
		/***2***/
		xc05 := xb05["ExecutionOutcomeCode"].(int)
		if slices.Contains(d.AllowedResponseCode, xc05) == false && xc05 != 500 {
			xb05["ExecutionOutcomeCode"] = 500
			xb05["ExecutionOutcomeNote"] = fmt.Sprintf(
				`Unexpected response code %d`, xc05,
			)
		}
		/***3***/
		if xb05["ExecutionOutcomeCode"].(int) == 500 {
			xd05, xd10 := xb05["ExecutionOutcomeNote"].(string)
			if xd10 == false {
				xd05 = "Execution Outcome Note not a string"
			}
			Output_Logg("ERR", "DHI2", xd05)
			delete(xb05, "ExecutionOutcomeNote")
		}
		/***4***/
		for _, xd05 := range d.ResponseHeaders {
			R.Header().Set(xd05[0], xd05[1])
		}
		/***5***/
		xc10, _ := json.MarshalIndent(xb05, "", "    ")
		xc10 = append(xc10, '\n')
		R.Write(xc10)
	}()
	/***3***/
	xb15, xb20 := io.ReadAll(r.Body)
	if xb20 != nil {
		xb05["ExecutionOutcomeCode"] = 500
		xb05["ExecutionOutcomeNote"] = fmt.Sprintf(
			`Request read failed [%s]`, xb20.Error(),
		)
		return
	}
	if json.Valid(xb15) == false {
		xb05["ExecutionOutcomeCode"] = 400
		xb05["ExecutionOutcomeNote"] = fmt.Sprintf(`Request JSON formatting invalid`)
		return
	}
	xb25 := &DHI0_Request{}
	xb30 := json.Unmarshal(xb15, xb25)
	if xb30 != nil {
		xb05["ExecutionOutcomeCode"] = 400
		xb05["ExecutionOutcomeNote"] = fmt.Sprintf(
			`Request unmarshal failed [%s]`, xb30.Error(),
		)
		return
	}
	if xb25.SrID == "" {
		xb05["ExecutionOutcomeCode"] = 400
		xb05["ExecutionOutcomeNote"] = fmt.Sprintf(`No service specified`)
		return
	}
	/***4***/
	xb35, xb40, xb45 := d.Route(r, xb25, R)
	xb05["ExecutionOutcomeCode"] = xb35
	xb05["ExecutionOutcomeNote"] = xb40
	if xb45 != nil {
		xb05["Yield"] = xb45
	}
}

/* Selects the correct service provider for a request and executes it (Router).
 * Takes as input the request, the service provider ID and the seed.
 * Returns the response code, note and yield
*/
func (d *DHI) Route(r *http.Request, s *DHI0_Request, R http.ResponseWriter) (
	C int, N string, Y any,
) {
	/***1***/
	C = 500
	var ServiceProvider *DHI0_SP
	for _, xc10 := range d.SPRegister {
		if s.SrID == xc10.Code {
			ServiceProvider = xc10
		}
	}
	if ServiceProvider == nil {
		C = 400
		N = fmt.Sprintf(`Service specified not supported`)
		return
	}
	/***2***/
	C, N, Y = ServiceProvider.Program(r, s.SrID, s.Seed)
	return
}

// ============================================================================================//
// 12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012//
// 12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012//
// ============================================================================================//
type DHI0_Request struct {
	SrID string         `json:"SrID"`
	Seed map[string]any `json:"Seed"`
}
type DHI0_SP struct {
	Code    string
	Program func(*http.Request, string, map[string]any) (int, string, any)
}


