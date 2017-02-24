package dom

import (
	"fmt"
)

// Value wraps the value of a Node so we have types here
type Value struct {
	value interface{}
}

// String returns the Value as string
func (v *Value) String() string {
	if v.value == nil {
		return ""
	}
	return fmt.Sprint(v.value)
}

// IsNil returns true if the stored value is nil
func (v *Value) IsNil() bool {
	return v.value == nil
}

// SetValue sets the value of this Node
func (v *Value) SetValue(val interface{}) {
	v.value = val
}

// CopyValue copies the source value to the given target Node
func (v *Value) CopyValue(target *Node) {
	target.Value.SetValue(v.value)
}
