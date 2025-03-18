package ptp4l_tests

import (
	"os"
	"testing"

	"go-remote-exec/pkg/remote"
)

func TestPtp4l(t *testing.T) {
	//start := time.Now()

	hostA, err := remote.CreateHost("raspi-a.lan", "jpmeijers", os.Getenv("REMOTE_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		hostA.Close()
	})

	hostB, err := remote.CreateHost("raspi-b.lan", "jpmeijers", os.Getenv("REMOTE_PASSWORD"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		hostB.Close()
	})

	// Install linuxptp snap
	err = remote.Setup(hostA)
	if err != nil {
		t.Fatal(err)
	}
	err = remote.Setup(hostB)
	if err != nil {
		t.Fatal(err)
	}

	err = remote.StartServer(hostA)
	if err != nil {
		t.Fatal(err)
	}
	err = remote.StartClient(hostB)
	if err != nil {
		t.Fatal(err)
	}
	//
	//checkSync(hosta, hostb)
	//
	//stop(hosta)
	//stop(hostb)
}
