package windigo

import termbox "github.com/nsf/termbox-go"

type ButtonType struct {
	WidgetType
}

func (b *ButtonType) Map(r *Region, ev *termbox.Event, cs FiniteState) interface{} {

	return nil
}

func (b *ButtonType) Init() error {

	x, y := b.Loc()
	w, h := b.Size()
	p := b.Ancestor()

	r := Region{TopLeft{x, y}, WidthHeight{w, h}, false, false, false}
	c, err := RegClickable(p, r)
	if err != nil {
		return err
	}
	b.InputChan = append(b.InputChan, c)

	b.Start()

	return nil
}

func NewButton(r *Region, activeStates ...Sigil) (*ButtonType, error) {

	b := new(ButtonType)
	b.X = r.X
	b.Y = r.Y
	b.W = r.W
	b.H = r.H
	b.Fsm = NewFSM(b, activeStates...)
	return b, nil
}
