package windigo

import termbox "github.com/nsf/termbox-go"

var screen Screen

// The size of the initial screen, set from termbox.Size().
var ScreenSizeX, ScreenSizeY int

type Screen struct {
	WidthHeight
	windows          []WindowType
	clickableRegions []ClickableRegion
	kbdChannel       chan *termbox.Event
	Comm             IChing
}

// Register interest in mouse input events by calling RegClickable
// with the following struct.
type ClickableRegion struct {
	X, Y int
	W, H int
	C    chan *termbox.Event
}

func newClickableRegion(x, y int, w, h int) *ClickableRegion {
	cr := new(ClickableRegion)
	cr.X = x
	cr.Y = y
	cr.W = w
	cr.H = h
	cr.C = make(chan *termbox.Event)
	return cr
}

func (s *Screen) Size() (int, int) {
	return s.W, s.H
}

func (s *Screen) RequestFocus() (chan *termbox.Event, error) {
	if s.kbdChannel != nil {
		close(s.kbdChannel)
	}
	s.kbdChannel = make(chan *termbox.Event)
	return s.kbdChannel, nil
}

func (s *Screen) InputEventRouter() {

	var n int = 0
	_ = n
mainloop:
	for {
		// If it was a mouse event, determine if event occurred
		// in registered region.  If so, send event down
		// the registered channel.
		// If it was a key event, send event down s.kbdChannel.
		switch ev := termbox.PollEvent(); ev.Type {
		// key events follow focus
		case termbox.EventKey:
			if s.kbdChannel != nil {
				s.kbdChannel <- &ev
			}
		case termbox.EventMouse:
			for _, r := range s.clickableRegions {
				if ev.MouseX >= r.X && ev.MouseX < r.X+r.W {
					if ev.MouseY >= r.Y && ev.MouseY < r.Y+r.H {
						// adjust coordinates of event within
						// region of object that registered it.
						ev.MouseX = ev.MouseX - r.X
						ev.MouseY = ev.MouseY - r.Y
						// Send the event to the object that owns
						// this cell.
						r.C <- &ev
					}
				}
			}
		case termbox.EventError:
			panic(ev.Err)
			break mainloop
		}
	}

}
