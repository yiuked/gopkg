// You can edit this code!
// Click here and start typing.
package main

import (
	"log"
	"math"
	"time"
)

func main() {
	log.Println(NaturalDayDiff(time.Now(), time.Date(2023, 5, 31, 23, 1, 1, 1, time.Local)))
}

// NaturalDayDiff 自然天数差
func NaturalDayDiff(from, to time.Time) int {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
	to = time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.Local)
	return int(math.Abs(float64(to.Sub(from) / (24 * time.Hour))))
}
