package main

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/Burmudar/soundbawr/fsm"
	Device "github.com/Burmudar/soundbawr/server/device"
	"github.com/godbus/dbus"
	"github.com/sqp/pulseaudio"
)

const (
	BarOff          fsm.State     = 0
	BarOn           fsm.State     = 1
	GracePeriod     fsm.State     = 2
	GracePeriodTime time.Duration = 15 * time.Minute
)

type Soundbar struct {
	fsm              fsm.FSM
	gracePeriodTimer *time.Timer
}

func NewSoundbar() *Soundbar {
	return &Soundbar{fsm.New(
		BarOff,
		map[fsm.State][]fsm.State{
			BarOff:      []fsm.State{BarOn},
			BarOn:       []fsm.State{GracePeriod},
			GracePeriod: []fsm.State{BarOff, BarOn},
		},
		[]fsm.Callback{commandDevice},
	), nil}
}

func (s *Soundbar) onAudioDeviceChange(state DeviceState) {
	switch state {
	case DeviceRunning:
		{
			s.AudioStarted()
			break
		}
	case DeviceSuspended:
	case DeviceIdle:
		{
			s.AudioStopped()
			break
		}
	default:
		{
			fmt.Printf("Device State: %s\n", state)
		}
	}
}

func (s *Soundbar) AudioStarted() {
	s.fsm.Transition(BarOn)
	if s.gracePeriodTimer != nil {
		s.gracePeriodTimer.Stop()
		s.gracePeriodTimer = nil
	}
}

func (s *Soundbar) AudioStopped() {
	if (s.fsm.CurrentState() != GracePeriod) && s.gracePeriodTimer == nil {
		s.fsm.Transition(GracePeriod)
		s.gracePeriodTimer = time.AfterFunc(GracePeriodTime, func() {
			s.fsm.Transition(BarOff)
		})
	}
}

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
	soundbar *Soundbar
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
	cl.soundbar.onAudioDeviceChange(deviceState)
}

func UnkownSignal(s *dbus.Signal) {
	log.Println("Unkown Signal!")
	log.Println("Name", s.Name)
	log.Println("Path", s.Path)
	log.Println("Body", s.Body)
	log.Println("Sender", s.Sender)
}

func commandDevice(old, new fsm.State) {
	if new == BarOn {
		log.Println("Sending command to turn sound bar ON")
		err := sendCommand(&Device.Command{
			Action: Device.Command_TURN_ON,
			Device: Device.Command_SOUND_BAR,
		})
		if err != nil {
			log.Fatalf("Failed to send command: %v\n", err)
		} else {
			log.Println("Command sent!")
		}
	}
}

func sendCommand(cmd *Device.Command) error {
	conn, err := net.Dial("tcp", "192.168.1.134:30000")
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		return err
	}

	var command = Device.Command{Action: Device.Command_TURN_ON, Device: Device.Command_SOUND_BAR}

	data, err := proto.Marshal(&command)
	if err != nil {
		return err
	}

	written, err := conn.Write(data)

	if err != nil {
		return err
	}

	log.Printf("Wrote %d bytes", written)
	return nil
}

func main() {
	fsm.StateToString[BarOff] = "BarOff"
	fsm.StateToString[BarOn] = "BarOn"
	fsm.StateToString[GracePeriod] = "GracePeriod"

	pulseaudio.PulseCalls["Device.StateUpdated"] = func(m pulseaudio.Msg) {
		m.O.(OnDeviceStateUpdated).DeviceStateUpdated(m.P, m.D[0].(uint32))
	}
	pulseaudio.PulseTypes["Device.StateUpdated"] = reflect.TypeOf((*OnDeviceStateUpdated)(nil)).Elem()

	pulse, e := pulseaudio.New()
	if e != nil {
		log.Panicln("connect", e)
	}

	client := &Client{pulse, NewSoundbar()}
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
		client.soundbar.onAudioDeviceChange(DeviceState(state))
	}
	fmt.Printf("Path: %v\n", sinks[1])
	client.ListenForSignal("Device.StateUpdated", sinks[1])
	client.SetOnUnknownSignal(UnkownSignal)

	client.Listen()
}
