package windigo

type GravityType int

// Note: (GravityTop | GravityBottom) or (GravityLeft | GravityRight)
//       makes no sense, however (GravityTop | GravityRight) and the
//       other 3 similar constructs should be useful.
const (
	GravityNone GravityType = 0
	GravityTop  GravityType = 1 << iota
	GravityBottom
	GravityLeft
	GravityRight
)

// Default gravity setting for new widgets.
var defWidgetGravity = GravityRight

// Default gravity setting for new gadgets.
var defWindowGravity = GravityTop
