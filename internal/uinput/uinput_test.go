package uinput

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestCloseDeviceClosesFileAfterDestroyFailure(t *testing.T) {
	file, err := ioutil.TempFile("", "not-uinput")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := closeDevice(file); err == nil {
		t.Fatal("expected the destroy ioctl to fail on a regular file")
	}
	if _, err := file.Write([]byte{0}); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("file remained open after a failed destroy ioctl: %v", err)
	}
}

func TestValidateDevicePathEmptyPathPanics(t *testing.T) {
	expected := "device path must not be empty"
	err := validateDevicePath("")
	if err.Error() != expected {
		t.Fatalf("Expected: %s\nActual: %s", expected, err)
	}
}

func TestValidateDevicePathInvalidPathPanics(t *testing.T) {
	path := "/some/bogus/path"
	err := validateDevicePath(path)
	if !os.IsNotExist(err) {
		t.Fatalf("Expected: os.IsNotExist error\nActual: %s", err)
	}
}

func TestValidateUinputNameEmptyNamePanics(t *testing.T) {
	expected := "device name may not be empty"
	err := validateUinputName(nil)
	if err.Error() != expected {
		t.Fatalf("Expected: %s\nActual: %s", expected, err)
	}
}

func TestFailedDeviceFileCreationGeneratesError(t *testing.T) {
	expected := "could not open device file"
	_, err := createDeviceFile("/root/testfile")
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error, but got none")
	}
}

func TestNonExistentDeviceFileCausesError(t *testing.T) {
	expected := "failed to write uidev struct to device file:"
	_, err := createUsbDevice(nil, uinputUserDev{})
	if err == nil {
		t.Fatalf("expected error, but got none")
	}
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("got '%v', but expected '%v'", err.Error(), expected)
	}
}
