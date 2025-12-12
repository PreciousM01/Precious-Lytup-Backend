package main

import  "bytes"
import  "encoding/json"
import  "errors"
import  "fmt"
import  "io"
import  "log"
import  "net/http"
import  "regexp"
import  "runtime/debug"
import  "slices"
import  "strings"
import  "sync"
import  "time"

// Interface manager
func    DHI1 (Clap <-chan map[string]string, Flap chan <- map[string]string) (E error) {
	/***1***/
	xb05 := []*http.Server {}
	if err := DHI1ValidateCreateServers(&xb05, Flap); err != nil {
		return err
	}
	
	/***2***/
	xb10 := false
	xb15 := 0
	xb20 := &sync.Mutex{}
	for _ , xc10 := range xb05 {
		xc10.MaxHeaderBytes   = DHI0_MaxHeaderSize
		xc10.ReadTimeout      = DHI0_ReadTimeout
		xc10.ReadHeaderTimeout= DHI0_ReadTimeout
		xc10.WriteTimeout     = DHI0_WrttTimeout
		xc10.IdleTimeout      = DHI0_IdleTimeout
		xc15 := bytes.NewBuffer ([]byte {})
		xc10.ErrorLog = log.New (xc15 , "", log.Lshortfile)
		if xc10.Addr == DHI0_Addr1 {
			  DHI1StartServer(xc10, `HTTP`, &xb15, xb20, &xb10, &E)
		}  else {
			DHI1StartServer(xc10, `HTTPS`, &xb15, xb20, &xb10, &E)
		}
	}
	/***3***/
	for {
		select  {
			case _ = <- Clap: {
				xb10  = true
				for  _, xf10 := range xb05 { xf10.Close () }
				break
			}
			default:{
				if E != nil { return }
				xb20.Lock ( )
				xe05 := xb15
				xb20.Unlock ( )
				if xe05 == len(xb05) { return }
			}
		}
		time.Sleep (time.Millisecond*100)
	}
}

