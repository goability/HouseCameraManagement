package main

import (
	"fmt"
	"os"

	"github.com/goability/ReplayAdminTools/ReplayFileUtilities"
)

const programName = "AdminPanel"

func main() {

	if len(os.Args) == 1 {
		help()

	} else {
		switch os.Args[1] {
		case "help":
			help()
			break
		case "StartHTTP":
			fmt.Println("Starting HTTP Service on Port 3000")
			ReplayFileUtilities.StartWebServer()
			break
		case "MoveNightFiles":
			ReplayFileUtilities.MoveNightFiles()
			break
		case "FixDateTimeErrors":
			fmt.Println("Fixing date time sync errors")
			ReplayFileUtilities.FixDateTimeErrors()
			break
		}
	}

}
func help() {
	fmt.Println("USAGE: " + programName + " {help}; {StartHTTP}; {MoveNightFiles {Folderbase;}}; {FixDateTimeErrors {Folderbase;}}")
}
