package json

type IPAddress struct {
	IP   string
	Port string
}

type Path struct {

}

type PathEntry struct {}

type NetFlow struct {
	Dst  IPAddress
	Path []PathEntry
}

type RoutingEntry struct {
	Key  IPAddress
	Flow NetFlow
}
