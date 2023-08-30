package main

import (
	"encoding/json"
	"os"
	"proj2/server"
	"strconv"
	// "runtime"
)


func main() {
	args := os.Args

	// runtime.GOMAXPROCS(2)
	
	var mode string
	var nConsumers int

	
	// retrieve the number of consumers from the command line
	if len(args) != 2 {
		nConsumers = 1
	} else {
		nConsumers, _ = strconv.Atoi(args[1])
	}
	 
	// set the mode
	if nConsumers > 1 {
		mode = "p"
	} else {
		mode = "s"
	}
	
	// encoder to send responses to the client via os.stdout
	enc := json.NewEncoder(os.Stdout)
	// decoder to read requests from the client via os.stdin
	dec := json.NewDecoder(os.Stdin)

	// create server configuration
	conf := server.Config {
		Encoder: enc,
		Decoder: dec,
		Mode: mode,
		ConsumersCount: nConsumers,
	}
	
	// deploy the server
	server.Run(conf)
}
