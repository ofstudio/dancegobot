package errutil

// DefaultDelim is the default delimiter used by Append.
const DefaultDelim = "; "

var delim = DefaultDelim

// SetDelim modifies the delimiter used by Append.
func SetDelim(d string) {
	delim = d
}

// Append returns an error that wraps the given errors.
// Any nil error values are discarded.
// Append returns nil if every value in errs is nil.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a delimiter
// between each string.
//
// A non-nil error returned by Append implements the Unwrap() []error method.
func Append(errs ...error) error {
	// Count non-nil errors.
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}

	// If there are no errors, return nil.
	if n == 0 {
		return nil
	}

	// Create a list of errors.
	l := make(list, 0, n)
	for _, err := range errs {
		if err != nil {
			l = append(l, err)
		}
	}

	return &l
}

// Unwrap returns the list of errors.
func (e list) Unwrap() []error {
	return e
}

type list []error

// Error implements the error interface.
func (e list) Error() string {
	// Len is always > 0
	if len(e) == 1 {
		return e[0].Error()
	}
	b := []byte(e[0].Error())
	for _, err := range e[1:] {
		b = append(b, delim...)
		b = append(b, err.Error()...)
	}
	return string(b)
}
