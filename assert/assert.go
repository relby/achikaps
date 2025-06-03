package assert

import "fmt"

func True(condition bool) {
	if !condition {
		panic("assertion error: expected true, got false")
	}
}

func Equals[T comparable](v1, v2 T) {
	if v1 != v2 {
		panic(fmt.Sprintf("assertion error (v1 != v2): v1(%#v) v2(%#v)", v1, v2))
	}
}

func NotEquals[T comparable](v1, v2 T) {
	if v1 == v2 {
		panic(fmt.Sprintf("assertion error (v1 == v2): v1(%#v) v2(%#v)", v1, v2))
	}
}

// TODO: this doesn't work as expected
// func Nil(v any) {
// 	if v != nil {
// 		panic(fmt.Sprintf("assertion error (v != nil): %#v", v))
// 	}
// }

// func NotNil(v any) {
// 	if v == nil {
// 		panic(fmt.Sprintf("assertion error (v == nil): %#v", v))
// 	}
// }

func NoError(err error) {
	if (err != nil) {
		panic(fmt.Sprintf("assertion error (err != nil): %#v", err))
	}
}

func Unreachable() {
	panic("assertion error (unreachable)")
}