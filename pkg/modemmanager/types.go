package modemmanager

import (
	"strconv"
)

// Boolean represents a boolean value with custom JSON marshalling.
type Boolean bool

// MarshalJSON marshals the boolean value.
func (b *Boolean) MarshalJSON() ([]byte, error) {
	if *b {
		return []byte("true"), nil
	}

	return []byte("false"), nil
}

// UnmarshalJSON unmarshals the boolean value.
func (b *Boolean) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "yes":
		*b = true
	case "no":
		*b = false
	}

	return nil
}

// Int represents an integer value with custom JSON marshalling.
type Int int

// MarshalJSON marshals the integer value.
func (i *Int) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(*i))), nil
}

// UnmarshalJSON unmarshals the integer value.
func (i *Int) UnmarshalJSON(data []byte) error {
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}

	*i = Int(value)

	return nil
}
