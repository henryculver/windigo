package windigo

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

type LayoutType struct {
	Regions []Region
	Lines   []Line
}

type Window interface {
	Container
	Main()
}

type WindowType struct {
	border bool

	TopLeft
	WidthHeight

	Color
	Elastic ElasticType
	Gravity GravityType

	BackingStore []Cell

	Children []Object
	Comm     []IChing
	managed  bool
	Parent   Container
	wg       sync.WaitGroup

	Layout     LayoutType
	RightMost  bool
	BottomMost bool
}

func NewWindow(r *Region, fg, bg Attribute) *WindowType {

	w := new(WindowType)
	w.X = r.X
	w.Y = r.Y
	w.W = r.W
	w.H = r.H
	w.Fg = fg
	w.Bg = bg

	w.addEdges2Lines()
	w.Layout.Regions = append(w.Layout.Regions, *r)

	return w
}

// drawLine is used by AddBorder
func (w *WindowType) drawLine(l *Line) {
	p1, p2 := l.endPoints()
	fg, bg := w.Colors()
	if l.horizontal() {
		for i := p1.X + 1; i < p2.X; i++ {
			SetCell(w, i, p1.Y, OutlineChars[HB], fg, bg)
		}
	}

	if l.vertical() {
		for i := p1.Y + 1; i < p2.Y; i++ {
			SetCell(w, p1.X, i, OutlineChars[VB], fg, bg)
		}
	}
	return
}

// AddBorder takes the lines which were used to create regions to draw
// the border(s) for this window.  Instantiating the border has
// side effects.  The bottom/rightmost window will contain both a top
// AND bottom (left AND right) abstract line within the region.
// In an 80x24 window, you can draw a horizontal line at 12 and have 2
// regions, 12 high.  The line (once drawn) has to come from somewhere
// and since we are left-right/top-down oriented, the top(left) border
// is in the top(left) region and the new line IS the top(left)  border
// of the bottom(right) region.  This means that we can split the window
// into 2 regions 12 high, but if we draw borders, the top is then 11
// high and the bottom is 10 high.  The border belongs to the containing
// window.  That said, NOT instantiating a window's borders means
// subordinate objects may use the full area, including windows which
// draw their own borders.  Drawing borders in the containing window
// means that we can use one character to separate objects instead of 2.
// Drawing borders in the containing window may be a requirement of
// resizing.
func (w *WindowType) AddBorder() error {

	if w.Border() {
		err := errors.New("borders allready enabled on window")
		return err
	}

	w.border = true

	regions := w.Regions()
	for i, _ := range regions {

		regions[i].X += 1
		regions[i].Y += 1

		if regions[i].RightMost {
			regions[i].W -= 2
		} else {
			regions[i].W -= 1
		}
		if regions[i].BottomMost {
			regions[i].H -= 2
		} else {
			regions[i].H -= 1
		}
	}
	w.drawBorder()
	return nil
}

func (w *WindowType) drawBorder() {

	var nt nexusType
	var p Point

	hl := w.HorizLines()
	vl := w.VertLines()

	fg, bg := w.Colors()
	for _, l := range hl {
		w.drawLine(&l)
	}
	for _, l := range vl {
		w.drawLine(&l)
	}
	for _, l := range hl {
		for _, m := range vl {
			p.X, p.Y, nt = nexus(&l, &m)
			if nt != nexusNone {
				SetCell(w, p.X, p.Y, OutlineChars[OutlineChar(nt)], fg, bg)
			}
		}
	}
}

func (w *WindowType) addEdges2Lines() error {

	if len(w.Layout.Lines) > 0 {
		err := errors.New("Initializing non-empty lines")
		return err
	}
	width, height := w.Size()

	left := NewLine(0, 0, 0, height-1)
	right := NewLine(width-1, 0, width-1, height-1)
	top := NewLine(0, 0, width-1, 0)
	bottom := NewLine(0, height-1, width-1, height-1)

	w.Layout.Lines = append(w.Layout.Lines, *left)
	w.Layout.Lines = append(w.Layout.Lines, *right)
	w.Layout.Lines = append(w.Layout.Lines, *top)
	w.Layout.Lines = append(w.Layout.Lines, *bottom)

	return nil
}

