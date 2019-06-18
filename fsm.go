package windigo

import (
	"errors"
	"fmt"
	"reflect"

	termbox "github.com/nsf/termbox-go"
)

type RetCode int

const (
	Fail RetCode = iota - 1
	Ok
	Repeat
	Nop
)

type FiniteState int

type Transition struct {
	SrcState FiniteState
	Rc       RetCode
	DstState FiniteState
}

type WidgetStateFunc func(ev *termbox.Event) *Event
type GadgetStateFunc func(ev *Event) *Event

type FiniteStateMachine struct {
	CurrentState, EntryState, ExitState FiniteState
	// On screen representation of widget. An array of termbox Cells
	// representing enough information for the widget writer's supplied
	// Refresh to draw the widget via the widget's SetCell call.
	// This may represent every cell of the widget or as in
	// the case of a scrollbar, plus, minus and filler characters.
	// This info is separate for each "state", so each state can
	// represent the widget via different colors and or characters.
	Sigil []Sigil
	// State functions with index representing state.
	StateFunc []WidgetStateFunc
	// Table of src, rc, dst transitions, where src and dst are
	// indices(states) into the statefunc array, and rc is one of the
	// Finite State Machine return codes(Ok, Fail, Repeat, Nop).
	Transitions []Transition
}

// NewFSM returns a State (the entry state for the new state machine),
// and a pointer to the new FSM struct.  It expects a Sigil for
// each of the widget's active states.  This may be 1 for an indicator
// or 1 or more for a button/switch.  An indicator is just a button with
// no clickable regions and only windigo events to transition states.
// Normally, widgets may transition states on windigo events or termbox
// input events.
//
// This function provides a simple round-robin state machine.
// For more complex state machines, create the new state machine with
// 0 states, and add the states and transition table entries manually
// with AddState and AddTransition.
// Calling NewFSM with no Sigils(state representations) will return
// a pointer to a new FSM, with NO state table entries and
// NO transition table entries and -1 as the entry point.
// You should then use AddState and AddTransition to populate the
// finite state machine's tables.  This allows for a more complex
// state machine (than the one provided here) to be used. State functions
// are: func (*termbbox.Event) *Event.  WidgetResult should be used
// for return values.
func NewFSM(w Widget, activeStates ...Sigil) *FiniteStateMachine {

	var entryState, exitState, activeState, firstActiveState FiniteState

	fsm := new(FiniteStateMachine)

	entryState = FiniteState(-1)
	entry := func(ev *termbox.Event) *Event { return WidgetResult(Ok, 0) }
	exit := func(ev *termbox.Event) *Event { return WidgetResult(Ok, 0) }
	active := func(ev *termbox.Event) *Event {
		if ev.Type == termbox.EventMouse && ev.Key == termbox.MouseLeft {
			//w.Press(250)
			return WidgetResult(Repeat, 1)
		}
		return WidgetResult(Nop, 0)
	}

	n := len(activeStates)

	if n > 0 {
		glyph := activeStates[0]
		entryState = fsm.AddState(entry, glyph)
		exitState = fsm.AddState(exit, glyph)
		// Add first active state.
		firstActiveState = fsm.AddState(active, glyph)
		activeState = firstActiveState
		// Add entry state transitions.
		fsm.AddTransition(entryState, Ok, activeState)
		fsm.AddTransition(entryState, Fail, exitState)
		// No transitions are necessary for the exit state.

		// Add Transition table entries for 1st active state.
		// except for the one that references the transition
		// to the next state, as it doesn't exist yet.
		fsm.AddTransition(activeState, Fail, exitState)
		fsm.AddTransition(activeState, Repeat, activeState)
		fsm.AddTransition(activeState, Nop, activeState)

		// Add additional states creating round-robin transition table
		// entries.
		for s := FiniteState(1); s < FiniteState(n); s++ {
			as := fsm.AddState(active, activeStates[s])
			// Add transition for previous state.
			fsm.AddTransition(activeState, Ok, as)
			// Add transition entries for this state.
			fsm.AddTransition(as, Fail, exitState)
			fsm.AddTransition(as, Repeat, as)
			fsm.AddTransition(as, Nop, as)

			activeState = as
		}
		fsm.AddTransition(activeState, Ok, firstActiveState)

	}

	fsm.EntryState = entryState
	fsm.ExitState = exitState
	fsm.CurrentState = entryState
	return fsm
}

