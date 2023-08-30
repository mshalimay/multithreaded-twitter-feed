package main

import (
	"encoding/json"
	"os"
	"proj2/server"
	// "strconv"
)


func main() {
	// args := os.Args


	// file, _ := os.Open("/mnt/g/My Drive/!Projects/ParallelProgramming/Projects/Project1/project-1/proj1/twitter/test.txt")
	// defer file.Close()

	//
	enc := json.NewEncoder(os.Stdout)
	//dec := json.NewDecoder(os.Stdin)
	dec := json.NewDecoder(os.Stdin)

	// create server configuration
	conf := server.Config {
		Encoder: enc,
		Decoder: dec,
		Mode: "s",
		ConsumersCount: 1,
	}
	
	// deploy the server
	server.Run(conf)
	
}
