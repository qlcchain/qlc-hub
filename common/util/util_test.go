package util

import (
	"os"
	"testing"
)

type StringTag struct {
	BoolStr    bool    `json:",string"`
	IntStr     int64   `json:",string"`
	UintptrStr uintptr `json:",string"`
	StrStr     string  `json:",string"`
}

func TestCreateDirIfNotExist(t *testing.T) {
	err := CreateDirIfNotExist("./test")
	if err != nil {
		t.Fatal(err)
	}
	_ = os.RemoveAll("./test")
}

func TestToIndentString(t *testing.T) {
	var s StringTag
	s.BoolStr = true
	s.IntStr = 42
	s.UintptrStr = 44
	s.StrStr = "xzbit"
	st := ToIndentString(s)
	t.Log(st)
}
