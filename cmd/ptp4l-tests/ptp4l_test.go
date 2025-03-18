package ptp4l_tests

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"go-remote-exec/pkg/remote"
)

func TestPtp4l(t *testing.T) {
	//start := time.Now()

	hostA, err := remote.CreateHost("raspi-a.lan", "jpmeijers", os.Getenv("REMOTE_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		t.Log("Closing host A")
		hostA.Close()
	})

	hostB, err := remote.CreateHost("raspi-b.lan", "jpmeijers", os.Getenv("REMOTE_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		t.Log("Closing host B")
		hostB.Close()
	})

	// Install linuxptp snap
	//err = remote.Setup(hostA)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//err = remote.Setup(hostB)
	//if err != nil {
	//	t.Fatal(err)
	//}

	// start server
	t.Log("=== Starting server ===")
	server, serverStdOut, serverStdErr, err := remote.StartServer(hostA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		t.Log("Closing server")
		remote.Stop(server)
	})

	err = remote.WaitFor(serverStdOut, "assuming the grand master role", 20*time.Second)
	t.Log(*serverStdOut)
	if err != nil {
		t.Log(*serverStdErr)
		t.Fatal(err)
	}
	t.Log("=== Server started ===")

	t.Log("=== Starting client ===")
	client, clientStdOut, clientStdErr, err := remote.StartClient(hostB)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		t.Log("Closing client")
		remote.Stop(client)
	})

	err = remote.WaitFor(clientStdOut, "INITIALIZING to LISTENING on INIT_COMPLETE", 20*time.Second)
	if err != nil {
		t.Log(*clientStdErr)
		t.Fatal(err)
	}
	t.Log("=== Client started ===")

	// Watch client logs for synchronisation with server
	foundSyncMessage := false
	period := 20 * time.Second
	endTime := time.Now().Add(period)
	log.Printf("Waiting until %s", endTime)
	for time.Now().Before(endTime) {
		before, after, found := strings.Cut(*clientStdOut, "\n")
		if found {
			*clientStdOut = after
			//log.Println("CLIENT | ", before)
			fields := strings.Fields(before)
			if fields[1] == "master" && fields[2] == "offset" {
				foundSyncMessage = true
				log.Printf("Synchronising. Offset %snS", fields[3])
				break
			}
		}
	}

	if !foundSyncMessage {
		t.Fatal("Synchronising failed!")
	}

}
