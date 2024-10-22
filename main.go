package main

import (
	"fmt"
	"github.com/anisurrahman75/sh-playground/go-sh"
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
}
