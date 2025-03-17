package main

import (
	"log"
	"os"
	"time"

	"go-remote-exec/pkg/execute"
)

var (
	remoteHost     string
	remoteUser     string
	remotePassword string
)

func main() {
	var found bool
	remoteHost, found = os.LookupEnv("REMOTE_HOST")
	if !found {
		log.Fatal("REMOTE_HOST environment variable not set")
	}
	remoteUser, found = os.LookupEnv("REMOTE_USER")
	if !found {
		log.Fatal("REMOTE_USER environment variable not set")
	}
	remotePassword, found = os.LookupEnv("REMOTE_PASSWORD")
	if !found {
		log.Fatal("REMOTE_PASSWORD environment variable not set")
	}

	local := getLocal()
	defer local.Disconnect()
	remote := getRemote()
	defer remote.Disconnect()

	local.ExecuteBlocking("cat", "/etc/hostname")
	remote.ExecuteBlocking("cat", "/etc/hostname")
}

func getRemote() execute.Target {
	remote := execute.Remote{
		Environment: map[string]string{"FOO": "BAR"},
		Timeout:     time.Duration(10) * time.Second,
	}
	err := remote.Connect(remoteHost, remoteUser, remotePassword)
	if err != nil {
		log.Fatal(err)
	}

	return remote
}

func getLocal() execute.Target {
	local := execute.Local{
		Environment: map[string]string{"FOO": "BAR"},
	}
	return local

}
