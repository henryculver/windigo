package windigo

type ElasticType int

// Elastic is actually a bit field.  None = 00, Horz = 01, Vert = 10,
// Both = 11.
const (
	ElasticNone ElasticType = iota
	ElasticHorz
	ElasticVert
	ElasticBoth
)

// Default Elastic setting for windows and gadgets.
var defWindowElastic ElasticType = ElasticNone
