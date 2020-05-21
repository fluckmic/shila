package connection

import "fmt"

// Keys
type Key string 		// (flow-key,flow-kind,current-flow-state)

func (conn *Connection) Key() Key {
	return Key(fmt.Sprint("(", conn.flow.Key(), ",", conn.kind, ",", conn.state.current, ")"))
}