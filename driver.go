package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

const stateFile = "local-mapping.json"

type localPersistDriver struct {
	mutex *sync.Mutex

	name      string
	volumes   map[string]string
	statePath string
}

type saveData struct {
	State map[string]string `json:"state"`
}

func newLocalPersistDriver(stateDir string) (*localPersistDriver, error) {
	debug := os.Getenv("DEBUG")
	if ok, _ := strconv.ParseBool(debug); ok {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.WithField("stateDir", stateDir).Debug("Driver init ...")

	driver := &localPersistDriver{
		volumes:   map[string]string{},
		mutex:     &sync.Mutex{},
		name:      "local-mapping",
		statePath: path.Join(stateDir, stateFile),
	}

	os.Mkdir(stateDir, 0700)

	driver.volumes, _ = driver.findExistingVolumesFromStateFile()

	logrus.WithField("volumes", driver.volumes).Debugf("Init success, found %s volumes", strconv.Itoa(len(driver.volumes)))
	return driver, nil
}

func (driver *localPersistDriver) Create(req *volume.CreateRequest) error {
	logrus.WithField("args", req).Debug("Create called")

	mountpoint := req.Options["mountpoint"]
	if mountpoint == "" {
		return logError("The `mountpoint` option is required")
	}

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	if driver.exists(req.Name) {
		return logError("The volume %s already exists", req.Name)
	}

	err := os.MkdirAll(mountpoint, 0755)
	if err != nil {
		return logError("Could not create directory %s", " ", mountpoint)
	}

	driver.volumes[req.Name] = mountpoint
	e := driver.saveState(driver.volumes)
	if e != nil {
		return logError(e.Error())
	}

	logrus.WithField("volume", req.Name).WithField("mountpoint", mountpoint).Debug("Create success")
	return nil
}

func (driver *localPersistDriver) List() (*volume.ListResponse, error) {
	logrus.Debug("List called")

	var volumes []*volume.Volume
	for name, _ := range driver.volumes {
		volumes = append(volumes, driver.volume(name))
	}

	logrus.WithField("volumes", driver.volumes).Debugf("List success, found %s volumes", strconv.Itoa(len(driver.volumes)))

	return &volume.ListResponse{
		Volumes: volumes,
	}, nil
}

func (driver *localPersistDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	logrus.WithField("args", req).Debug("Get Called")

	if driver.exists(req.Name) {
		logrus.WithField("volume", req.Name).Debug("Get success")
		return &volume.GetResponse{
			Volume: driver.volume(req.Name),
		}, nil
	}

	return &volume.GetResponse{}, logError("Volume %s not found", req.Name)
}

func (driver *localPersistDriver) Remove(req *volume.RemoveRequest) error {
	logrus.WithField("args", req).Debug("Remove Called")
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	delete(driver.volumes, req.Name)

	err := driver.saveState(driver.volumes)
	if err != nil {
		return logError(err.Error())
	}

	logrus.WithField("volume", req.Name).Debug("Remove success")
	return nil
}

func (driver *localPersistDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	logrus.WithField("args", req).Debug("Path Called")

	mountPoint, err := driver.volumes[req.Name]
	if !err {
		return &volume.PathResponse{}, logError("Volume %s not found", req.Name)
	}

	logrus.WithField("mountPoint", mountPoint).Debug("Path success")
	return &volume.PathResponse{Mountpoint: mountPoint}, nil
}

func (driver *localPersistDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	logrus.WithField("args", req).Debug("Mount Called")

	mountPoint, err := driver.volumes[req.Name]
	if !err {
		return &volume.MountResponse{}, logError("Volume %s not found", req.Name)
	}

	logrus.WithField("mountPoint", mountPoint).Debug("Mount success")
	return &volume.MountResponse{Mountpoint: mountPoint}, nil
}

func (driver *localPersistDriver) Unmount(req *volume.UnmountRequest) error {
	logrus.WithField("args", req).Debug("Unmount Called")

	logrus.WithField("volume", req.Name).Debug("Unmount success")
	return nil
}

func (driver *localPersistDriver) Capabilities() *volume.CapabilitiesResponse {
	logrus.Debug("Capabilities Called")

	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{Scope: "local"},
	}
}

func logError(format string, args ...interface{}) error {
	logrus.Errorf(format, args)
	return fmt.Errorf(format, args...)
}

func (driver *localPersistDriver) exists(name string) bool {
	return driver.volumes[name] != ""
}

func (driver *localPersistDriver) volume(name string) *volume.Volume {
	return &volume.Volume{
		Name:       name,
		Mountpoint: driver.volumes[name],
	}
}

func (driver *localPersistDriver) findExistingVolumesFromStateFile() (map[string]string, error) {
	fileData, err := ioutil.ReadFile(driver.statePath)
	if err != nil {
		return map[string]string{}, err
	}

	var data saveData
	e := json.Unmarshal(fileData, &data)
	if e != nil {
		return map[string]string{}, e
	}

	return data.State, nil
}

func (driver *localPersistDriver) saveState(volumes map[string]string) error {
	data := saveData{
		State: volumes,
	}

	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(driver.statePath, fileData, 0600)
}
