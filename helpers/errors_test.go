package helpers

import (
	"errors"
	"fmt"
)

func ExampleErrors_Error() {
	errs := Errors{}
	errs["first"] = errors.New("first error")
	errs["second"] = nil
	errs["third"] = errors.New("third error")

	fmt.Println(errs.Filter().Error())
	// Example output: first: first error; third: third error
}
