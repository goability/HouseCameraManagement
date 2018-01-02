package main

//A change
import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var camFolderName = "IMPORT"
var baseSearchFolder = "/home/matt/images/" + camFolderName
var nightBaseFolder = "/home/matt/images/night/" + camFolderName
var destFolderDate = ""
var destFolderTemp = ""

var datesToIgnore = map[string]bool{"20171224": true, "20171231": true, "20170906": false}

var totalFilesMoved = 0
var totalBytesMoved int64 = 0

var totalFilesDiscovered = 0
var totalBytesDiscovered int64 = 0

var totalFileMovesFailed = 0

var NightFolders = ""

const DEBUG bool = true

func main() {
	showStart()

	//Walk the baseSearchFolder
	err := filepath.Walk(baseSearchFolder, walkFunc)
	if err != nil {
		fmt.Print("ERROR Walking folder:  ")
		fmt.Print(baseSearchFolder)
		fmt.Println(" ERR:", err)
	} else {
		showSummary()
	}

}
func showSummary() {
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
func showStart() {
	fmt.Println("\n\n---------------------------")
	fmt.Println("STARTING SCAN: " + baseSearchFolder)
	fmt.Println("")
}
func createNightFolderForCamera(folder string) {

	fmt.Println("\n[CREATE DIRECTORY] :  " + folder)
	fmt.Println("Making night folder for camera: " + folder)
	if os.MkdirAll(folder, 0777) != nil {
		fmt.Println("Error making folder: " + folder)
	}
}
func walkFunc(path string, info os.FileInfo, err error) error {
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
					destFolderDate = nightBaseFolder + "/" + currentDirectory
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
					destFolderTemp = destFolderDate + "/" + currentDirectory
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
func MoveAllFilesInFolder(folderName string, destFolderName string) {
	files, err := ioutil.ReadDir(folderName)
	if err != nil {
		log.Fatal(err)
	}
	if files != nil {
		for _, file := range files {
			totalBytesDiscovered += file.Size()
			totalFilesDiscovered++

			if DEBUG == false {
				fileDest := destFolderName + "/" + file.Name()
				fileSrc := folderName + "/" + file.Name()
				errMv := os.Rename(fileSrc, fileDest)

				if errMv != nil {
					log.Fatal(errMv)
					totalFileMovesFailed++
				} else {
					totalBytesMoved += file.Size()
					totalFilesMoved++
					//fmt.Println("[MOVE FI LE] : '" + fileSrc + "' to " + fileDest)
				}
			}
		}
	}
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
func IsFolderADate(fileName string) bool {
	if len(fileName) == 8 {
		return true
	} else {
		return false
	}
}
