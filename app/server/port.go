package server

import "time"
import "net"

type basePort struct {
	name       string
	jun        uint16
	vlan       uint16
	mac        []*net.HardwareAddr
	updated_at *time.Time
}

func newPort() *basePort {
	return &basePort{}
}

func (m *basePort) parseMacAddress(mac []*string) *appError {

	for _,v := range mac {
		hwAddr,e := net.ParseMAC(*v); if e != nil {
			ae := newAppError(errPortsAbnormalMac)
			return ae.log(e, "Could not parse the given MAC address!", ae.glCtx().Str("mac", *v))
		}

		m.mac = append(m.mac, &hwAddr)
	}

	return nil
}

func (m *basePort) importAllFields() *appError {

	// parse rsview by mac
	return nil
}

func (m *basePort) rsviewAttributesParse() *appError {

	// var rsResult []*string

	//for _,v := range m.mac {

		//rsResult,e := globRsview.parsePortAttributes(v); if e != nil {
		//	return e
		//}

		// SOME MAGIC; SOME LOGIC

		// todo/plan:
		// - check vlan
		// - check jun
		// - check zone name

		// - get req id, get all jobs for this req id
		// - check all jobs; if spme jobs are not read; wait them and check status to BLOCKED

		// - if all jobs are OK, check all created host by reqid
		// - compare hostname from rsview and from found job
		// - if comparison failed - error

		// - if hostname comparison is OK, check existed MAC for host.
		// - if MAC is not NULL and is not equal to our, check existed MAC in rsview
		// - if old MAC was found in rsview, then NONE - (to hard, fuck this)

		// - if all comparisons are OK;

		// - do we need a table with hard-coded relationships?

//	}
	// if parseResult,ec := globRsview.parsePortAttributes

	return nil
}
