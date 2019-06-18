package windigo

type PanelType struct {
	GadgetType
}

func NewPanel(r *Region, fg, bg Attribute) *PanelType {
	p := new(PanelType)
	p.X = r.X
	p.Y = r.Y
	p.W = r.W
	p.H = r.H
	p.Fg = fg
	p.Bg = bg
	return p
}

func (p *PanelType) Manage(o Object) error {
	Manage(p, o)
	o.Init()
	return nil
}
