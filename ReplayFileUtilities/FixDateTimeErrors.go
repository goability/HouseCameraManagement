/*  REPLAY Tool
-   Fix Date and Time our of sync errors for files
- (c) Matt Chandler, 2018

-  POSSIBLE ERRORS TO FIX:
-    Camera's DateTime is not the same as the FTP server
-    FTP Server dateTime is off because it was not set correctly with NTS and is not connected to Internet
-  LOGIC:
		>  Walk folders starting with root
		>   Identify any files that are out of sync with current datetime
		>  Rename file with correct [prefix]YYYYMMDDHHMMSSms
		>  IF files were renamed in a given folder, rename that folder with correct HHMM IsFileTimeStampCorrect
		>  IF there are files outside the range of the given folder, push them into a collection to be  moved ONLY when folder exists:
		>  i.e.  files found between 10:30 and 11:04.  Get a list of files past 10:59 and move them into the 1100 folder on next iteration

*/
package main

//A change
import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

// Walk the folder
//  Get the timestamp of the file
//  If timestamp is off by more than 60 seconds, rename it

const DEBUG bool = false
const RUNASWEBSERVER = false

var serverPort = "9090"
var camFolderName = "IMPORT"
var baseSearchFolder = filepath.Join("/", "home", "matt", "images")

var datesToIgnore = map[string]bool{"20171224": true, "20171231": true, "20170906": false}

var totalFilesDiscovered = 0
var currentDirectoryFileCount = 0
var renamedFileByDirectory = make(map[string]int)
var existingFilesLeftInPlace = make(map[string]int)
var createdDirectories = make(map[string]bool) // Used to track AND to not retraverse
var existingDestinationDirectoryCount = 0
var existingFilesAlreadyinDestination = 0 // Track files that already existing before renaming
var lastConstructedDestFolder = ""        // used to track when destfolder name changes
var totalFileCountBeforeScan = 0
var totalFileCountAfterScan = 0

var totalFilesRenamed = 0
var totalFilesRenamedFailed = 0

var CurrentDirectory = ""
var CurrentDateDirectory = ""
var CurrentFile = ""

var fileExtension = ".jpg"

//File index markers

var lengthOfDateTimeStringInFileName = 16 //this is number of indexes from end of file that YYYY starts
//i.e. YYYYMMDDHHMMSSms.jpg = 20

var indexOfDateStart = 0

func setWindowsPaths() {
	baseSearchFolder = filepath.Join("c:\\", "FTPUploads", camFolderName)
}
func main() {
	CurrentDirectory = baseSearchFolder
	if runtime.GOOS == "windows" {
		setWindowsPaths()
	} else {
		baseSearchFolder = filepath.Join(baseSearchFolder, camFolderName)
	}

	//Scan all files to get a pre-run file Count
	totalFileCountBeforeScan = TotalFileCountofBase(baseSearchFolder)

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
		fmt.Printf("\nTotal Files Discovered: %d ", totalFilesDiscovered)

	} else {
		fmt.Printf("\nTotal files scanned: %d", totalFilesDiscovered)
		fmt.Printf("\nTotal files renamed: %d ", totalFilesRenamed)
		fmt.Printf("\nDirectories that already existed in destination: %d - (meaning not created again)", existingDestinationDirectoryCount)
		fmt.Printf("\nFiles that already existed in destination: %d - (meaning not copied so left in place)", existingFilesAlreadyinDestination)

		fmt.Printf("\n\n-- Count of files renamed by Directory --- ")
		for directory, fileCount := range renamedFileByDirectory {
			fmt.Println(directory, " : ", fileCount)
		}
		fmt.Printf("\n\n-- Existing files left in place --- ")
		for fileName, notused := range existingFilesLeftInPlace {
			fmt.Println(fileName, " : ", notused)
		}
		totalFilesDiscovered = 0
		totalFileCountAfterScan = TotalFileCountofBase(baseSearchFolder)
		fmt.Printf("\n Total file count sanity check:  Before===After [%d]===[%d]", totalFileCountBeforeScan, totalFileCountAfterScan)

	}
	fmt.Printf("\n\n")

}

