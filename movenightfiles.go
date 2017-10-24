package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"path/filepath"
)
var camFolderName = "ImportCam7"
var basefolder = "C:\\FTPUploads" + "\\" + camFolderName
var nightStagingFolder = "c:\\temp\\nightFiles" + "\\" + camFolderName
var destFolderDate = ""
var totalFilesMoved = 0
var totalBytesMoved int64 = 0
const DEBUG bool = true

func main() {

	err := filepath.Walk(basefolder, walkFunc)
	if err != nil{
		fmt.Println("Error :", err)
	}

	fmt.Printf("\nTotal Bytes Moved: %d", totalBytesMoved)
	
	fmt.Printf("\nTotal Files Moved: %d ", totalFilesMoved)
}
func walkFunc(path string, info os.FileInfo, err error) error{
	if err!= nil{
		return err;
	}
	if info.IsDir() {
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
			}else if info.Name() != camFolderName{
				fmt.Println("Skipping Day Folder and ALL FILES: " + info.Name())
				return filepath.SkipDir
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