func DHI1ValidateCreateServers(srv *[]*http.Server, Flap chan<- map[string]string) (E error) {
	xb01 := map[string]string {}
	xb05 := []*http.Server {}

	// Server 1
	if DHI0_Addr1 != ""  {
		xc05 := &http.Server { Addr: DHI0_Addr1, Handler: &DHI2 {} }
		xb05  = append (xb05 , xc05)
	}

	// Server 2
	if DHI0_Addr2 != ""  {
		xc05 := &http.Server { Addr: DHI0_Addr2, Handler: &DHI2 {} }
		xb05  = append (xb05 , xc05)
	}

	// No servers configured
	if len (xb05) < 1 {
		xb01["StartupCode"] = "500"
		xb01["StartupNote"] =  fmt.Sprintf (`HTTP and HTTPS addresses not configured`)
		Flap <- xb01
		return
	}

	// Redirect Config Check
	if DHI0_RedirectHTTP&&
	   regexp.MustCompile(`^https\:\/\/.+$`).MatchString(DHI0_RedirectDestination) ==
	   false{
		xb01["StartupCode"] = "500"
		xb01["StartupNote"] =  fmt.Sprintf (
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

func DHI1StartServer (srv *http.Server, label string, serverCount *int, mutex *sync.Mutex, shutdownFlag *bool, E *error) {

	go func ( ) {
				// Update Server count
				defer func (  ) {
					mutex.Lock ( )
					*serverCount = *serverCount + 1
					mutex.Unlock ( )
				} ( )

				time.Sleep (time.Millisecond*100)

				xb05 := fmt.Sprintf (
					`%s interface listener started on %s`, label, srv.Addr)

				Output_Logg ("OUT", "DHI1", xb05)
				
				// Starting Server
				var xc05 error
				if label == "HTTPS" {
					xc05 = srv.ListenAndServeTLS (DHI0_Addr2_Crt, DHI0_Addr2_Key)
				} else {
					xc05 = srv.ListenAndServe ( )
				}

				// Error Handling
				mutex.Lock()
				if xc05 != nil && !(*shutdownFlag) {
					*E = errors.New (fmt.Sprintf (
						`%s interface listener unexpectedly shutdown [%s]`, label, xc05.Error (),
					))
				}
				mutex.Unlock()
			} ( )
}

// Panic manager
type    DHI2 struct {   }
func(d *DHI2)ServeHTTP (R http.ResponseWriter, r *http.Request) {
	/***1***/
	if r.TLS == nil&& DHI0_RedirectHTTP {
		http.Redirect (R, r, DHI0_RedirectDestination, http.StatusTemporaryRedirect)
	}
	/***2***/
	xb05 := map[string]any {}
	xb05["ExecutionOutcomeCode"] = 500
	defer func ( )  {
		/***1***/
		xc01 := recover()
		if xc01 != nil {
			xb05["ExecutionOutcomeCode"]= 500
			xb05["ExecutionOutcomeNote"]= fmt.Sprintf (
				`Panic sighted [%v : %s]`, xc01, string(debug.Stack ()),
			)
		}
		/***2***/
		xc05 := xb05["ExecutionOutcomeCode"].(int)
		if slices.Contains(DNI0_AllowedResponseCode,xc05) == false && xc05 != 500 {
			xb05["ExecutionOutcomeCode"]= 500
			xb05["ExecutionOutcomeNote"]= fmt.Sprintf (
				`Unexpected response code %d`, xc05,
			)
		}
		/***3***/
		if xb05["ExecutionOutcomeCode"].(int) == 500 {
			xd05  , xd10 := xb05["ExecutionOutcomeNote"].(string)
			if xd10 == false {
				xd05  = "Execution Outcome Note not a string"
			}
			Output_Logg ("ERR", "DHI2", xd05)
			delete (xb05, "ExecutionOutcomeNote")
		}
		/***4***/
		for _ , xd05 := range DNI0_ResponseHeaders {
			R.Header ().Set (xd05[0], xd05[1])
		}
		/***5***/
		xc10 ,_ := json.MarshalIndent (xb05, "", "    ")
		xc10  = append (xc10, '\n')
		R.Write (xc10)
	} ( )
	/***3***/
	xb15  , xb20 := io.ReadAll (r.Body)
	if xb20 != nil  {
		xb05["ExecutionOutcomeCode"] = 500
		xb05["ExecutionOutcomeNote"] = fmt.Sprintf (
			`Request read failed [%s]`, xb20.Error (),
		)
		return
	}
	if json.Valid(xb15) == false {
		xb05["ExecutionOutcomeCode"] = 400
		xb05["ExecutionOutcomeNote"] = fmt.Sprintf (`Request JSON formatting invalid`)
		return
	}
	xb25 :=&DHI0_Request {}
	xb30 := json.Unmarshal (xb15, xb25)
	if xb30 != nil {
		xb05["ExecutionOutcomeCode"] = 400
		xb05["ExecutionOutcomeNote"] = fmt.Sprintf (
			`Request unmarshal failed [%s]`, xb30.Error (),
		)
		return
	}
	if xb25.SrID == "" {
		xb05["ExecutionOutcomeCode"] = 400
		xb05["ExecutionOutcomeNote"] = fmt.Sprintf (`No service specified`)
		return
	}
	/***4***/
	xb35  , xb40, xb45 := DHI3 (r, xb25, R)
	xb05["ExecutionOutcomeCode"] = xb35
	xb05["ExecutionOutcomeNote"] = xb40
	if xb45 != nil {
		xb05["Yield"]=xb45
	}
}
// Router
func    DHI3 ( r *http.Request, s *DHI0_Request, R http.ResponseWriter) (
	C int, N string, Y any,
	)    {
	/***1***/
	C=500
	var ServiceProvider *DHI0_SP
	for _ , xc10:=range  DHI0_SPRegister {
		if s.SrID == xc10.Code { ServiceProvider = xc10 }
	}
	if  ServiceProvider == nil {
		C = 400
		N = fmt.Sprintf (`Service specified not supported`)
		return
	}
	/***2***/
	C , N , Y = ServiceProvider.Program (r, s.SrID, s.Seed)
	return
}
//============================================================================================//
//12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012//
//12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012//
//============================================================================================//
type    DHI0_Request struct {
	SrID string `json:"SrID"`
	Seed map[string]any `json:"Seed"`
}
type    DHI0_SP struct {
	Code    string
	Program func (*http.Request, string, map[string]any) (int, string, any)
}
