package main

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const (
	tmpDir = "____tmp"
)

type BuildSupport struct {
	Path    string
	GoFile  string
	Binary  string
	Data    []byte
	Logger  *logging.Logger
	Backend *WebsocketBackend
}

var logger *logging.Logger

func init() {
	logging.SetFormatter(logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`))
	logger = logging.MustGetLogger("execute-go")
}

func (t *BuildSupport) GoBuild() error {
	if err := t.WriteSource(); err != nil {
		return err
	}
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	gopath := os.Getenv("GOPATH")
	t.Logger.Debugf("\n     GOOS : %s \n     GOARCH : %s \n     GOPATH : %s \n", goos, goarch, gopath)
	t.Logger.Debugf("  Compile and Build file : %s ", t.GoFile)
	e := exec.Command("go", "build", "-o", t.Binary)
	// must set the right working dir for 'go build'
	e.Dir = t.Path
	result, err := e.CombinedOutput()
	if err != nil {
		t.Logger.Errorf("Error running go build : %s ", err)
		return err
	}
	logger.Debug("\n%s\n", string(result))
	return nil
}
func (t *BuildSupport) GoRun() error {
	t.Logger.Debugf("Exec the built file : %s", t.Binary)
	e := exec.Command(t.Binary)
	res, err := e.CombinedOutput()
	if err != nil {
		t.Logger.Errorf("Error running the binary : %s", err)
		return err
	}
	t.Logger.Infof("\n%s\n", string(res))
	// send a close msg to ws backend
	t.Logger.Warning("close")
	return nil
}
func NewBuildSupport(data []byte, backend *WebsocketBackend) (*BuildSupport, error) {
	now := time.Now().Unix()
	digest := md5.Sum(data)
	file := strconv.FormatInt(now, 10) + "-" + hex.EncodeToString(digest[:])
	abs, err := filepath.Abs("")
	if err != nil {
		logger.Errorf("Error getting the abs path : %s", err)
		return nil, err
	}
	absTmpDir := path.Join(abs, tmpDir)
	if _, err := os.Stat(absTmpDir); os.IsNotExist(err) || err == nil {
		if err == nil {
			if err := os.RemoveAll(absTmpDir); err != nil {
				logger.Errorf("Error removing legacy tmp dir : %s", err)
				return nil, err
			}
		}
		if err = os.Mkdir(absTmpDir, 0777); err != nil {
			logger.Errorf("Error creating tmp dir '%s' : %s", absTmpDir, err)
			return nil, err
		}
	} else {
		logger.Errorf("Path error : %s ", err)
		return nil, err
	}

	file = path.Join(absTmpDir, file)
	logger := logging.MustGetLogger("BuildSupport")

	b := logging.NewBackendFormatter(backend, logging.MustStringFormatter(`%{color}%{time:15:04:05.000} ▶ %{level:.4s} %{color:reset} %{message}`))
	logger.SetBackend(logging.MultiLogger(b))
	return &BuildSupport{Path: absTmpDir, GoFile: file + ".go", Binary: file, Data: data, Logger: logger, Backend: backend}, nil
}
func (t *BuildSupport) WriteSource() error {
	if err := ioutil.WriteFile(t.GoFile, t.Data, 0777); err != nil {
		logger.Errorf("Error creating file '%s' : %s", t.GoFile, err)
		return err
	}

	return nil
}
func (t *BuildSupport) RemoveAll() error {
	if err := os.RemoveAll(t.Path); err != nil {
		logger.Errorf("Error remove the tmp dir : %s ", err)
		return err
	}
	return nil
}

func (t *BuildSupport) Start() {
	t.Logger.Debugf(" ==================== Start running the source file ====================  ")
	defer t.RemoveAll()
	if err := t.GoBuild(); err != nil {
		t.Backend.AbnormalClose(err)
		return
	}
	if err := t.GoRun(); err != nil {
		t.Backend.AbnormalClose(err)
		return
	}
}
