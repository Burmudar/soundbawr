package fsm

import (
	"fmt"
	"log"
)

var StateToString map[State]string = make(map[State]string)

type UnkownStateTransition struct {
	from, to State
}

func (e *UnkownStateTransition) Error() string {
	return fmt.Sprintf("No state transition exists %v -> %v\n", e.from, e.to)
}

type State int

func (s State) String() string {
	if v, ok := StateToString[s]; ok {
		return v
	}
	return fmt.Sprintf("Unknown: %d", s)
}

type Callback func(oldState, newState State)

func DebugCallback(oldState, newState State) {
	log.Println("========================")
	log.Printf("Current State: %s\n", oldState)
	log.Printf("Transition: %s -> %s\n", oldState, newState)
	log.Printf("New Current State: %s\n", newState)
	log.Println("========================")
}

type FSM interface {
	CurrentState() State
	Transition(newState State) error
}

type fsm struct {
	currentState State
	transitions  map[State][]State
	callbacks    []Callback
}

func New(initialState State, transitions map[State][]State, callbacks []Callback) *fsm {
	return &fsm{
		currentState: initialState,
		transitions:  transitions,
		callbacks:    callbacks,
	}
}

func (fsm *fsm) CurrentState() State {
	return fsm.currentState
}

func (fsm *fsm) fireOnTransition(oldState, newState State) {
	for _, cb := range fsm.callbacks {
		cb(oldState, newState)
	}
}

func (fsm *fsm) Transition(newState State) error {
	if fsm.currentState == newState {
		return nil
	}

	states, ok := fsm.transitions[fsm.currentState]

	if !ok {
		return fmt.Errorf("No transitions exist from: %v", fsm.currentState)
	}

	canTransition := false
	for _, state := range states {
		if state == newState {
			canTransition = true
		}
	}

	if !canTransition {
		return &UnkownStateTransition{
			fsm.currentState,
			newState,
		}

	}

	oldState := fsm.currentState
	fsm.currentState = newState
	fsm.fireOnTransition(oldState, fsm.currentState)
	return nil
}
