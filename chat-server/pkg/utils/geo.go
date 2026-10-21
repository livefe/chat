package utils

import (
	"encoding/json"
	"math"

	"gorm.io/datatypes"
)

const earthRadiusMeters = 6371000.0

func Haversine(lat1, lng1, lat2, lng2 float64) float64 {
	toRad := math.Pi / 180
	dLat := (lat2 - lat1) * toRad
	dLng := (lng2 - lng1) * toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusMeters * c
}

func InCircle(lat, lng, centerLat, centerLng, radiusMeters float64) bool {
	if radiusMeters <= 0 {
		return false
	}
	return Haversine(lat, lng, centerLat, centerLng) <= radiusMeters
}

func Centroid(points [][2]float64) (lat, lng float64) {
	if len(points) == 0 {
		return 0, 0
	}
	var sumLat, sumLng float64
	for _, p := range points {
		sumLng += p[0]
		sumLat += p[1]
	}
	n := float64(len(points))
	return sumLat / n, sumLng / n
}

func BoundingRadius(centerLat, centerLng float64, points [][2]float64) uint {
	var maxMeters float64
	for _, p := range points {
		d := Haversine(centerLat, centerLng, p[1], p[0])
		if d > maxMeters {
			maxMeters = d
		}
	}
	return uint(math.Ceil(maxMeters))
}

func InPolygon(lat, lng float64, points [][2]float64) bool {
	if len(points) < 3 {
		return false
	}
	inside := false
	j := len(points) - 1
	for i := 0; i < len(points); i++ {
		xi, yi := points[i][0], points[i][1]
		xj, yj := points[j][0], points[j][1]
		intersect := ((yi > lat) != (yj > lat)) &&
			(lng < (xj-xi)*(lat-yi)/(yj-yi+1e-12)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

func DecodePolygon(raw datatypes.JSON) ([][2]float64, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var pts [][2]float64
	if err := json.Unmarshal(raw, &pts); err != nil {
		return nil, err
	}
	return pts, nil
}

func EncodePolygon(pts [][2]float64) (datatypes.JSON, error) {
	if len(pts) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(pts)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}
