package main

import (
	"github.com/ulrichwisser/dnssectiming/cmd"
)

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
)

func init() {
	// Set default log handler
	log.SetHandler(logfmt.New(os.Stderr))
	log.SetLevel(log.DebugLevel)
}

func main() {
	cmd.Execute()
}
