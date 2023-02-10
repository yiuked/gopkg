package utils

func IsSubSlice(parent, sub []any, f func(x, y any) bool) bool {
	for _, s := range sub {
		eq := false
		for _, p := range parent {
			if f(s, p) {
				eq = true
				break
			}
		}
		if !eq {
			return false
		}
	}
	return true
}

func IsSubSliceUint(parent, sub []uint) bool {
	for _, s := range sub {
		eq := false
		for _, p := range parent {
			if s == p {
				eq = true
				break
			}
		}
		if !eq {
			return false
		}
	}
	return true
}
