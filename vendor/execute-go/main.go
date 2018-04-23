package main

import (
	"fmt"
	"os/exec"
	"log"
	"path/filepath"
	"path"
	"io/ioutil"
	"os"
	"time"
	"crypto/md5"
	"strconv"
	"runtime"
)
type TmpFile struct {
	GoFile string
	Binary string
	Data   []byte
}

func main () {
	test := exec.Command("go", "build","-o","test" ,"-v")
	output, err := test.CombinedOutput() 
	if err != nil {
		log.Fatal(err)	
	}
	fmt.Println(string(output))
}

func gobuild(data []byte) error {
	tmp, err := generateTmpFile(data)
	if err != nil {
		return err
	}
	if err = tmp.WriteSource() ; err != nil {
		return err
	}
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	gopath:=os.Getenv("GOPATH")
	// todo print log to a io.Writer
	log.Printf("     GOOS : %s \n     GOARCH : %s \n     GOPATH : %s \n", goos, goarch, gopath)
	log.Println(" ================= Start compile and build file ===============")
	e := exec.Command("go","build","-o",tmp.Binary,"-v")

	if err != nil {
		return err
	}
	return nil
}

func generateTmpFile(data []byte) (*TmpFile, error) {
	now := time.Now().Unix()
	digest := md5.Sum(data)
	file := strconv.FormatInt(now, 10)+"-"+string(digest[:])
	abs , err := filepath.Abs("")
	if err != nil {
		return nil, err
	}
	file = path.Join(abs, file)
	return &TmpFile{ GoFile: file + ".go", Binary: file, Data: data}, nil
}
func (t *TmpFile) WriteSource() error {
	if err := ioutil.WriteFile(t.GoFile, t.Data, 0644); err != nil { return err }
	return nil
}
func (t *TmpFile) RemoveAll() error {
	if err := os.Remove(t.GoFile); err !=nil { return err }
	if err := os.Remove(t.Binary); err !=nil { return err }
	return nil
}