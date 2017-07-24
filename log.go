package main

import (
	"log"
	"os"
	"github.com/fatih/color"
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func init() {

	Info = log.New(os.Stdout,
		color.GreenString("[INFO] "),
		log.Lshortfile|log.Ltime)

	Warn = log.New(os.Stdout,
		color.GreenString("[WARN] "),
		log.Lshortfile|log.Ltime)

	Error = log.New(os.Stderr,
		color.GreenString("[ERROR] "),
		log.Lshortfile|log.Ltime)

}
