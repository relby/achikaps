package vec2

import (
	"fmt"
	"math"
)

type Vec2 struct {
	X, Y float64
}

func Dot(ihs Vec2, rhs Vec2) float64 {
	return ihs.X*rhs.X + ihs.Y*rhs.Y
}

func Lerp(a Vec2, b Vec2, t float64) Vec2 {
	return New(
		a.X+(b.X-a.X)*t,
		a.Y+(b.Y-a.Y)*t,
	)
}

func Distance(a Vec2, b Vec2) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func Reflect(ihs Vec2, rhs Vec2) Vec2 {
	factor := -2.0 * Dot(ihs, rhs)
	return New(
		factor*ihs.X+rhs.X,
		factor*ihs.Y+rhs.Y,
	)
}

func New(x float64, y float64) Vec2 {
	return Vec2{x, y}
}

func (v *Vec2) Set(x float64, y float64) {
	v.X = x
	v.Y = y
}

func (v Vec2) Add(other Vec2) Vec2 {
	return New(v.X+other.X, v.Y+other.Y)
}

func (v Vec2) AddScalar(scalar float64) Vec2 {
	return New(v.X+scalar, v.Y+scalar)
}

func (v Vec2) AddScalars(x float64, y float64) Vec2 {
	return New(v.X+x, v.Y+y)
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return New(v.X-other.X, v.Y-other.Y)
}

func (v Vec2) SubScalar(scalar float64) Vec2 {
	return New(v.X-scalar, v.Y-scalar)
}

func (v Vec2) SubScalars(x float64, y float64) Vec2 {
	return New(v.X-x, v.Y-y)
}

func (v Vec2) Mul(other Vec2) Vec2 {
	return New(v.X*other.X, v.Y*other.Y)
}

func (v Vec2) MulScalar(scalar float64) Vec2 {
	return New(v.X*scalar, v.Y*scalar)
}

func (v Vec2) MulScalars(x float64, y float64) Vec2 {
	return New(v.X*x, v.Y*y)
}

func (v Vec2) Div(other Vec2) Vec2 {
	return New(v.X/other.X, v.Y/other.Y)
}

func (v Vec2) DivScalar(scalar float64) Vec2 {
	return New(v.X/scalar, v.Y/scalar)
}

func (v Vec2) DivScalars(x float64, y float64) Vec2 {
	return New(v.X/x, v.Y/y)
}

func (v Vec2) DistanceTo(other Vec2) float64 {
	return Distance(v, other)
}

func (v Vec2) Dot(other Vec2) float64 {
	return v.X*other.X + v.Y*other.Y
}

func (v Vec2) Lerp(other Vec2, t float64) Vec2 {
	return New(
		v.X+(other.X-v.X)*t,
		v.Y+(other.Y-v.Y)*t,
	)
}

func (v Vec2) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) Normalize() Vec2 {
	m := v.Magnitude()

	if m > 0.0 {
		return v.DivScalar(m)
	} else {
		return v
	}
}

func (v Vec2) Reflect(other Vec2) Vec2 {
	factor := -2.0 * v.Dot(other)
	return New(
		factor*v.X+other.X,
		factor*v.Y+other.Y,
	)
}

func (v Vec2) Equals(other Vec2) bool {
	return v.X == other.X && v.Y == other.Y
}

func (v Vec2) String() string {
	return fmt.Sprintf("Vec2(%f, %f)", v.X, v.Y)
}
