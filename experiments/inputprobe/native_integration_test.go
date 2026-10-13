//go:build linux && cgo && integration

package main

import (
	"os"
	"testing"
)

func TestSyntheticLifecycle(t *testing.T) {
	if os.Getenv("TFD_UINPUT_TEST") != "1" {
		t.Skip("set TFD_UINPUT_TEST=1 with /dev/uinput access to run the real lifecycle test")
	}
	v, err := createVirtual(nil, false, "tfd-probe-integration")
	if err != nil {
		t.Fatal(err)
	}
	node := v.node()
	v.close()
	v.close()
	if node == "" {
		t.Fatal("created device had no event node identity")
	}
	// Compositor teardown is a separate external observation, not asserted here.
}
