// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package simulator

import (
	simapi "github.com/onosproject/onos-api/go/onos/fabricsim"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	p4api "github.com/p4lang/p4runtime/go/p4/v1"
	"sync"
)

var log = logging.GetLogger("simulator", "device")

// DeviceSimulator simulates a single device
type DeviceSimulator struct {
	Device                   *simapi.Device
	Ports                    map[simapi.PortID]*simapi.Port
	Agent                    DeviceAgent
	ForwardingPipelineConfig *p4api.ForwardingPipelineConfig

	lock          sync.RWMutex
	roleElections map[uint64]*p4api.Uint128
}

// NewDeviceSimulator initializes a new device simulator
func NewDeviceSimulator(device *simapi.Device, agent DeviceAgent) *DeviceSimulator {
	log.Infof("Device %s: Creating simulator", device.ID)

	// Build a port map
	ports := make(map[simapi.PortID]*simapi.Port)
	for _, port := range device.Ports {
		ports[port.ID] = port
	}

	// Construct and return simulator from the device and the port map
	return &DeviceSimulator{
		Device:        device,
		Ports:         ports,
		Agent:         agent,
		roleElections: make(map[uint64]*p4api.Uint128),
	}
}

// Start spawns the device simulator background tasks and its agent API server, also in the background
func (ds *DeviceSimulator) Start(simulation *Simulation) error {
	log.Infof("Device %s: Starting simulator", ds.Device.ID)

	// Start any background simulation tasks

	// Starts the simulated device agent
	err := ds.Agent.Start(simulation, ds)
	if err != nil {
		log.Errorf("Device %s: Unable to run simulator: %+v", ds.Device.ID, err)
		return err
	}
	return nil
}

// Stop stops the device simulation agent and stops any background simulation tasks
func (ds *DeviceSimulator) Stop(mode simapi.StopMode) {
	log.Infof("Device %s: Stopping simulator using %s", ds.Device.ID, mode)
	if err := ds.Agent.Stop(mode); err != nil {
		log.Errorf("Device %s: Unable to stop simulator: %+v", ds.Device.ID, err)
	}

	// Stop any background simulation tasks
}

// EnablePort enables the specified simulated device port
func (ds *DeviceSimulator) EnablePort(id simapi.PortID) error {
	log.Infof("Device %s: Enabling port %s", ds.Device.ID, id)
	// TODO: Implement this
	// Look for any links or interfaces using this port and enable them
	return nil
}

// DisablePort disables the specified simulated device port
func (ds *DeviceSimulator) DisablePort(id simapi.PortID, mode simapi.StopMode) error {
	log.Infof("Device %s: Disabling port %s using %s", ds.Device.ID, id, mode)
	// TODO: Implement this
	// Look for any links or interfaces using this port and disable them
	return nil
}

// RecordRoleElection checks the given election ID for the specified role and records it
// if the given election ID is larger than a previously recorded election ID for the same
// role. It returns error (if election for role not secured) and the latest election ID for the role.
func (ds *DeviceSimulator) RecordRoleElection(role *p4api.Role, electionID *p4api.Uint128) (*p4api.Uint128, error) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	roleID := uint64(0)
	if role != nil {
		roleID = role.Id
	}

	maxID, ok := ds.roleElections[roleID]
	if !ok || isNewMaster(maxID, electionID) {
		ds.roleElections[roleID] = electionID
		return electionID, nil
	}
	return maxID, errors.NewInvalid("Mastership for role %d has not been secured with election ID %d",
		roleID, electionID)
}

func isNewMaster(current *p4api.Uint128, new *p4api.Uint128) bool {
	return current.High < new.High || (current.High == new.High && current.Low < new.Low)
}

// TODO: Additional simulation logic goes here
