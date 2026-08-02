package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("eng version " + version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`eng - Engineering Memory CLI

Usage:
  eng <command>

Commands:
  version   print the eng version

Planned, not yet implemented (see CLI.md):
  init, add, index, search, ask, status, doctor`)
}
