package errutil

import (
	"errors"
	"fmt"
	"testing"
)

func ExampleAppend() {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err := Append(err1, err2)

	fmt.Println(err)
	fmt.Println(errors.Is(err, err1))
	fmt.Println(errors.Is(err, err2))

	// Output:
	// error 1; error 2
	// true
	// true
}

func TestAppend(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if err := Append(); err != nil {
			t.Error("expected nil, got", err)
		}
	})

	t.Run("nil", func(t *testing.T) {
		if err := Append(nil); err != nil {
			t.Error("expected nil, got", err)
		}
	})

	t.Run("multiple nil", func(t *testing.T) {
		if err := Append(nil, nil, nil); err != nil {
			t.Error("expected nil, got", err)
		}
	})

	t.Run("single", func(t *testing.T) {
		err := Append(errors.New("test"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "test" {
			t.Error("expected test, got", err.Error())
		}
	})

	t.Run("multiple", func(t *testing.T) {
		err := Append(errors.New("test1"), errors.New("test2"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "test1; test2" {
			t.Error("expected test1; test2, got", err.Error())
		}
	})

	t.Run("multiple with nil", func(t *testing.T) {
		err := Append(errors.New("test1"), nil, errors.New("test2"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "test1; test2" {
			t.Error("expected test1; test2, got", err.Error())
		}
	})

	t.Run("nested", func(t *testing.T) {
		err := Append(Append(errors.New("test1"), errors.New("test2")), errors.New("test3"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "test1; test2; test3" {
			t.Error("expected test1; test2; test3, got", err.Error())
		}
	})

	t.Run("errors.Is", func(t *testing.T) {
		err1 := errors.New("test1")
		err2 := errors.New("test2")
		err := Append(err1, err2)
		if !errors.Is(err, err1) {
			t.Error("expected true, got false")
		}
		if !errors.Is(err, err2) {
			t.Error("expected true, got false")
		}
	})

	t.Run("delim", func(t *testing.T) {
		SetDelim(" | ")
		err := Append(errors.New("test1"), errors.New("test2"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "test1 | test2" {
			t.Error("expected test1 | test2, got", err.Error())
		}
	})
}
