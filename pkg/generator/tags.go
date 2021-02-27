package generator

const (
	TagExample = "example"

	TagUpstreamE2E = "upstream e2e"

	TagEmptyPortSlice   = "empty port slice"
	TagSinglePortSlice  = "single port slice"
	TagTwoPlusPortSlice = "two or more port slice"
	TagNilPort          = "nil port"
	TagNumberedPort     = "numbered port"
	TagNamedPort        = "named port"
	TagNilProtocol      = "nil protocol"
	TagTCPProtocol      = "TCP protocol"
	TagUDPProtocol      = "UDP protocol"
	TagSCTPProtocol     = "SCTP protocol"
)

var AllTags = []string{
	TagEmptyPortSlice,
	TagNilPort,
	TagNumberedPort,
	TagNamedPort,
	TagNilProtocol,
	TagTCPProtocol,
	TagUDPProtocol,
	TagSCTPProtocol,
}

type StringSet map[string]bool

func NewStringSet(elems ...string) StringSet {
	dict := map[string]bool{}
	for _, e := range elems {
		dict[e] = true
	}
	return dict
}

func (s StringSet) Keys() []string {
	var slice []string
	for k := range s {
		slice = append(slice, k)
	}
	return slice
}

func (s StringSet) ContainsAny(slice []string) bool {
	for _, e := range slice {
		if _, ok := s[e]; ok {
			return true
		}
	}
	return false
}