// Get count of all files below a base directory
func TotalFileCountofBase(baseFolder string) int {
	//Walk the folder, increment with each file count
	//Walk the baseSearchFolder
	err := filepath.Walk(baseFolder, walkCountFunc)
	if err != nil {
		fmt.Print("ERROR Walking folder:  ")
		fmt.Print(baseFolder)
		fmt.Println(" ERR:", err)
	}
	return totalFilesDiscovered
}
func walkCountFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		totalFilesDiscovered++
	}
	return nil
}
func showStart() {
	fmt.Println("\n\n---------------------------")
	fmt.Println("STARTING SCAN: " + baseSearchFolder)
	fmt.Println("")
	totalFilesDiscovered = 0
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
		//  Directories are :  /basefolder/YYYYMMDD/HHMM/fileName
		//  Current Directory is tracked to concat to filename below
		// LOGIC:
		//		If a date directory, reset CurrentDateDirectory
		//		Else if a time directory, simply concat that with CurrentDateDirectory
		//

		//fmt.Println(CurrentDirectory + " is a directory.")
		if IsFolderADate(info.Name()) {
			CurrentDateDirectory = filepath.Join(baseSearchFolder, info.Name())

			//fmt.Println("Resetting Date folder to " + CurrentDirectory)
		} else if IsFolderATime(info.Name()) {
			//A time folder, concat current HHMM to CurrentDate
			CurrentDirectory = filepath.Join(CurrentDateDirectory, info.Name())
			currentDirectoryFileCount = 0 //Reset the current folder file count (This var is for protection to ensure numbers match before and after )
			if createdDirectories[CurrentDirectory] {
				fmt.Println("[IGNORE NEWLY CREATED DIRECTORY ]  " + CurrentDirectory)
				return filepath.SkipDir
			}
		}
		//	return filepath.SkipDir
	} else {
		CurrentFile = filepath.Join(CurrentDirectory, info.Name())
		fi, err := os.Stat(CurrentFile)
		if err != nil {
			return err
		} else {
			var correctDatetimeStamp = IsFileTimeStampCorrect(info.Name(), fi.ModTime())
			if correctDatetimeStamp != "" {
				fmt.Println("[DATETIME MISMATCH]: ", CurrentFile, "-", fi.ModTime())

				fixFileName(CurrentFile, correctDatetimeStamp)

			}
			totalFilesDiscovered++ // this is the master count just to give an idea of scope
		}
	}
	return nil
}
func fixFileName(filePath string, correctDatetimeStamp string) {

	//filePath:  /home/matt/images/IMPORT/20170906/0000/camKitchen2017090500000201.jpg
	//fmt.Println("FILEname length: ", len(filePath))
	var startIndexOfDateTimeStampInFileName = len(filePath) - (lengthOfDateTimeStringInFileName + len(fileExtension))
	//  This will get:  2017090500000201.jpg

	//baseSearchFolder:  /home/matt/images/IMPORT
	var startIndexOfFolderDate = len(baseSearchFolder)
	var startIndexOfFilenamePrefix = startIndexOfFolderDate + 15
	// where 15 is always equal to /YYYYMMDD/HHMM

	// So fileNamePrefix = camKitchen
	var fileNamePrefix = filePath[startIndexOfFilenamePrefix:startIndexOfDateTimeStampInFileName]

	//var fullFileprefix = filePath[0:startIndexOfDateTimeStampInFileName]
	var correctedFileName = correctDatetimeStamp + fileExtension

	//i.e. 20170906/0000/camKitchen2017090500000201.jpg =

	//Given YYYYMMDDHHMMSSms, return : YYYY/MMDD/HH00  or HH30

	var fixedDirectoryName = fixDirectoryName(correctDatetimeStamp)

	var correctedDirectoryPath = filepath.Join(baseSearchFolder, fixedDirectoryName, fileNamePrefix+correctedFileName)

	fmt.Println("[FIXING FILE] : ", filePath, " TO: ", correctedDirectoryPath)

	if !DEBUG {
		var destFolder = filepath.Join(baseSearchFolder, fixedDirectoryName)

		if lastConstructedDestFolder != destFolder { // Only check this when destFolder changes

			lastConstructedDestFolder = destFolder
			// Only create a new destination folder if one does not already exist!!
			if _, err := os.Stat(destFolder); os.IsNotExist(err) {
				//Folder does not exist, so create it
				if os.MkdirAll(destFolder, 0777) != nil {
					createdDirectories[destFolder] = true
					fmt.Println("Error making folder: " + destFolder)
				}
			} else {
				existingDestinationDirectoryCount++
			}
		}

		// Make sure file in destination does not already exist !!
		if _, err := os.Stat(correctedDirectoryPath); os.IsNotExist(err) {

			errMv := os.Rename(filePath, correctedDirectoryPath)
			if errMv != nil {
				fmt.Println("FATAL ERROR ON RENAME", errMv)
				log.Fatal(errMv)
				totalFilesRenamedFailed++
			} else {
				totalFilesRenamed++
				currentDirectoryFileCount++
				renamedFileByDirectory[destFolder] = currentDirectoryFileCount
			}
		} else {
			fmt.Println("File ", correctedDirectoryPath, " already existed!  Not copying..")
			existingFilesAlreadyinDestination++
			existingFilesLeftInPlace[correctedDirectoryPath] = 0 //matt used map, but list would be better
		}
	} //end if DEBUG
}
func fixDirectoryName(fileDateTimestamp string) string {
	// Given a file's name that containts:  YYYYMMDDHHMMSSms,
	//    return a folderName that is of format YYYYMMDD/HHMM/ where MM{00 or 30}
	// RETURN YYYY/HHMM

	//return "UNKNOWN"
	//fmt.Println("FIXING DIRECTORY from timestamp: ", fileDateTimestamp)
	var year = fileDateTimestamp[:4]
	var month = fileDateTimestamp[4:6]
	var day = fileDateTimestamp[6:8]
	var hour = fileDateTimestamp[8:10]
	var min = fileDateTimestamp[10:12]
	var minDirectory = "30"

	var minVal, _ = strconv.Atoi(min)
	if minVal >= 0 && minVal < 30 {
		minDirectory = "00"
	}

	return filepath.Join(year+month+day, hour+minDirectory)

}
func IsFileTimeStampCorrect(fileName string, actualTime time.Time) string {
	// INPUTS
	//     fileName = somePrefixYYYYMMDDHHMMSSxx
	//     actualTime is a time structure

	//OUTPUT :  Corrected YYMMDDHHMMSSxx
	// Collapse  the YYYYMMDD from the actual time and first compare that
	// If that is incorrect at all, return NULL or ""

	// If date is correct, look deeper and check the HHMMSS
	// If off by more than 30 seconds, return false

	// Extract an YYYYMMDDHHMMSS from the fileName

	var fileNameLen = len(fileName)
	var startIndex = 0
	var endIndex = 0
	var year = ""

	var month = ""
	var day = ""

	var fileNeedsFixing = false
	//	var hour = ""
	//	var minute = ""

	var corectedDateTime = ""

	//fmt.Println("[ANALYZE ]", fileName, " has length of: ", fileNameLen)

	startIndex = fileNameLen - (lengthOfDateTimeStringInFileName + len(fileExtension))
	endIndex = startIndex + 4

	//fmt.Println("FILE ", fileName, " will start at index: ", startIndex)

	if startIndex > 0 {
		year = fileName[startIndex:endIndex]
		month = fileName[endIndex : endIndex+2]
		day = fileName[endIndex+2 : endIndex+4]
		//hour = fileName[endIndex+4 : endIndex+6]
		//minute = fileName[endIndex+6 : endIndex+8]

		if !DEBUG {
			//	fmt.Print("[FILENAME DATE] :", year, month, day, hour, minute)
			//fmt.Println(" = [MONTH]: ", int(actualTime.Month()))
		}

		var cmp, _ = strconv.Atoi(year)
		if cmp != actualTime.Year() {
			fileNeedsFixing = true
			//fmt.Println("YEAR: ", cmp, " != ", actualTime.Year())
		} else {
			cmp, _ = strconv.Atoi(month)
			var realMonth = int(actualTime.Month())
			if cmp != realMonth {
				fileNeedsFixing = true
			} else {
				cmp, _ = strconv.Atoi(day)
				if cmp != actualTime.Day() {
					fmt.Println("FIXING DAY ", cmp, " != ", actualTime.Day())
					fileNeedsFixing = true
				}
			}
		}
		if fileNeedsFixing {
			corectedDateTime = strconv.Itoa(actualTime.Year()) + fmt.Sprintf("%02d", int(actualTime.Month())) +
				fmt.Sprintf("%02d", actualTime.Day()) + fmt.Sprintf("%02d", actualTime.Hour()) +
				fmt.Sprintf("%02d", actualTime.Minute()) + fmt.Sprintf("%02d", actualTime.Second()) + "01"
		}
	}
	if DEBUG {
		fmt.Println("RETURNING: ", corectedDateTime)
	}
	return corectedDateTime
}
func IsFolderADate(fileName string) bool {
	if len(fileName) == 8 {
		return true
	} else {
		return false
	}
}
func IsFolderATime(fileName string) bool {
	if len(fileName) == 4 && (fileName[2] == '0' || fileName[2] == '3') {
		return true
	} else {
		return false
	}
}

func printLog(w http.ResponseWriter, txt string, sendToPage bool) {

	fmt.Println(txt)
	if sendToPage {
		fmt.Fprint(w, html.EscapeString(txt))
	}
}
func printSummary(w http.ResponseWriter) {
	if totalFilesRenamed > 0 {
		printLog(w, fmt.Sprintf("\nTotal Files Rename: %d ", totalFilesRenamed), true)
		printLog(w, fmt.Sprintf("\nTotal Files Rename Failures: %d ", totalFilesRenamedFailed), true)
	} else {
		printLog(w, "\nNothing was done...", true)
	}
	totalFilesRenamed = 0

}
