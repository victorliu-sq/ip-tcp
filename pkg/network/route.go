package network

type Route struct {
	Dest string
	Next string
	Cost uint32
}

func NewRoute(dest, next string, cost uint32) Route {
	route := Route{
		Dest: dest,
		Next: next,
		Cost: cost,
	}
	return route
}