// For use in a widget's finite state machine state functions.
// Widgets are finite state machines that turn termbox.Events
// into Windigo EventOut Events.  These may be 0 or more ints,
// strings, or termbox.Events (altho this is intended for a
// passthru of a single termbox event.)
//
// Result constructs and returns a Windigo event from
// RetCode and 0 or more int, string or termbox.Event results.
// RetCode is the Finite State Machine state transition value,
// Ok, Fail, Repeat or Nop.  Ok typically transitions to the next
// state, Fail transitions to exit, Repeat returns to the first
// active state, and Nop is the same as repeat, but generates NO
// windigo event.  The termbox.Event result is meant
// to provide a passthru function.
func WidgetResult(rc RetCode, n ...interface{}) *Event {
	e := NewEvent(WindEventOutput)
	e.Result.Rc = rc
	e.Result.Type = None

	for _, x := range n {
		switch v := reflect.ValueOf(x); v.Kind() {
		case reflect.String:
			e.Result.Sval = append(e.Result.Sval, v.String())
			if e.Result.Type == None {
				e.Result.Type = String
			}
		case reflect.Int:
			e.Result.Val = append(e.Result.Val, int(v.Int()))
			if e.Result.Type == None {
				e.Result.Type = Int
			}
		case reflect.Ptr:
			vp := reflect.New(reflect.TypeOf(v))
			str := fmt.Sprintf("%s", vp.Elem().Type())
			vp.Elem().Set(reflect.ValueOf(v))
			vp = vp.Elem()
			if str == "*termbox.Event" {
				e.Result.Tbox = vp.Interface().(*termbox.Event)
				if e.Result.Type == None {
					e.Result.Type = PassThru
				}
			}
		default:
		}
	}

	return e
}

// AddState adds a StateFunc to the widget's Finite State Machine
// statefunc table and establishes the characters and colors
// associated with it's on screen look.  These are used by
// the widget writer's Refresh method to actually draw the widget
// on screen.
func (fsm *FiniteStateMachine) AddState(f WidgetStateFunc, glyph Sigil) FiniteState {
	s := len(fsm.Sigil)
	fsm.StateFunc = append(fsm.StateFunc, f)
	fsm.Sigil = append(fsm.Sigil, glyph)
	return FiniteState(s)
}

// There are no checks done here to ensure that destination states
// actually exist.  If the destination state doesn't exist the state
// machine's NextState function will return Fail when it attempts to
// transition to  that state.  Source states are not a problem b/c
// if the source state doesn't exist, the entry will simply never
// be used.
func (fsm *FiniteStateMachine) AddTransition(src FiniteState, rc RetCode, dst FiniteState) {

	fsm.Transitions = append(fsm.Transitions, Transition{src, rc, dst})
}

func (fsm *FiniteStateMachine) NextState(rc RetCode) (FiniteState, error) {
	for i, _ := range fsm.Transitions {
		if fsm.CurrentState == fsm.Transitions[i].SrcState &&
			rc == fsm.Transitions[i].Rc {
			dst := fsm.Transitions[i].DstState
			if int(dst) < 0 || int(dst) >= len(fsm.StateFunc) {
				err := errors.New("state machine transition table error: destination state out of range")
				return fsm.CurrentState, err
			}
			return dst, nil
		}
	}
	err := errors.New("state machine transition table error: no entry matching source state and given RetCode")
	return fsm.CurrentState, err
}

// Getter and Setter methods.
//
func (fsm *FiniteStateMachine) SetState(s FiniteState) {
	if fsm != nil {
		fsm.CurrentState = s
	}
}

func (fsm *FiniteStateMachine) State() FiniteState {
	if fsm != nil {
		return fsm.CurrentState
	}
	return FiniteState(-1)
}

func (fsm *FiniteStateMachine) Entry() FiniteState {
	return fsm.EntryState
}

func (fsm *FiniteStateMachine) SetEntry(e FiniteState) {
	fsm.EntryState = e
}

func (fsm *FiniteStateMachine) Exit() FiniteState {
	return fsm.ExitState
}

func (fsm *FiniteStateMachine) SetExit(e FiniteState) {
	fsm.ExitState = e
}
