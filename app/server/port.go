package server

import "time"


type basePort struct {
	name string
	jun uint16
	vlan uint16
	mac string
	updated_at *time.Time
}


func newPort(r *httpRequest) *basePort {
	return &basePort{}
}
