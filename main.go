package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/godbus/dbus"
	"github.com/sqp/pulseaudio"
)

type DeviceState uint32

const DeviceRunning DeviceState = 0
const DeviceIdle DeviceState = 1
const DeviceSuspended DeviceState = 2

func (s DeviceState) String() string {
	switch s {
	case DeviceRunning:
		return "RUNNING"
	case DeviceIdle:
		return "IDLE"
	case DeviceSuspended:
		return "SUSPENDED"
	}
	return "UNKNOWN"
}

type Client struct {
	*pulseaudio.Client
}

type OnDeviceStateUpdated interface {
	DeviceStateUpdated(dbus.ObjectPath, uint32)
}

func (cl *Client) NewPlaybackStream(path dbus.ObjectPath) {
	log.Println("NewPlaybackStream", path)
}

func (cl *Client) PlaybackStreamRemoved(path dbus.ObjectPath) {
	log.Println("PlaybackStreamRemoved", path)
}

func (cl *Client) DeviceVolumeUpdated(path dbus.ObjectPath, values []uint32) {
	log.Println("device volume", path, values)
}

func (cl *Client) StreamVolumeUpdated(path dbus.ObjectPath, values []uint32) {
	log.Println("stream volume", path, values)
}

func (cl *Client) NewSink(path dbus.ObjectPath, values []uint32) {
	log.Println("New Sink", path, values)
}
func (cl *Client) SinkRemoved(path dbus.ObjectPath, values []uint32) {
	log.Println("Sink Removed", path, values)
}
func (cl *Client) FallbackSinkUpdated(path dbus.ObjectPath, values []uint32) {
	log.Println("Fallback Sink updated", path, values)
}
func (cl *Client) FallbackSinkUnset(path dbus.ObjectPath, values []uint32) {
	log.Println("Fallback Sink unset", path, values)
}

func (cl *Client) DeviceActivePortUpdated(path dbus.ObjectPath, values []uint32) {
	log.Println("Device Active Port", path, values)
}

func (cl *Client) DeviceStateUpdated(path dbus.ObjectPath, state uint32) {
	var deviceState = DeviceState(state)
	log.Printf("State updated: %v %v", path, deviceState)
}

func UnkownSignal(s *dbus.Signal) {
	log.Println("Unkown Signal!")
	log.Println("Name", s.Name)
	log.Println("Path", s.Path)
	log.Println("Body", s.Body)
	log.Println("Sender", s.Sender)
}

func main() {
	pulseaudio.PulseCalls["Device.StateUpdated"] = func(m pulseaudio.Msg) {
		m.O.(OnDeviceStateUpdated).DeviceStateUpdated(m.P, m.D[0].(uint32))
	}
	pulseaudio.PulseTypes["Device.StateUpdated"] = reflect.TypeOf((*OnDeviceStateUpdated)(nil)).Elem()

	pulse, e := pulseaudio.New()
	if e != nil {
		log.Panicln("connect", e)
	}

	client := &Client{pulse}
	pulse.Register(client)
	defer pulse.Unregister(client)

	defer pulse.StopListening()

	sinks, e := client.Core().ListPath("Sinks")

	if len(sinks) == 0 {
		fmt.Println("no sinks to test")
		return
	}

	for _, sink := range sinks {
		dev := client.Device(sink)
		state, _ := dev.Uint32("State")
		name, _ := dev.String("Name")
		log.Printf("%s current state: %v", name, state)
	}
	fmt.Printf("Path: %v\n", sinks[1])
	client.ListenForSignal("Device.StateUpdated", sinks[1])
	client.SetOnUnknownSignal(UnkownSignal)

	client.Listen()
}
