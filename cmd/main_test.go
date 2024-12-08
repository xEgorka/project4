package main

import (
	"testing"
)

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func Test_main(t *testing.T) {
	tests := []struct{ name string }{{name: "negative test #1"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertPanic(t, main) })
	}
}
