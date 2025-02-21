package place

import "github.com/hadroncorp/service-template/domain/valueobject"

type Place struct {
	id       string
	name     string
	location *valueobject.SpatialLocation
}

func (p Place) ID() string {
	return p.id
}

func (p Place) Name() string {
	return p.name
}

func (p Place) Location() *valueobject.SpatialLocation {
	return p.location
}
