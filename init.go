package main

import (
	"bytes"
	"fmt"
	"github.com/anisurrahman75/sh-playground/go-sh"
	"log"
	"os"
	"os/exec"
	"strings"
)

func containerExists(name string) bool {
	cmd := exec.Command("docker", "ps", "--filter", "name="+name, "--format", "{{.Names}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to check if container exists: %v", err)
	}

	fmt.Println(strings.TrimSpace(out.String()))

	return strings.TrimSpace(out.String()) == name
}

func initContainer() {
	containerName := "mysql-container"

	if containerExists(containerName) {
		log.Printf("Container %s already exists. Skipping run command.", containerName)
		return
	}

	cmd := exec.Command("docker", "run", "--name", containerName,
		"-e", "MYSQL_ROOT_PASSWORD=my-secret-pw", "-p", "3306:3306", "-d", "mysql:latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run container: %v", err)
	} else {
		log.Printf("Container %s started successfully.", containerName)
	}
}

func getEnvs(envList []string) map[string]string {
	envs := make(map[string]string)
	for _, env := range envList {
		value, exist := os.LookupEnv(env)
		if exist {
			envs[env] = value
		} else {
			log.Printf("Environment variable %s not set.", env)
		}
	}
	return envs
}

func initResticRepos() {
	for _, val := range resticRepos {
		repo, envList := val["repo"][0], val["envs"]
		envs := getEnvs(envList)
		sess := sh.NewSession()
		sess.Env = envs
		sess.SetEnv("RESTIC_PASSWORD", "my-secret-pw")
		sess.Command("restic", "snapshots", "-r", repo)

		if err := sess.Run(); err != nil {
			initSession := sh.NewSession()
			initSession.Env = envs
			initSession.SetEnv("RESTIC_PASSWORD", "my-secret-pw")
			initSession.Command("restic", "init", "-r", repo)
			if err := initSession.Run(); err != nil {
				log.Printf("Failed to init restic repo %s: %v", repo, err)
			}
		}
	}
}
