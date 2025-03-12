package main

import "os"

var step = 2

func main() {

	// Capture data.
	err := capture(pathData)
	if err != nil {
		os.Exit(1)
	}

	// Analyze data.
	err = analysis(pathData)
	if err != nil {
		os.Exit(1)
	}
}
