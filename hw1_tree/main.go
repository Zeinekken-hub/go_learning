package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

const (
	subDirectoriesCapacity = 32
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	err := recTree(path, out, printFiles, 0, make([]int, 0, subDirectoriesCapacity))

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func recTree(filePath string, out io.Writer, printFiles bool, level int, notPrintLevels []int) error {
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	//update slice with new 0 level folder
	if level == 0 {
		notPrintLevels = make([]int, 0, subDirectoriesCapacity)
	}

	//check if printFiles value is false, than we find only directories
	if !printFiles {
		files = onlyDirectories(files)
	}

	for indx, file := range files {
		var isLast bool
		if indx == len(files)-1 {
			isLast = true
			if file.IsDir() {
				//Add not printable level if folder is last
				notPrintLevels = append(notPrintLevels, level+1)
			}
		}

		err := printLevel(out, level, file.Name(), isLast, notPrintLevels...)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		//if file is directory, then we go to the next directory by new path
		//but if not, then we print file name and count of bytes, also then we
		//stay in the same folder and continue the loop
		if file.IsDir() {
			fmt.Fprint(out, "\n")
			newPath := path.Join(filePath, file.Name())
			recTree(newPath, out, printFiles, level+1, notPrintLevels)
		} else {
			bytesS, err := bytesS(path.Join(filePath, file.Name()))
			if err != nil {
				return fmt.Errorf(err.Error())
			}
			fmt.Fprint(out, " ", bytesS, "\n")
		}
	}

	return nil
}

func printLevel(out io.Writer, level int, fileName string, isLast bool, notPrint ...int) error {
	if level < 0 {
		return fmt.Errorf("level can not be less then 0")
	}

Loop:
	for i := 1; i <= level; i++ {
		//check if level is in not printable levels
		for j := 0; j < len(notPrint); j++ {
			if i == notPrint[j] {
				fmt.Fprint(out, "\t")
				continue Loop
			}
		}
		fmt.Fprint(out, "│", "\t")
	}

	if isLast {
		fmt.Fprint(out, "└───", fileName)
	} else {
		fmt.Fprint(out, "├───", fileName)
	}

	return nil
}

//get bytes count of file by path
//if count == 0 then return "empty"
//else return "(123b)", where "123" can be size of file
func bytesS(path string) (string, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	if len(file) == 0 {
		return "(empty)", nil
	}

	return fmt.Sprint("(", len(file), "b)"), nil
}

//return files where type only directory
//using FileInfo method IsDir() bool
func onlyDirectories(files []os.FileInfo) []os.FileInfo {
	onlyDirs := make([]os.FileInfo, 0, len(files))

	for _, file := range files {
		if file.IsDir() {
			onlyDirs = append(onlyDirs, file)
		}
	}

	return onlyDirs
}