// Return the position/dimensions of the region n.
func (w *WindowType) GetRegion(n int) *Region {

	if n < 0 || n >= len(w.Layout.Regions) {
		return nil
	}
	return &w.Layout.Regions[n]
}

// AddLine creates a new line orthogonal to l1 and l2 at a distnace n
// from p1 of the left/top line.  Each new line must have end points on
// an existing line.  When a New Window is created, the first
// 4 lines are established.  These represent the edges of the
// window. This will define a single "region" should GetRegion be called.
// The correct order is to add any additional lines to the parent window
// using NewLine and then call GetRegion/NewWindow to
// create the non-overlapping subordinate windows.  Although there is no
// reason this can't be done in any window, the idea was to allow one to
// establish regions, have the borders drawn and not waste any more
// real estate.
//
// Every new line divides at least one region, invalidating it/them and
// adding 2 for each one invalidated.
func (w *WindowType) AddLine(l1, l2 *Line, n int) (*Line, error) {

	var l *Line
	var length int

	myl1 := l1
	myl2 := l2

	if !parallel(myl1, myl2) {
		err := errors.New("Lines must be parallel")
		return nil, err
	}

	if l1.GreaterThan(l2) {
		myl1 = l2
		myl2 = l1
	}

	regions := w.Regions()

	if myl1.horizontal() {
		l = NewLine(myl1.P1.X+n, myl1.P1.Y, myl1.P1.X+n, myl2.P1.Y)
		l.Orientation = Vertical
		length = myl1.P2.X - myl1.P1.X + 1
	} else {
		l = NewLine(myl1.P1.X, myl1.P1.Y+n, myl2.P1.X, myl1.P1.Y+n)
		l.Orientation = Horizontal
		length = myl1.P2.Y - myl1.P1.Y + 1
	}

	if n < 0 || n > length {
		msg := fmt.Sprintf("%d exceeds this area's size %d", n, length)
		err := errors.New(msg)
		return nil, err
	}

	w.Layout.Lines = append(w.Layout.Lines, *l)

	// If one of the two provided lines (either actually, since they
	// are parallel) is horizontal, our line is vertical.

	// If line divides region, create 2 new regions and invalidate
	// the original.
	for i, r := range regions {

		if myl1.horizontal() {
			top := r.Top()
			bottom := r.Bottom()
			p1, _ := Intersection(l, top)
			p2, _ := Intersection(l, bottom)

			// Does l divide the region?
			if p1.isOn(top) && p2.isOn(bottom) {
				// Create Western region.
				r1 := NewRegion(r.X, r.Y, n-r.X, r.H)
				r1.RightMost = false
				r1.BottomMost = r.BottomMost

				// Create Eastern region.
				r2 := NewRegion(n, r.Y, r.W-(n-r.X), r.H)
				r2.RightMost = r.RightMost
				r2.BottomMost = r.BottomMost

				w.Layout.Regions[i].Valid = false
				w.Layout.Regions = append(w.Layout.Regions, *r1)
				w.Layout.Regions = append(w.Layout.Regions, *r2)
			}

		} else {
			left := r.Left()
			right := r.Right()

			p1, _ := Intersection(l, left)
			p2, _ := Intersection(l, right)

			// Does l divide the region?
			if p1.isOn(left) && p2.isOn(right) {
				// create northern region.
				r1 := NewRegion(r.X, r.Y, r.W, n-r.Y)
				r1.RightMost = r.RightMost
				r1.BottomMost = false

				// create southern region.
				r2 := NewRegion(r.X, r.Y+n, r.W, r.H-(n-r.Y))
				r2.RightMost = r.RightMost
				r2.BottomMost = r.BottomMost

				w.Layout.Regions[i].Valid = false
				w.Layout.Regions = append(w.Layout.Regions, *r1)
				w.Layout.Regions = append(w.Layout.Regions, *r2)
			}
		}
	}

	var myregions []Region
	for _, r := range w.Layout.Regions {
		if r.Valid {
			myregions = append(myregions, r)
		}
	}

	// SetRegions
	w.Layout.Regions = make([]Region, len(myregions))

	sort.Sort(RegionType(myregions))
	copy(w.Layout.Regions, myregions)

	return l, nil
}

