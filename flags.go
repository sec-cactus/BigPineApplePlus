package main

import (
	"flag"
)

var (
	confFile string
)

/*
/	parseFlags parses all command line arguments.
*/
func parseFlags() {
	//Get flags
	flagConfFile := flag.String("c", "", "path to conf file")

	//Get Values
	flag.Parse()
	confFile = *flagConfFile
}

/*	flagsComplete
/	Returns true if all flags are satisfied and valid
*/
func flagsComplete() (allValid bool, err string) {
	allValid = true
	if confFile == "" {
		err = err + "Invalid conf file path.\n"
		allValid = false
	}

	return allValid, err
}
