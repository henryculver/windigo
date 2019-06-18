package windigo

import "sync"

type Gadget interface {
	Manage(Object)
}

type GadgetType struct {
	TopLeft
	WidthHeight
	MinSize WidthHeight

	Color
	Gravity GravityType
	Elastic ElasticType

	BackingStore []Cell

	// comm[n].yin and comm[n].yang channels for communication with
	// parent window(comm[0]) and subordinate widgets(n > 0).
	Children []Object
	Comm     []IChing

	managed bool
	wg      sync.WaitGroup
	Parent  Container
}

func NewGadget(r *Region, fg, bg Attribute) *GadgetType {
	g := new(GadgetType)
	g.X = r.X
	g.Y = r.Y
	g.W = r.W
	g.H = r.H
	g.Fg = fg
	g.Bg = bg
	// Gravity and Elastic will default to GravityNone and ElasticNone
	// managed will default to false and Parent, Children and Comm nil.

	return g
}

func (g *GadgetType) Loc() (int, int) {
	return g.X, g.Y
}

func (g *GadgetType) SetLoc(x, y int) {
	g.X = x
	g.Y = y
}

func (g *GadgetType) Size() (int, int) {
	return g.W, g.H
}

func (g *GadgetType) SetSize(width, height int) {
	g.W = width
	g.H = height
}

func (g *GadgetType) Colors() (Attribute, Attribute) {
	return g.Fg, g.Bg
}

func (g *GadgetType) SetColors(fg, bg Attribute) {
	g.Fg = fg
	g.Bg = bg
}

func (g *GadgetType) Ancestor() Container {
	return g.Parent
}

func (g *GadgetType) SetAncestor(p Container) {
	g.Parent = p
}

func (g *GadgetType) Init() error {
	g.wg.Add(1)
	go g.EventMgr()
	return nil
}

func (g *GadgetType) AddComm(c IChing) {
	g.Comm = append(g.Comm, c)
}

func (g *GadgetType) GetComm() []IChing {
	return g.Comm
}

func (g *GadgetType) AddChild(o Object) {
	g.Children = append(g.Children, o)
}

func (g *GadgetType) Managed() bool {
	return g.managed
}

func (g *GadgetType) SetManaged() {
	g.managed = true
}

func (g *GadgetType) Manage(o Object) error {
	Manage(g, o)
	o.Init()
	Flush()
	return nil
}

func (g *GadgetType) Clear() {
	Fill(g, ' ')
}

func (g *GadgetType) Refresh() {
	g.Clear()
	for _, w := range g.Children {
		w.Refresh()
	}
}

func (g *GadgetType) EventMgr() {
	EventMgr(g)
	p := g.Ancestor()
	p.Done()
}

func (g *GadgetType) Done() {
	g.wg.Done()
}
