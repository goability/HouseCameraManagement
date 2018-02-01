/*  Move night files

-
(c) Matt Chandler, 2018
*/

package ReplayFileUtilities

//A change
import (
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

func MoveNightFiles() {
	if runtime.GOOS == "windows" {
		setWindowsPaths()
	} else {
		nightBaseFolder = filepath.Join(baseSearchFolder, camFolderName)
		baseSearchFolder = filepath.Join(baseSearchFolder, camFolderName)
	}

	if RUNASWEBSERVER {
		startWebServer()
	} else {
		showStart()

		//Walk the baseSearchFolder
		err := filepath.Walk(baseSearchFolder, walkFuncMoveNightFiles)
		if err != nil {
			fmt.Print("ERROR Walking folder:  ")
			fmt.Print(baseSearchFolder)
			fmt.Println(" ERR:", err)
		} else {
			showSummaryMoveNightFiles()
		}
	}
}

func showSummaryMoveNightFiles() {
	fmt.Println("\n\n--------- FINISHED -------")

	if DEBUG == true {
		fmt.Println("DEBUG MODE")
		fmt.Println("-------------------")
		fmt.Printf("\nTotal Bytes Discovered: %d", totalBytesDiscovered)
		fmt.Printf("\nTotal Files Discovered: %d ", totalFilesDiscovered)

	} else {
		fmt.Println("FILES WERE COPIED to: ", nightBaseFolder)
		fmt.Printf("\nTotal Bytes Moved: %d", totalBytesMoved)
		fmt.Printf("\nTotal Files Moved: %d ", totalFilesMoved)
	}
	fmt.Printf("\n SIZE: ")
	if totalBytesDiscovered > 1000000000 {
		fmt.Printf(" %d GB", totalBytesDiscovered/(1024*1024*1024))
	} else if totalBytesDiscovered > 1000000 {
		fmt.Printf(" %d MB", totalBytesDiscovered/(1024*1024))
	} else {
		fmt.Printf(" %d KB", totalBytesDiscovered/(1024))
	}
	fmt.Printf("\n\n")
	fmt.Println(NightFolders)
}

func createNightFolderForCamera(folder string) {

	fmt.Println("\n[CREATE DIRECTORY] :  " + folder)
	fmt.Println("Making night folder for camera: " + folder)
	if os.MkdirAll(folder, 0777) != nil {
		fmt.Println("Error making folder: " + folder)
	}
}
func walkFuncMoveNightFiles(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		fmt.Println(info.Name() + " is a directory.")
		var currentDirectory = info.Name()
		if currentDirectory == camFolderName {
			fmt.Println("[SKIPPING BASE FOLDER ] " + currentDirectory)
			createNightFolderForCamera(nightBaseFolder)
		} else {
			if IsFolderADate(currentDirectory) { //8 chars 20171212

				if datesToIgnore[currentDirectory] {
					fmt.Println("[IGNORE DATE ]  " + currentDirectory)
					return filepath.SkipDir
				} else {
					destFolderDate = filepath.Join(nightBaseFolder, currentDirectory)
					fmt.Println("\n[CREATE DIRECTORY] :  " + destFolderTemp)
					fmt.Println("Making staging folder for date: " + destFolderDate)
					if os.MkdirAll(destFolderDate, 0777) != nil {
						fmt.Println("Error making folder: " + destFolderDate)
					}
				}
			} else {
				// Folder is NOT a date
				fmt.Println(currentDirectory + " is NOT a date  directory of format YYYYMMDD")
				if IsFolderNightTime(currentDirectory) {
					NightFolders += currentDirectory + ", "
					fmt.Println(currentDirectory + " is a nighttime folder")
					destFolderTemp = filepath.Join(destFolderDate, currentDirectory)
					fmt.Println("\n[CREATE DIRECTORY] :  " + destFolderTemp)

					if os.MkdirAll(destFolderTemp, 0777) != nil {
						fmt.Println("Error making folder: " + destFolderTemp)
					} else {
						MoveAllFilesInFolder(path, destFolderTemp)
					}
				} else {
					fmt.Println("Skipping Day Folder and ALL FILES: " + currentDirectory)
					return filepath.SkipDir
				}
			}
		}
	} else {
		fmt.Println("[SKIP FILE] :  " + info.Name())
		// Files do not need to be analyzed individually
		//       the entire contents of that folder are either copied or skipped
		return filepath.SkipDir
	}
	return nil
}

func IsFolderNightTime(fileName string) bool {
	var hourStr = fileName[0:2]
	//hourVal, _ := strconv.ParseInt(hourStr, 0, 64)

	hourVal, _ := strconv.Atoi(hourStr)
	if hourVal >= 23 || hourVal < 5 {
		fmt.Printf("\n IDENTIFIED NIGHT FOLDER %d ", hourVal)
		return true
	} else {
		return false
	}
}

func moveNightFilesHTTPHandler(w http.ResponseWriter, r *http.Request) {

	keys := r.URL.Query()["cameraID"]

	if len(keys) < 1 {
		printLog(w, "Invalid Input - missing param", true)
	} else {
		baseSearchFolder = filepath.Join(baseSearchFolder, keys[0])
		nightBaseFolder = filepath.Join(nightBaseFolder, camFolderName)
		printLog(w, "Requesting to move nightfiles from folder: "+html.EscapeString(string(baseSearchFolder))+" to "+html.EscapeString(string(nightBaseFolder)), true)
	}

	err := filepath.Walk(baseSearchFolder, walkFunc)
	if err != nil {
		fmt.Println("Error :", err)
	}

	defer httpResponseSummaryForMoveNightFiles(w)

}

func httpResponseSummaryForMoveNightFiles(w http.ResponseWriter) {
	if totalBytesMoved > 0 {
		printLog(w, fmt.Sprintf("\nTotal Bytes Moved: %d", totalBytesMoved), true)
		printLog(w, fmt.Sprintf("\nTotal Files Moved: %d ", totalFilesMoved), true)
	} else {
		printLog(w, "\nNothing was done...", true)
	}
	totalBytesMoved = 0
	totalFilesMoved = 0

}
