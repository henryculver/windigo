package windigo

import termbox "github.com/nsf/termbox-go"

type ResultType struct {
	Rc   RetCode
	Err  error
	Type PayloadType
	Val  []int
	Sval []string
	Cval chan *termbox.Event
	Tbox *termbox.Event
}

type ArgType struct {
	Type PayloadType
	Val  []int
	Sval []string
	Cval chan *termbox.Event
	Tbox *termbox.Event
}

type Event struct {
	EventType WindigoEventType
	Args      *ArgType
	Result    *ResultType
}

type WindigoEventType int

const (
	WindEventNone WindigoEventType = iota
	WindEventInit
	WindEventExit
	WindEventError
	WindEventRestart
	WindEventOutput
	WindEventMove
	WindEventResize
	nWindigoEvents
)

type PayloadType int

const (
	// StateFuncs always return a *Event with an Rc
	// set to inform the state machine.
	Int PayloadType = iota
	// passthru means that Result.Tbox is a valid termbox input
	// event and is being passed thru as the result of the widget.
	String
	Channel
	PassThru
	None
)

func NewEvent(et WindigoEventType, args ...interface{}) *Event {
	e := new(Event)
	e.Result = new(ResultType)
	e.EventType = et
	return e
}
