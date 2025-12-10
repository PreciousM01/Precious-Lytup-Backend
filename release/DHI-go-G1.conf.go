package main

import  "time"

var     DHI0_Addr1 string = ":80"
var     DHI0_Addr2 string = ":443"
var     DHI0_Addr2_Key string = "tls.key"
var     DHI0_Addr2_Crt string = "tls.crt"
var     DHI0_RedirectHTTP bool = true
var     DHI0_RedirectDestination string = "https://localhost"
var     DHI0_MaxHeaderSize int = 1*1024 * 1024
var     DHI0_ReadTimeout time.Duration  = time.Minute*5
var     DHI0_WrttTimeout time.Duration  = time.Minute*5
var     DHI0_IdleTimeout time.Duration  = time.Minute*5
var     DHI0_SPRegister []*DHI0_SP = [ ]*DHI0_SP { }
var     DNI0_AllowedResponseCode []int  = [ ]int { 500, 400, 406, 200 }
var     DNI0_ResponseHeaders [][]string = [ ][ ] string {
	[ ]string {"Content-Type", "application/json"} ,
}
