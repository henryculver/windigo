package windigo

type Point struct {
	X, Y int
}

type TopLeft Point
type NexusType Point

type nexusType int

const (
	nexusTopLeft nexusType = iota
	nexusTopRight
	nexusBottomLeft
	nexusBottomRight
	_
	_
	nexusLeftT
	nexusRightT
	nexusTopT
	nexusBottomT
	nexusCross
	nexusNone
)

type WidthHeight struct {
	W, H int
}

type Region struct {
	TopLeft
	WidthHeight
	RightMost, BottomMost bool
	Valid                 bool
}

func NewRegion(x, y, width, height int) *Region {

	r := new(Region)
	r.X = x
	r.Y = y
	r.W = width
	r.H = height
	r.Valid = true

	return r
}

func (r *Region) Loc() (x, y int) {
	return r.X, r.Y
}

func (r *Region) Size() (w, h int) {
	return r.W, r.H
}

type RegionType []Region

func (r RegionType) Len() int {
	return len(r)
}

func (r RegionType) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RegionType) Less(i, j int) bool {
	x1 := r[i].X + 1
	y1 := r[i].Y + 1
	x2 := r[j].X + 1
	y2 := r[j].Y + 1

	if x1 < x2 && y1 < y2 {
		return true
	}

	if x1 < x2 && y1 >= y2 {
		if y1 == y2 {
			return true
		}
		return false
	}

	if x1 >= x2 && y1 < y2 {
		if x1 == x2 {
			return true
		}
		return false
	}

	if x1 >= x2 && y1 >= y2 {
		return false
	}
	return false
}

// Return the integer absolute value.
func abs(i int) int {
	if i >= 0 {
		return i
	}
	return -i
}

// This is the left edge of the region in it's window's coordinates.
func (r *Region) Left() *Line {
	left := Line{P1: Point{X: r.X, Y: r.Y}, P2: Point{X: r.X, Y: r.Y + r.H - 1}}
	left.Orientation = Vertical
	return &left
}

func (r *Region) Right() *Line {
	right := Line{P1: Point{X: r.X + r.W - 1, Y: r.Y}, P2: Point{X: r.X + r.W - 1, Y: r.Y + r.H - 1}}
	right.Orientation = Vertical
	return &right
}

func (r *Region) Top() *Line {
	top := Line{P1: Point{X: r.X, Y: r.Y}, P2: Point{X: r.X + r.W - 1, Y: r.Y}}
	top.Orientation = Horizontal
	return &top
}

func (r *Region) Bottom() *Line {
	bottom := Line{P1: Point{X: r.X, Y: r.Y + r.H - 1}, P2: Point{X: r.X + r.W - 1, Y: r.Y + r.H - 1}}
	bottom.Orientation = Horizontal
	return &bottom
}
