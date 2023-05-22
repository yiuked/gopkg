package main

import (
	"fmt"
)

type Slice[T int | float32 | float64] []T
type SliceS[T ~int | ~float32 | ~float64] []T

func main() {
	price := Slice[float32]{-1.1, -1.2, 0.0}
	fmt.Println(price)

	priceS := SliceS[float32]{-1.1, -1.2, 0.0}
	fmt.Println(priceS)
}
