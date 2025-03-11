package main

import "math"

type Position struct {
	X float64
	Y float64
}

func getPositionOnCircle(i, n int, r float64) Position {
	angle := float64(i) * 2.0 * math.Pi / float64(n)

	return Position{
		X: r * math.Cos(angle),
		Y: r * math.Sin(angle),
	}
}
