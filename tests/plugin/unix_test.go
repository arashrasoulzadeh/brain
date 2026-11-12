package plugin_test

import (
	"fmt"
	"net"
	"os"
	"testing"

	"brain/internal/plugin"
)

func TestListenUnixCreatesPrivateTransportEndpoint(t *testing.T) {
	path := fmt.Sprintf("/tmp/brain-plugin-%d.sock", os.Getpid())
	defer os.Remove(path)
	listener, err := plugin.ListenUnix(path)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	connection, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	connection.Close()
}
