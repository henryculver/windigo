package windigo

import "errors"

// Default Box drawing characters.
const defTopLeft rune = 0x250C  // '┌' U+250C
const defTopRight rune = 0x2510 // '┐' U+2510
const defBotLeft rune = 0x2514  // '└' U+2514
const defBotRight rune = 0x2518 // '┘' U+2518
const defVertBar rune = 0x2502  // '│' U+2502
const defHorzBar rune = 0x2500  // '─' U+2500
const defLeftT rune = 0x251C    // '├' U+251C
const defRightT rune = 0x2524   // '┤' U+2524
const defTopT rune = 0x252C     // '┬' U+252C
const defBotT rune = 0x2534     // '┴' U+2534
const defCross rune = 0x253C    // '┼' U+253C

// The outline characters to use when drawing borders.
// There is also a set of const defOutlinechars.
type OutlineChar int

// This is the order that outline characters need to be in an
// outlinChar type
const (
	TL OutlineChar = iota
	TR
	BL
	BR
	VB
	HB
	LT
	RT
	TT
	BT
	CR
	nOutlineChars
)

var defOutlineChars = [nOutlineChars]rune{defTopLeft, defTopRight,
	defBotLeft, defBotRight, defVertBar, defHorzBar, defLeftT, defRightT,
	defTopT, defBotT, defCross}

var OutlineChars []rune

type Line struct {
	P1, P2      Point
	Orientation OrientationType
}

func NewLine(x1, y1, x2, y2 int) *Line {

	line := new(Line)
	line.P1.X = x1
	line.P1.Y = y1
	line.P2.X = x2
	line.P2.Y = y2

	if x1 == x2 && y1 == y2 {
		line.Orientation = UnOriented
	} else {
		if x1 == x2 {
			line.Orientation = Vertical
		} else {
			line.Orientation = Horizontal
		}
	}
	return line
}

type OrientationType int

const (
	UnOriented OrientationType = iota
	Horizontal
	Vertical
)

// Is the line horizontal?
func (l *Line) horizontal() bool {
	if l.Orientation == Horizontal {
		return true
	}
	return false
}

// Is the line vertical?
func (l *Line) vertical() bool {

	if l.Orientation == Vertical {
		return true
	}
	return false
}

// Also, if they are parallel and distance is 0
func coincident(l1, l2 *Line) bool {
	if l1.horizontal() && l2.horizontal() {
		if l1.P1.Y == l2.P1.Y {
			return true
		}
		return false
	}
	if l1.vertical() && l2.vertical() {
		if l1.P1.X == l2.P1.X {
			return true
		}
		return false
	}
	return false
}

// What is the distance between 2 parallel lines?
func distance(l1, l2 *Line) (int, error) {

	if !parallel(l1, l2) {
		err := errors.New("Lines must be parallel")
		return 0, err
	}
	if coincident(l1, l2) {
		return 0, nil
	}

	if l1.horizontal() {
		return abs(l1.P1.Y - l2.P1.Y), nil
	}

	return abs(l1.P1.X - l2.P1.X), nil
}

// Are the two lines parallel?  In the restricted case of only
// horizontal and vertical lines, this is pretty straight forward.
func parallel(l1, l2 *Line) bool {

	if l1.horizontal() && l2.horizontal() {
		return true
	}
	if l1.vertical() && l2.vertical() {
		return true
	}
	return false
}

// Are the two lines orthogal?  Again, in the restricted case of only
// horizontal and vertical lines, this is straight forward.
func (l1 *Line) orthogonal(l2 *Line) bool {
	if l1.horizontal() && l2.vertical() {
		return true
	}
	if l1.vertical() && l2.horizontal() {
		return true
	}
	return false
}

// Return the two points defining the line from it's structure
// in left to right/top to bottom order.
func (l *Line) endPoints() (*Point, *Point) {

	if l.horizontal() {
		if l.P1.X < l.P2.X {
			return &l.P1, &l.P2
		}
	}
	if l.vertical() {
		if l.P1.Y < l.P2.Y {
			return &l.P1, &l.P2
		}
	}
	return &l.P2, &l.P1
}

// Is ol (other line) further down/to the right of l.
func (l *Line) GreaterThan(ol *Line) bool {
	if l.horizontal() {
		if l.P1.Y > ol.P1.Y {
			return true
		}
	}
	if l.vertical() {
		if l.P1.X > ol.P1.X {
			return true
		}
	}
	return false
}

