package main

import (
	"github.com/anisurrahman75/sh-playground/go-sh"
	"log"
	"os/exec"
)

//export GOROOT=/usr/local/go
//export PATH=$PATH:$GOROOT/bin

func main() {
	session := sh.NewSession()
	session.ShowCMD = true
	cmd := sh.CMD{
		Cmd: exec.Command("echo", "/usr/local"),
		ChildCmds: []*sh.CMD{
			{
				Cmd: exec.Command("ls", "/lost+found/"),
			},
		},
	}

	cmds := []*sh.CMD{&cmd}
	session.Command("ls", "/usr/local/", cmds)

	session.Command("ls", "/usr/local/lib/")

	err := session.Run()
	if err != nil {
		log.Fatal(err)
	}
}
