package badcli

import "errors"
import "strconv"

// Assert interface compliance.
var _ FlagValue = (*BoundedInt)(nil)

type BoundedInt struct {
	value int
	min int
	max int
}

func NewBoundedInt(value, min, max int) *BoundedInt {
	return &BoundedInt{ value: value, min: min, max: max }
}

func (self BoundedInt) Value() int {
	return self.value
}

func (self BoundedInt) String() string {
	return strconv.Itoa(int(self.value))
}

func (self *BoundedInt) ParseFromArg(arg string) error {
	argInt64, err := strconv.ParseInt(arg, 10, strconv.IntSize)
	if err != nil { return err }

	argInt := int(argInt64)
	if argInt < self.min {
		return errors.New("minimum value is '" + strconv.Itoa(self.min) + "', but got '" + arg + "' instead")
	}
	if argInt > self.max {
		return errors.New("maximum value is '" + strconv.Itoa(self.max) + "', but got '" + arg + "' instead")
	}
	self.value = argInt
	
	return nil
}