// Is ol (other line) further up/to the left than l.
func (l *Line) LessThan(ol *Line) bool {
	if l.horizontal() {
		if l.P1.Y < ol.P1.Y {
			return true
		}
	}
	if l.vertical() {
		if l.P1.X < ol.P1.X {
			return true
		}
	}
	return false
}

// Is the point p on the line l?
func (p *Point) isOn(l *Line) bool {
	p1, p2 := l.endPoints()
	if l.horizontal() {
		if p.X >= p1.X && p.X <= p2.X && p.Y == p1.Y {
			return true
		}
	}
	if l.vertical() {
		if p.Y >= p1.Y && p.Y <= p2.Y && p.X == p1.X {
			return true
		}
	}
	return false
}

// Return the intersection of the 2 lines.
func Intersection(l1, l2 *Line) (*Point, error) {
	var x, y int

	p1, p2 := l1.endPoints()
	q1, q2 := l2.endPoints()
	_ = p2
	_ = q2
	if parallel(l1, l2) {
		err := errors.New("Lines must be orthogonal")
		return nil, err
	}

	if l1.vertical() {
		x = p1.X
		y = q1.Y
	} else {
		x = q1.X
		y = p1.Y
	}

	p := &Point{X: x, Y: y}
	if p.isOn(l1) && p.isOn(l2) {
		return p, nil
	}
	err := errors.New("Line segments do not intersect")
	return &Point{X: -1, Y: -1}, err
}

// Return the type of nexus 2 lines make onscreen
// TopLeft, TopRight, BottomLeft, BottomRight, LeftT, RightT, TopT,
// BottomT, Cross.
func nexus(l1, l2 *Line) (int, int, nexusType) {

	if !l1.orthogonal(l2) {
		return -1, -1, nexusNone
	}

	myl1 := l1
	myl2 := l2
	if l1.vertical() {
		myl1 = l2
		myl2 = l1
	}

	p, _ := Intersection(myl1, myl2)

	pMinus := Point{X: p.X - 1, Y: p.Y}
	pPlus := Point{X: p.X + 1, Y: p.Y}
	qMinus := Point{X: p.X, Y: p.Y - 1}
	qPlus := Point{X: p.X, Y: p.Y + 1}

	// 2 lines intersect, neither line ends there.  nexusCross.
	if pMinus.isOn(l1) && pPlus.isOn(l1) &&
		qMinus.isOn(l2) && qPlus.isOn(l2) {

		return p.X, p.Y, nexusCross
	}

	// nexusTopLeft
	if !pMinus.isOn(l1) && pPlus.isOn(l1) &&
		!qMinus.isOn(l2) && qPlus.isOn(l2) {

		return p.X, p.Y, nexusTopLeft
	}

	// nexusTopRight
	if pMinus.isOn(l1) && !pPlus.isOn(l1) &&
		!qMinus.isOn(l2) && qPlus.isOn(l2) {

		return p.X, p.Y, nexusTopRight
	}

	// nexusBottomLeft
	if !pMinus.isOn(l1) && pPlus.isOn(l1) &&
		qMinus.isOn(l2) && !qPlus.isOn(l2) {

		return p.X, p.Y, nexusBottomLeft
	}

	// nexusBottomRight
	if pMinus.isOn(l1) && !pPlus.isOn(l1) &&
		qMinus.isOn(l2) && !qPlus.isOn(l2) {

		return p.X, p.Y, nexusBottomRight
	}

	// nexusLeftT
	if !pMinus.isOn(l1) && pPlus.isOn(l1) &&
		qMinus.isOn(l2) && qPlus.isOn(l2) {

		return p.X, p.Y, nexusLeftT
	}

	// nexusRightT
	if pMinus.isOn(l1) && !pPlus.isOn(l1) &&
		qMinus.isOn(l2) && qPlus.isOn(l2) {

		return p.X, p.Y, nexusRightT
	}

	// nexusTopT
	if pMinus.isOn(l1) && pPlus.isOn(l1) &&
		!qMinus.isOn(l2) && qPlus.isOn(l2) {

		return p.X, p.Y, nexusTopT
	}

	// nexusBottomT
	if pMinus.isOn(l1) && pPlus.isOn(l1) &&
		qMinus.isOn(l2) && !qPlus.isOn(l2) {

		return p.X, p.Y, nexusBottomT
	}

	return -1, -1, nexusNone
}
