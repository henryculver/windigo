package windigo

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	runewidth "github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

// Alias these from termbox.
type (
	InputMode  termbox.InputMode  // int
	OutputMode termbox.OutputMode // int
	EventType  termbox.EventType  // uint8
	Modifier   termbox.Modifier   // uint8
	Key        termbox.Key        // uint16 Attribute  termbox.Attribute
)

const (
	EventKey       EventType = EventType(termbox.EventKey)
	EventResize    EventType = EventType(termbox.EventResize)
	EventMouse     EventType = EventType(termbox.EventMouse)
	EventError     EventType = EventType(termbox.EventError)
	EventInterrupt EventType = EventType(termbox.EventInterrupt)
	EventRaw       EventType = EventType(termbox.EventRaw)
	EventNone      EventType = EventType(termbox.EventNone)
	EventWindigo   EventType = EventNone + 1 + iota
)

type Cell struct {
	Ch     rune
	Fg, Bg Attribute
}

type IChing struct {
	Yin  chan *Event
	Yang chan *Event
}

//type WidgetStateFunc func(ev *termbox.Event) *Event
//type GadgetStateFunc func(ev *Event) *Event

type Attribute termbox.Attribute

type Color struct {
	Fg, Bg Attribute
}

// Alias some useful termbox constants.
const (
	ColorDefault Attribute = Attribute(termbox.ColorDefault)
	ColorBlack   Attribute = Attribute(termbox.ColorBlack)
	ColorRed     Attribute = Attribute(termbox.ColorRed)
	ColorGreen   Attribute = Attribute(termbox.ColorGreen)
	ColorYellow  Attribute = Attribute(termbox.ColorYellow)
	ColorBlue    Attribute = Attribute(termbox.ColorBlue)
	ColorMagenta Attribute = Attribute(termbox.ColorMagenta)
	ColorCyan    Attribute = Attribute(termbox.ColorCyan)
	ColorWhite   Attribute = Attribute(termbox.ColorWhite)
)

const AttrBold Attribute = Attribute(termbox.AttrBold)
const AttrReverse Attribute = Attribute(termbox.AttrReverse)
const AttrUnderline Attribute = Attribute(termbox.AttrUnderline)

var defWindowFG = ColorBlue
var defWindowBG = ColorBlack

type Poser interface {
	Loc() (int, int)
	SetLoc(int, int)
}

type Sizer interface {
	Size() (int, int)
	SetSize(int, int)
}

type Embelishments interface {
	Colors() (Attribute, Attribute)
	SetColors(Attribute, Attribute)
}

type PosSizer interface {
	Poser
	Sizer
}

type PosSizeColor interface {
	Poser
	Sizer
	Embelishments
}

type Keyboarder interface {
	HaveFocus() bool
	ReqFocus() (chan *termbox.Event, error)
	SetAllowFocus(bool)
}

type Mouser interface {
	Click(FiniteState, *Region, termbox.Key)
	RegClickable(*Region) (chan *termbox.Event, error)
}

type Graviton interface {
	Gravity() GravityType
	SetGravity(GravityType)
}

type Resizer interface {
	Elastic() ElasticType
	SetElastic(ElasticType)
}

type Container interface {
	Object
	Manage(Object) error
	AddChild(Object)
	Done()
}

type Communicator interface {
	AddComm(IChing)
	GetComm() []IChing
}

type Genealogy interface {
	Managed() bool
	SetManaged()
	Ancestor() Container
	SetAncestor(Container)
}

type Object interface {
	Poser
	Sizer
	Embelishments
	Genealogy
	Communicator
	Refresh()
	Init() error
}

// SetCell function that should be used by a widget's SetCell method.
func SetCell(w Object, x, y int, r rune, fg, bg Attribute) error {
	if w.Managed() {
		p := w.Ancestor()
		if p != nil {
			// add dx and dy
			dx, dy := w.Loc()
			w, h := w.Size()
			if x < 0 || x >= w || y < 0 || y >= h {
				str := fmt.Sprintf("SetCell: %d, %d out of range: %d %d", x, y, w, y)
				err := errors.New(str)
				return err
			}
			err := SetCell(p, x+dx, y+dy, r, fg, bg)
			return err
		} else {
			// We are managed and our parent is nil.  We are the root win.
			termbox.SetCell(x, y, r, termbox.Attribute(fg),
				termbox.Attribute(bg))
			return nil
		}
	}
	err := errors.New("SetCell: unmanaged object")
	return err
}

func RegClickable(w Object, r Region) (chan *termbox.Event, error) {
	if w.Managed() {
		p := w.Ancestor()
		if p != nil {
			// add dx and dy
			dx, dy := w.Loc()
			r.X += dx
			r.Y += dy
			c, err := RegClickable(p, r)
			return c, err
		} else {
			cr := newClickableRegion(r.X, r.Y, r.W, r.H)
			c := cr.C
			screen.clickableRegions = append(screen.clickableRegions, *cr)
			return c, nil
		}
	}
	err := errors.New("registerClickable: unmanaged object")
	return nil, err
}

// Make WindowType a Window.
func (w *WindowType) Main() {
}

