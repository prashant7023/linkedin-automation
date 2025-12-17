package stealth

import (
	"math"
	"math/rand"
	"time"
)

// Point represents a 2D coordinate
type Point struct {
	X, Y float64
}

// GenerateBezierPath generates a smooth Bézier curve path from start to end
// with natural overshoot and micro-corrections
func GenerateBezierPath(start, end Point, numPoints int) []Point {
	// Add random control points for natural curve
	controlPoint1 := Point{
		X: start.X + (end.X-start.X)*0.3 + float64(rand.Intn(50)-25),
		Y: start.Y + (end.Y-start.Y)*0.3 + float64(rand.Intn(50)-25),
	}
	controlPoint2 := Point{
		X: start.X + (end.X-start.X)*0.7 + float64(rand.Intn(50)-25),
		Y: start.Y + (end.Y-start.Y)*0.7 + float64(rand.Intn(50)-25),
	}

	// Generate curve points
	path := make([]Point, 0, numPoints)
	for i := 0; i <= numPoints; i++ {
		t := float64(i) / float64(numPoints)
		point := cubicBezier(start, controlPoint1, controlPoint2, end, t)
		path = append(path, point)
	}

	// Add overshoot at the end (human tendency)
	if rand.Float64() < 0.7 { // 70% chance of overshoot
		overshoot := Point{
			X: end.X + float64(rand.Intn(10)-5),
			Y: end.Y + float64(rand.Intn(10)-5),
		}
		path = append(path, overshoot)
		
		// Add correction back to target
		path = append(path, end)
	}

	return path
}

// cubicBezier calculates a point on a cubic Bézier curve
func cubicBezier(p0, p1, p2, p3 Point, t float64) Point {
	// B(t) = (1-t)³P₀ + 3(1-t)²tP₁ + 3(1-t)t²P₂ + t³P₃
	oneMinusT := 1 - t
	oneMinusT2 := oneMinusT * oneMinusT
	oneMinusT3 := oneMinusT2 * oneMinusT
	t2 := t * t
	t3 := t2 * t

	return Point{
		X: oneMinusT3*p0.X + 3*oneMinusT2*t*p1.X + 3*oneMinusT*t2*p2.X + t3*p3.X,
		Y: oneMinusT3*p0.Y + 3*oneMinusT2*t*p1.Y + 3*oneMinusT*t2*p2.Y + t3*p3.Y,
	}
}

// GetVariableSpeed returns a random delay to vary mouse movement speed
func GetVariableSpeed() time.Duration {
	// Variable speed between 1-5ms per point
	return time.Duration(rand.Intn(4)+1) * time.Millisecond
}

// AddMicroCorrections adds small random adjustments to simulate human imprecision
func AddMicroCorrections(path []Point) []Point {
	corrected := make([]Point, len(path))
	for i, p := range path {
		jitter := 1.0
		if rand.Float64() < 0.3 { // 30% chance of minor correction
			jitter = float64(rand.Intn(3) - 1) // -1, 0, or 1 pixel
		}
		corrected[i] = Point{
			X: p.X + jitter,
			Y: p.Y + jitter,
		}
	}
	return corrected
}

// CalculateDistance calculates Euclidean distance between two points
func CalculateDistance(p1, p2 Point) float64 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return math.Sqrt(dx*dx + dy*dy)
}
