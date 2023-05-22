package utils

type AllowCompare interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~string | ~float32 | ~float64
}

func IsSubSlice[T AllowCompare](parent, sub []T) bool {
	dst := make(map[T]bool, len(parent))
	for _, v := range parent {
		dst[v] = true
	}

	for _, t := range sub {
		if v := dst[t]; !v {
			return false
		}
	}

	return true
}

func InSlice[T AllowCompare](src []T, dst T) bool {
	for _, t := range src {
		if t == dst {
			return true
		}
	}
	return false
}

func RemoveDuplicate[T AllowCompare](src []T) (dst []T) {
	mp := make(map[T]bool, len(src))
	for _, v := range src {
		if _, b := mp[v]; !b {
			dst = append(dst, v)
			mp[v] = true
		}
	}
	return
}