// Make WindowType an Object.
func (w *WindowType) Init() error {
	go w.EventMgr()
	return nil
}

func (w *WindowType) Ancestor() Container {
	return w.Parent
}

func (w *WindowType) SetAncestor(p Container) {
	w.Parent = p
}

func (w *WindowType) Loc() (int, int) {
	return w.X, w.Y
}

func (w *WindowType) SetLoc(x, y int) {
	w.X = x
	w.Y = y
}

func (w *WindowType) Size() (int, int) {
	return w.W, w.H
}

func (w *WindowType) SetSize(width, height int) {
	w.W = width
	w.H = height
}

func (w *WindowType) Colors() (Attribute, Attribute) {
	return w.Fg, w.Bg
}

func (w *WindowType) SetColors(fg, bg Attribute) {
	w.Fg = fg
	w.Bg = bg
}

func (w *WindowType) Border() bool {
	return w.border
}

func (w *WindowType) AddComm(c IChing) {
	w.Comm = append(w.Comm, c)
}

func (w *WindowType) GetComm() []IChing {
	return w.Comm
}

func (w *WindowType) AddChild(o Object) {
	w.Children = append(w.Children, o)
}

func (w *WindowType) Managed() bool {
	return w.managed
}

func (w *WindowType) SetManaged() {
	w.managed = true
}

func (w *WindowType) Regions() []Region {

	return w.Layout.Regions
}

func (w *WindowType) Lines() []Line {

	return w.Layout.Lines
}

func (w *WindowType) Clear() {
	Fill(w, ' ')
}

func (w *WindowType) Refresh() {
	w.Clear()
	if w.Border() {
		w.drawBorder()
	}
	for _, o := range w.Children {
		o.Refresh()
	}
	Flush()
}

// This is the left edge of the window in it's container's coordinates.
func (w *WindowType) Left() *Line {
	_, height := w.Size()
	left := Line{P1: Point{X: 0, Y: 0}, P2: Point{X: 0, Y: height - 1}}
	left.Orientation = Vertical
	return &left
}

func (w *WindowType) Right() *Line {
	width, height := w.Size()
	right := Line{P1: Point{X: width - 1, Y: 0},
		P2: Point{X: width - 1, Y: height - 1}}
	right.Orientation = Vertical
	return &right
}

func (w *WindowType) Top() *Line {
	width, _ := w.Size()
	top := Line{P1: Point{X: 0, Y: 0},
		P2: Point{X: width - 1, Y: 0}}
	top.Orientation = Horizontal
	return &top
}

func (w *WindowType) Bottom() *Line {
	width, height := w.Size()
	bottom := Line{P1: Point{X: 0, Y: height - 1}, P2: Point{X: width - 1, Y: height - 1}}
	bottom.Orientation = Horizontal
	return &bottom
}

// From a window's "lines" array, that was used to create regions,
// return only the horizontal lines.
func (w *WindowType) HorizLines() []Line {
	var hl []Line

	lines := w.Lines()
	for _, l := range lines {
		if l.horizontal() {
			hl = append(hl, l)
		}
	}
	return hl
}

func (w *WindowType) VertLines() []Line {
	var vl []Line

	lines := w.Lines()
	for _, l := range lines {
		if l.vertical() {
			vl = append(vl, l)
		}
	}
	return vl
}

// Make WindowType a Container.
func (w *WindowType) Manage(o Object) error {
	Manage(w, o)
	o.Init()
	return nil
}

func (w *WindowType) EventMgr() {
	EventMgr(w)
}

func (w *WindowType) Done() {
	w.wg.Done()
}
