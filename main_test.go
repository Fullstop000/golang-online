package main

import (
	"testing"
	"io/ioutil"
)

func Test_Main(t *testing.T) {
	data, err := ioutil.ReadFile("./test/example01.go")
	if err != nil {
		t.Fatalf("Error reading example01.go : %s", err)
	}
	tmp ,err  := NewTmpFile(data)
	defer tmp.RemoveAll()
	if err != nil {
		t.Fatalf("Error building the tmp file : %s", err)
	}
	if err = tmp.GoBuild() ; err != nil {
		t.Fatalf("Error compiling and building : %s ", err)
	}
	if err = tmp.GoRun() ; err != nil {
		t.Fatalf("Error running the binary : %s ", err)
	}
}
