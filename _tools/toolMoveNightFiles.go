package main

import (
	"html"
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"path/filepath"
	"net/http"
)
var test = "test"
var camFolderNameBaseName = "ImportCam"
var basefolder = "C:\\FTPUploads"
var nightStagingFolder = "c:\\temp\\nightFiles"
var nightStagingFolderBaseName = "c:\\temp\\nightFiles"
var destFolderDate = ""
var totalFilesMoved = 0
var totalBytesMoved int64 = 0
var serverPort = "9090"
var folderToWalk = basefolder

const DEBUG bool = true

func main() {

	http.HandleFunc("/", showMainToolingPage)
	http.HandleFunc("/_toolMoveFiles", movenightfiles)

	fmt.Println("Running Camera Management Webserver on port " + serverPort)
	log.Fatal(http.ListenAndServe(":" + serverPort, nil))

}
func showMainToolingPage(w http.ResponseWriter, r *http.Request){

	fmt.Fprint(w, "Welcome to the main page")
	//printLog(w, "\nVisitor on page...", false)
}
func movenightfiles(w http.ResponseWriter, r *http.Request){

	keys := r.URL.Query()["cameraID"]
	
	if len(keys)<1{
		printLog(w, "Invalid Input - missing param", true)
	}else{
		camFolderName := camFolderNameBaseName + keys[0]
		folderToWalk := basefolder + "\\" + camFolderName
		nightStagingFolder = nightStagingFolderBaseName + "\\" + camFolderName
		printLog(w, "Requesting to move nightfiles from folder: " + html.EscapeString(string(folderToWalk)) + " to " + html.EscapeString(string(nightStagingFolder)), true)
	}

	err := filepath.Walk(folderToWalk, walkFunc)
		if err != nil{
			fmt.Println("Error :", err)
		}

	defer printSummary(w)
		
}
func printLog(w http.ResponseWriter, txt string, sendToPage bool){
	
	fmt.Println(txt)
	if sendToPage{
		fmt.Fprint(w, html.EscapeString(txt))
	}
}
func printSummary(w http.ResponseWriter){
	if (totalBytesMoved>0){
		printLog(w, fmt.Sprintf("\nTotal Bytes Moved: %d", totalBytesMoved), true)
		printLog(w, fmt.Sprintf("\nTotal Files Moved: %d ", totalFilesMoved), true)
	}else{
		printLog(w, "\nNothing was done...", true)
	}
	totalBytesMoved = 0
	totalFilesMoved = 0

}
func walkFunc(path string, info os.FileInfo, err error) error{
	if err!= nil{
		return err;
	}
	if info.IsDir() {
		fmt.Println("folder: " + path)
		if IsFolderADate(info.Name()){
			destFolderDate = nightStagingFolder + "\\" + info.Name()
			fmt.Println("Making staging folder for date: " + destFolderDate)
			if os.Mkdir(destFolderDate, os.ModeDir) != nil {
				fmt.Println("Error making folder: " + destFolderDate)
			}
		}else{
			if IsFolderNightTime(info.Name()){
				destFolder := destFolderDate + "\\" + info.Name()
				fmt.Println("Making night staging folder: " + destFolder)

				if os.Mkdir(destFolder, os.ModeDir) != nil {
					fmt.Println("Error making folder: " + destFolder)
				}else{
					MoveAllFilesInFolder(path, destFolder)
				}
			}else if info.Name() != basefolder{
				fmt.Println("Skipping Day Folder and ALL FILES: " + info.Name())
				//return filepath.SkipDir
			}
		}
	}else{
		fmt.Println("Skipping File: " + info.Name())
		// return SkipDir for a file
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
				
			fileDest := destFolderName + "\\" + file.Name()
			fileSrc := folderName + "\\" + file.Name()
			errMv := os.Rename(fileSrc, fileDest)
			if errMv != nil{
				log.Fatal(errMv)
			}else{
				totalBytesMoved += file.Size()
				totalFilesMoved++
				//fmt.Println("Moving file '" + fileSrc + "' to " + fileDest)
			}
		}
	}
}
func IsFolderNightTime (fileName string) bool{
	var hourStr = fileName[0:2]
	hourVal, _ := strconv.ParseInt(hourStr, 0, 64)

	if hourVal >= 20{
		return true;
	} else{
		return false;
	}
}
func IsFolderADate (fileName string) bool{
	if len(fileName) == 8{
		return true;
	} else{
		return false;
	}
}
