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
	"github.com/op/go-logging"
	"encoding/hex"
)
const (
	tmpDir = "____tmp"
)
type TmpFile struct {
	Path string
	GoFile string
	Binary string
	Data   []byte
}
var logger *logging.Logger
func init() {
	logging.SetFormatter(logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`))
	logger = logging.MustGetLogger("execute-go")
}
func main () {
	test := exec.Command("go", "build","-o","test" ,"-v")
	output, err := test.CombinedOutput() 
	if err != nil {
		log.Fatal(err)	
	}
	fmt.Println(string(output))
}

func(t *TmpFile) GoBuild() error {
	if err := t.WriteSource() ; err != nil {
		return err
	}
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	gopath:=os.Getenv("GOPATH")
	// todo print log to a io.Writer
	logger.Debugf("\n     GOOS : %s \n     GOARCH : %s \n     GOPATH : %s \n", goos, goarch, gopath)
	logger.Debugf("  Compile and Build file : %s ", t.GoFile)
	e := exec.Command("go","build","-o",t.Binary)
	// must set the right working dir for 'go build'
	e.Dir = t.Path
	result, err := e.CombinedOutput()
	if err != nil {
		logger.Errorf("Error running go build : %s ", err)
		return err
	}
	logger.Debug("\n%s\n",string(result))
	return nil
}
func (t *TmpFile) GoRun() error {
	logger.Debugf("Exec the built file : %s", t.Binary)
	e := exec.Command(t.Binary)
	res , err := e.CombinedOutput()
	if err != nil {
		logger.Errorf("Error running the binary : %s", err)
		return err
	}
	logger.Infof("\n%s\n", string(res))
	return nil
}
func NewTmpFile(data []byte) (*TmpFile, error) {
	now := time.Now().Unix()
	digest := md5.Sum(data)
	file := strconv.FormatInt(now, 10)+"-"+hex.EncodeToString(digest[:])
	abs , err := filepath.Abs("")
	if err != nil {
		logger.Errorf("Error getting the abs path : %s", err)
		return nil, err
	}
	absTmpDir := path.Join(abs, tmpDir)
	if _ ,err := os.Stat(absTmpDir) ; os.IsNotExist(err) {
		if err = os.Mkdir(absTmpDir, 0777) ; err != nil {
			logger.Errorf("Error creating tmp dir '%s' : %s", absTmpDir, err)
			return nil, err
		}
	} else {
		logger.Errorf("Path error : %s ", err)
		return nil, err
	}

	file = path.Join(absTmpDir,file)
	return &TmpFile{ Path: absTmpDir, GoFile: file + ".go", Binary: file, Data: data}, nil
}
func (t *TmpFile) WriteSource() error {
	if err := ioutil.WriteFile(t.GoFile, t.Data, 0777); err != nil { return err }
	return nil
}
func (t *TmpFile) RemoveAll() error {
	if err := os.RemoveAll(t.Path); err != nil {
		logger.Errorf("Error remove the tmp dir : %s ", err)
		return err
	}
	return nil
}