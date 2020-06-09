//
package connection

import "fmt"

// Keys
type Key string 		// (flow-key)

// Identifier generators
func (conn *Connection) Key() Key {
	return Key(fmt.Sprint("(", conn.flow.Key(), ")"))
}