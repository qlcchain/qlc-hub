package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/abiosoft/ishell"
)

func hDelete(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "delete",
		Help: "delete swap info",
		Func: func(c *ishell.Context) {
			hDeleteLockerInfo()
		},
	}
	parentCmd.AddCmd(c)
}

func hDeleteLockerInfo() {
	log.Println("====hub delete====")
	//rHash := hNeo2Eth()
	rHash := "64aca8c20625adbcfc2ef8eb167f5440f60d506be27b97577356d976293661d0"
	time.Sleep(10 * time.Second)

	paras2 := fmt.Sprintf(`{
		"rHash": "%s",
	}`, rHash)
	r, err := post(paras2, fmt.Sprintf("%s/debug/deleteLockerInfo", hubUrl))
	if err != nil {
		log.Fatal(err)
	}
	if !r.(bool) {
		log.Fatal("")
	}

}
