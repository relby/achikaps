package assert

import "fmt"

func True(condition bool) {
	if !condition {
		panic("assertion error: expected true, got false")
	}
}

func NoError(err error) {
	if (err != nil) {
		panic(fmt.Sprintf("assertion error (err != nil): %v", err))
	}
}