package valueobject

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hadroncorp/geck/errors/syserr"
	"github.com/samber/lo"
)

type SpatialLocation struct {
	Latitude  float64
	Longitude float64
}

// compile-time assertion(s)
var _ fmt.Stringer = (*SpatialLocation)(nil)

func NewSpatialLocation(lat, lon float64) SpatialLocation {
	return SpatialLocation{
		Latitude:  lat,
		Longitude: lon,
	}
}

func ParseSpatialLocation(fieldName, v string) (SpatialLocation, error) {
	splitVal := strings.SplitN(v, ",", 2)
	if len(splitVal) != 2 {
		fieldName = lo.CoalesceOrEmpty(fieldName, "spatial_location")
		return SpatialLocation{}, syserr.NewInvalidFormat(fieldName, "coordinates (latitude, logitude)")
	}

	lat, err := strconv.ParseFloat(splitVal[0], 64)
	if err != nil {
		return SpatialLocation{}, err
	}
	lon, err := strconv.ParseFloat(splitVal[1], 64)
	if err != nil {
		return SpatialLocation{}, err
	}
	return SpatialLocation{
		Latitude:  lat,
		Longitude: lon,
	}, nil
}

func (s SpatialLocation) String() string {
	return fmt.Sprintf("%f,%f", s.Latitude, s.Longitude)
}