func Init() *WindowType {

	err := termbox.Init()

	if err != nil {
		panic(err)
	}

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	termbox.SetOutputMode(termbox.Output256)
	ScreenSizeX, ScreenSizeY = termbox.Size()
	screen.W = ScreenSizeX
	screen.H = ScreenSizeY

	OutlineChars = make([]rune, len(defOutlineChars))
	copy(OutlineChars, defOutlineChars[:])

	r0 := new(Region)
	r0.X = 0
	r0.Y = 0
	r0.W = ScreenSizeX
	r0.H = ScreenSizeY
	r0.Valid = true
	r0.RightMost = true
	r0.BottomMost = true

	win := NewWindow(r0, defWindowFG, defWindowBG)

	screen.Comm.Yin = make(chan *Event)
	screen.Comm.Yang = make(chan *Event)
	c := new(IChing)
	c.Yin = screen.Comm.Yang
	c.Yang = screen.Comm.Yin

	win.Comm = append(win.Comm, *c)

	win.Elastic = defWindowElastic
	win.Parent = nil
	win.managed = true

	win.addEdges2Lines()

	screen.windows = append(screen.windows, *win)

	go screen.InputEventRouter()

	return win
}

func Close() {
	termbox.Close()
}

func Flush() {
	termbox.Flush()
}

func InitObject(w Object) {
	w.SetLoc(-1, -1)
	w.SetSize(-1, -1)
	w.SetColors(ColorWhite, ColorBlack)
	switch v := w.(type) {
	case *WidgetType:
		v.kbd = -1
	}
}

func SetRegion(w Object, r *Region) {
	w.SetLoc(r.X, r.Y)
	w.SetSize(r.W, r.H)
	v, ok := w.(*WindowType)
	if ok {
		v.Layout.Regions = append(v.Layout.Regions, *r)
	}
}

func Yin(w Object) (chan *Event, error) {

	if !w.Managed() {
		err := errors.New("Yin: cannot request input channel for unmanaged Object")
		return nil, err
	}

	comm := w.GetComm()
	yin := comm[0].Yin

	return yin, nil
}

func Yang(w Object) (chan *Event, error) {

	if !w.Managed() {
		err := errors.New("Yin: cannot request output channel for unmanaged Object")
		return nil, err
	}

	comm := w.GetComm()
	yang := comm[0].Yang

	return yang, nil
}

func Wprint(o Object, x, y int, fg, bg Attribute, msg string) error {
	var err error

loop:
	for _, c := range msg {
		err = SetCell(o, x, y, c, fg, bg)
		if err != nil {
			break loop
		}
		x += runewidth.RuneWidth(c)
	}
	return err
}

func DrawScale(o Object) error {

	var err error

	w, h := o.Size()
	fg, bg := o.Colors()
loop1:
	for i := 0; i < w; i++ {
		s := fmt.Sprintf("%.1d", i%10)
		r := rune(s[0])
		err = SetCell(o, i, 0, r, fg, bg)
		if err != nil {
			break loop1
		}
	}
loop2:
	for i := 0; i < h; i++ {
		s := fmt.Sprintf("%.1d", i%10)
		r := rune(s[0])
		err = SetCell(o, 0, i, r, fg, bg)
		if err != nil {
			break loop2
		}
	}
	return err
}

func Fill(o Object, c rune) error {
	var err error

	w, h := o.Size()
	fg, bg := o.Colors()
loop:
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			err = SetCell(o, i, j, c, fg, bg)
			if err != nil {
				break loop
			}
		}
	}
	return err
}

func Within(o Sizer, p *Point) bool {
	w, h := o.Size()
	if p.X >= 0 && p.X < w && p.Y >= 0 && p.Y < h {
		return true
	}
	return false
}

func AddComm(c Container, w Object) {

	comm := new(IChing)
	comm.Yin = make(chan *Event)
	comm.Yang = make(chan *Event)

	c.AddComm(*comm)

	t := comm.Yin
	comm.Yin = comm.Yang
	comm.Yang = t

	w.AddComm(*comm)
}

func EventMgr(o Object) {
	comm := o.GetComm()
	var selectCase = make([]reflect.SelectCase, len(comm))

	for i, _ := range comm {
		selectCase[i].Dir = reflect.SelectRecv
		selectCase[i].Chan = reflect.ValueOf(comm[i].Yin)
	}

	for {
		chosen, recv, recvOk := reflect.Select(selectCase)
		if recvOk {
			// use recv.Pointer() which is uintptr cast of *Event.
			wev := *(*Event)(unsafe.Pointer(recv.Pointer()))
			// use EventType and chosen to determine which
			// child object produced the event.
			_ = fmt.Sprintf("%T", wev, chosen)
		}
	}
}

func Manage(c Container, w Object) error {

	if w.Managed() {
		err := errors.New("object already managed")
		return err
	}

	// Manage has to check that the object to be managed
	// fits within the container's range.
	cw, ch := c.Size()
	ww, wh := w.Size()
	wx, wy := w.Loc()
	fmt.Sprintf("%T %T %T %T %T %T", cw, ch, ww, wh, wx, wy)

	w.SetManaged()
	w.SetAncestor(c)
	AddComm(c, w)

	c.AddChild(w)

	// Call the managed object's Init() which should start the
	// object's EventMgr.
	// w.Init()
	//
	// Use Wait Group.
	// Start managed object's eventmgr.
	// A widget's eventmgr will also start it's InputEventMgr (if it
	// has requested focus or registered clickables.
	return nil
}
