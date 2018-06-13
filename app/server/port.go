package server

import "time"
import "net"

type basePort struct {
	req *httpRequest

	name       string
	jun        uint16
	vlan       uint16
	mac        net.HardwareAddr
	updated_at *time.Time
}

func newPort(r *httpRequest) *basePort {
	return &basePort{}
}

func (m *basePort) newError(e uint8) *apiError {
	return m.req.newError(e)
}

func (m *basePort) parseMacAddress(mac *string) bool {

	var e error
	m.mac, e = net.ParseMAC(*mac)
	if e != nil {
		m.newError(errHostsAbnormalMac).log(e, "[PORT]: Could not parse the given MAC address!")
		return false
	}

	return true
}

func (m *basePort) rsviewAttributesParse() {

	//
}
