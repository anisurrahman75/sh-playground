package main

import (
	"fmt"
	"github.com/anisurrahman75/sh-playground/go-sh"
	"log"
)

//export GOROOT=/usr/local/go
//export PATH=$PATH:$GOROOT/bin

func main() {
	session := sh.NewSession()
	session.ShowCMD = true
	session.Command("ls", "/usr/local/")
	session.Command("echo", "Anisur rahman")
	//session.Command("xargs")

	session.BackupFromStdinCommand("xargs")
	//session.BackupFromStdinCommand("xargs")

	out, err := session.Output()
	if err != nil {
		log.Fatal(err)
	}
	_ = out
	fmt.Println(string(out))
}
