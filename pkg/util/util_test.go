package util

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

func TestIsvalidNEOAddress(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ok",
			args: args{
				addr: "ARNpaFJhp6SHziRomrK4cenWw66C8VVFyv",
			},
			want: true,
		}, {
			name: "fail",
			args: args{
				addr: "ARNpaFJhp6SHziRomrK4cenWw66C8VVFyv1",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if r := IsvalidNEOAddress(tt.args.addr); r != tt.want {
				t.Errorf("IsvalidNEOAddress() got = %v, want %v", r, tt.want)
			}
		})
	}
}

func TestEthAddress(t *testing.T) {
	t.Log(common.IsHexAddress("0x2e1ac6242bb084029a9eb29dfb083757d27fced4"))
}
