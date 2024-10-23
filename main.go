package main

import (
	"fmt"
	"github.com/anisurrahman75/sh-playground/go-sh"
	"log"
)

func init() {
	initContainer()
	initResticRepos()
}

var resticRepos = map[string]map[string][]string{
	"s3": {
		"repo": []string{"s3:s3.us-east-1.amazonaws.com/anisur/multi-writter"},
		"envs": []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
	},
	"gcs": {
		"repo": []string{"gs:anisur:/multi-writter"},
		"envs": []string{"GOOGLE_PROJECT_ID", "GOOGLE_APPLICATION_CREDENTIALS"},
	},
}

func main() {
	session := sh.NewSession()
	session.ShowCMD = true
	session.Command("mysqldump", "-u", "root", "-h", "127.0.0.1",
		"-P", "3306", "-pmy-secret-pw", "--all-databases",
	)

	// Multiple backup Commands
	for _, val := range resticRepos {
		repo, envList := val["repo"][0], val["envs"]
		envs := getEnvs(envList)
		envs["RESTIC_PASSWORD"] = "my-secret-pw"
		session.BackupFromStdinCommand("restic", "backup", "-r", repo, "--stdin", envs)
	}

	if err := session.Run(); err != nil {
		fmt.Println(err)
	}

	/*
			---------------------------Sample Output-----------------------------------
		[golang-sh]$ mysqldump -u root -h 127.0.0.1 -P 3306 -pmy-secret-pw --all-databases | restic backup -r s3:s3.us-east-1.amazonaws.com/anisur/multi-writter --stdin , restic backup -r gs:anisur:/multi-writter --stdin

		Files:           1 new,     0 changed,     0 unmodified
		Dirs:            0 new,     0 changed,     0 unmodified
		Added to the repository: 592.321 MiB (30.527 MiB stored)

		processed 1 files, 592.295 MiB in 0:12
		snapshot ba7bdd20 saved

		Files:           1 new,     0 changed,     0 unmodified
		Dirs:            0 new,     0 changed,     0 unmodified
		Added to the repository: 592.320 MiB (30.112 MiB stored)

		processed 1 files, 592.295 MiB in 0:15
		snapshot 5aa8e802 saved

	*/
}

func databasseBackup() {
	session := sh.NewSession()
	session.ShowCMD = true
	session.Command("mysqldump", "-u", "root", "-h", "127.0.0.1",
		"-P", "3306", "-pmy-secret-pw", "--all-databases",
	)

	// Multiple backup Commands
	for _, val := range resticRepos {
		repo, envList := val["repo"][0], val["envs"]
		envs := getEnvs(envList)
		envs["RESTIC_PASSWORD"] = "my-secret-pw"
		session.BackupFromStdinCommand("restic", "backup", "-r", repo, "--stdin", envs)
	}

	if err := session.Run(); err != nil {
		fmt.Println(err)
	}

	/*
		---------------------------Sample Output-----------------------------------
		[golang-sh]$ cat dummy.data | restic backup -r s3:s3.us-east-1.amazonaws.com/anisur/multi-writter --stdin , restic backup -r gs:anisur:/multi-writter --stdin

		  Files:           1 new,     0 changed,     0 unmodified
		  Dirs:            0 new,     0 changed,     0 unmodified
		  Added to the repository: 1.000 GiB (1.000 GiB stored)

		  processed 1 files, 1.000 GiB in 3:23
		  snapshot e3c51182 saved

		  Files:           1 new,     0 changed,     0 unmodified
		  Dirs:            0 new,     0 changed,     0 unmodified
		  Added to the repository: 1.000 GiB (1.000 GiB stored)

		  processed 1 files, 1.000 GiB in 3:34
	*/
}

func debug() {
	session := sh.NewSession()
	session.ShowCMD = true
	session.Command("echo", "anisur rahman")
	session.Command("xargs")

	session.BackupFromStdinCommand("xargs")
	session.BackupFromStdinCommand("xargs")
	session.BackupFromStdinCommand("xargs")
	session.BackupFromStdinCommand("xargs")
	session.BackupFromStdinCommand("xargs")

	out, err := session.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-----out---")
	fmt.Println(string(out))

}

func largeFileTest() {
	fmt.Println("Testing 50GB file")
	sess := sh.NewSession()
	sess.Command("cat", "dummy.data")

	//sess.Command("tee", "copy1.data", ">", "/dev/null")
	// pipe into multiple commands input as stdin

	sess.BackupFromStdinCommand("tee", "copy1.data", ">", "/dev/null")
	sess.BackupFromStdinCommand("tee", "copy2.data", ">", "/dev/null")

	if err := sess.Run(); err != nil {
		fmt.Println(err)
	}
}
