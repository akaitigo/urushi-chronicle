package main

import "testing"

func TestMain_noPanic(t *testing.T) {
	t.Run("main function does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("main() panicked: %v", r)
			}
		}()
		main()
	})
}
