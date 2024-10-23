package main

import (
	"fmt"
	"github.com/anisurrahman75/sh-playground/go-sh"
	"testing"
)

func Test20GB(t *testing.T) {
	fmt.Println("Testing 50GB file")
	sess := sh.NewSession()
	sess.Command("cat", "dummy.data")

	// pipe into multiple commands input as stdin

	sess.BackupFromStdinCommand("xargs", ">", "copy_dummy1.data")
	sess.BackupFromStdinCommand("xargs", ">", "copy_dummy2.data")
}
