package assert

import (
	"fmt"
	"reflect"
)

func True(condition bool) {
	if !condition {
		panic("assertion error: expected true, got false")
	}
}

func False(condition bool) {
	if condition {
		panic("assertion error: expected false, got true")
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

func Nil(v any) {
	if !reflect.ValueOf(v).IsNil() {
		panic(fmt.Sprintf("assertion error (v != nil): %#v", v))
	}
}

func NotNil(v any) {
	if reflect.ValueOf(v).IsNil() {
		panic(fmt.Sprintf("assertion error (v == nil): %#v", v))
	}
}

func NoError(err error) {
	if err != nil {
		panic(fmt.Sprintf("assertion error (err != nil): %#v", err))
	}
}