package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-plugins-helpers/volume"
)

var (
	defaultTestStateDir   = "/tmp"
	defaultTestName       = "test-volume"
	defaultTestMountpoint = "/tmp/data/local-mapping-test"
	// Test, the root mountpoint is "/"
	defaultTestRootMountPoint = "/"
)

func TestCreate(t *testing.T) {
	driver, _ := newLocalPersistDriver(defaultTestRootMountPoint, defaultTestStateDir)

	defaultCreateHelper(driver, t)

	// test that a directory is created
	_, err := os.Stat(defaultTestMountpoint)
	if os.IsNotExist(err) {
		t.Error("Mountpoint directory was not created:", err.Error())
	}

	// test that volumes has one
	if len(driver.volumes) != 1 {
		t.Error("Driver should have exactly 1 volume")
	}

	defaultCleanupHelper(driver, t)

	// test that options are required
	err2 := driver.Create(&volume.CreateRequest{
		Name: defaultTestName,
	})

	if err2 == nil {
		t.Error("Should return error with no mountpoint option provided")
	}
	fmt.Print("\n")
}

func TestGet(t *testing.T) {
	driver, _ := newLocalPersistDriver(defaultTestRootMountPoint, defaultTestStateDir)

	defaultCreateHelper(driver, t)

	_, err := driver.Get(&volume.GetRequest{Name: defaultTestName})
	if err != nil {
		t.Error(err.Error() + "\n\n")
	}

	defaultCleanupHelper(driver, t)
}

func TestList(t *testing.T) {
	driver, _ := newLocalPersistDriver(defaultTestRootMountPoint, defaultTestStateDir)

	name := defaultTestName + "2"
	mountpoint := defaultTestMountpoint + "2"

	defaultCreateHelper(driver, t)
	res, err := driver.List()
	if err != nil {
		t.Error(err.Error())
	}
	if len(res.Volumes) != 1 {
		t.Error("Should have found 1 volume!")
	}

	createHelper(driver, t, name, mountpoint)
	res2, err2 := driver.List()
	if err2 != nil {
		t.Error(err2.Error())
	}
	if len(res2.Volumes) != 2 {
		t.Error("Should have found 1 volume!")
	}

	defaultCleanupHelper(driver, t)
	cleanupHelper(driver, t, name, mountpoint)
}

func TestMount(t *testing.T) {
	driver, _ := newLocalPersistDriver(defaultTestRootMountPoint, defaultTestStateDir)

	defaultCreateHelper(driver, t)

	// mount, mount and path should have same output (they all use Path under the hood)
	pathRes, pathErr := driver.Path(&volume.PathRequest{Name: defaultTestName})
	mountRes, mountErr := driver.Mount(&volume.MountRequest{Name: defaultTestName})
	unmountErr := driver.Unmount(&volume.UnmountRequest{Name: defaultTestName})

	if pathErr != nil {
		t.Error(pathErr.Error())
	}

	if mountErr != nil {
		t.Error(mountErr.Error())
	}

	if unmountErr != nil {
		t.Error(unmountErr.Error())
	}

	if !(pathRes.Mountpoint == mountRes.Mountpoint) {
		t.Error("Mount and Path should all return the same Mountpoint")
	}
	defaultCleanupHelper(driver, t)
}

func createHelper(driver *localPersistDriver, t *testing.T, name string, mountpoint string) {
	err := driver.Create(&volume.CreateRequest{
		Name: name,
		Options: map[string]string{
			"mountpoint": mountpoint,
		},
	})

	if err != nil {
		t.Error("ERROR: " + err.Error())
	}
	fmt.Print("\n")
}

func defaultCreateHelper(driver *localPersistDriver, t *testing.T) {
	createHelper(driver, t, defaultTestName, defaultTestMountpoint)
}

func cleanupHelper(driver *localPersistDriver, t *testing.T, name string, mountpoint string) {
	os.RemoveAll(mountpoint)

	_, err := os.Stat(mountpoint)
	if !os.IsNotExist(err) {
		t.Error("[Cleanup] Mountpoint still exists:", err.Error())
	}

	driver.Remove(&volume.RemoveRequest{Name: name})

	res, err2 := driver.Get(&volume.GetRequest{Name: name})
	if err2 == nil {
		t.Error("[Cleanup] Volume still exists:", res.Volume.Name)
	}
	fmt.Print("\n")
}

func defaultCleanupHelper(driver *localPersistDriver, t *testing.T) {
	cleanupHelper(driver, t, defaultTestName, defaultTestMountpoint)
}
