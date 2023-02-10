package utils

import (
	"testing"
)

func TestIsSubSlice(t *testing.T) {
	parent := []any{"a", "b", "c", "d", "e"}
	sub := []any{"a", "c", "e"}
	str := IsSubSlice(parent, sub, func(x, y any) bool {
		return x == y
	})
	if !str {
		t.Error("验证失败")
	}

	parentI := []any{1, 2, 3, 4, 5}
	subI := []any{1, 2, 3}
	ret := IsSubSlice(parentI, subI, func(x, y any) bool {
		return x == y
	})
	if !ret {
		t.Error("验证失败")
	}

}
