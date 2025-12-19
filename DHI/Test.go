package main

import  "fmt"
import  "net/http"
func    init (   ) {
	DHI0_Addr1 = ":8080"
	DHI0_Addr2 = ":8443"
	DHI0_RedirectDestination = "https://localhost:8443"
	DHI0_SPRegister= [ ]*DHI0_SP {
		{ Code: "weather", Program: SPWeatherForecast },

	}

	// Start cache cleanup
	GlobalWeatherCache.StartCleanup()
	
	// Print what's registered
	fmt.Println("=== Services Registered ===")
	for _, sp := range DHI0_SPRegister {
		fmt.Printf("Service: %s\n", sp.Code)
	}
	fmt.Println("===========================")
}

func    SP01 (R *http.Request, r string, s map[string]any) (C int, N string, Y any) {
	C = 200
	N = fmt.Sprintf (`Data`)
	Y = map[string]any {
		"Data01": "_",
		"Data02": 104,
	}
return
}
