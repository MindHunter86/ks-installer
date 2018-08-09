package server

import "net"
import "strings"
import "strconv"

type netPort struct {
	mac          string `bolt:"primary_key"`
	attachedHost string `bolt:"foreign_key"`
	lldpHostname string
	junHostname  string
	junPortName  string
	junVlan      string
}

type basePort struct {
	mac           net.HardwareAddr
	jun_name      string
	jun_port_name string
	jun_vlan      uint16
	lldp_host     string
}

func newPort() *basePort {
	return &basePort{}
}

func newPortWithMAC(mac *string) (*basePort, *appError) {

	var e error
	var port *basePort = newPort()

	if port.mac, e = net.ParseMAC(*mac); e != nil {
		err := newAppError(errPortsAbnormalMac)
		return nil, err.log(e, "Could not parse the given MAC address!", err.glCtx().Str("mac", *mac))
	}

	return port, port.getOrCreate()
}

func (m *basePort) getOrCreate() *appError {

	rws, e := globSqlDB.Query("SELECT 1 FROM macs WHERE mac = ? LIMIT 2", m.mac.String())
	if e != nil {
		return newAppError(errInternalSqlError).log(e, "Could not get result from DB!")
	}
	defer rws.Close()

	if !rws.Next() {
		if rws.Err() != nil {
			return newAppError(errInternalSqlError).log(e, "Could not exec rows.Next method!")
		}

		if _, e = globSqlDB.Exec("INSERT INTO macs (mac) VALUES (?)", m.mac.String()); e != nil {
			return newAppError(errInternalSqlError).log(e, "Could not save the mac into DB!")
		}

		return nil
	}

	if rws.Next() {
		return newAppError(errInternalSqlError).log(nil, "Rows is not equal to 1. The DB has broken!")
	}

	return nil
}

func (m *basePort) parseRsviewProperties() *appError {

	rsResult, err := globRsview.getPortAttributes(m.mac)
	if err != nil {
		return err
	}

	// parse rsview VLANs:
	for _, v := range globConfig.Base.Rsview.AllowRules.Vlans {
		if strings.Contains(rsResult[rsviewTableVlans], v) {
			if m.jun_vlan == 0 {
				buf, e := strconv.ParseUint(v, 10, 16)
				if e == nil {
					m.jun_vlan = uint16(buf)
					continue
				}

				return newAppError(errInternalCommonError).log(e, "Could not convert string to uint16!")
			}

			globLogger.Warn().Msg("Something is wrong in (*basePort).getRsviewProperties(). Given VLANs have two or more configuration matches!")
		}
	}

	if m.jun_vlan == 0 {
		return newAppError(errRsviewUnknownVLAN).log(nil, "Given VLANs don't have configuration matches!")
	}

	// parse port name:
	for _, v := range globConfig.Base.Rsview.AllowRules.PortNames {
		if strings.Contains(rsResult[rsviewTablePort], v) {
			if m.jun_port_name == "" {
				m.jun_port_name = rsResult[rsviewTablePort]
				continue
			}

			globLogger.Warn().Msg("Something is wrong in (*basePort).getRsviewProperties(). Given Ports have two or more configuration matches!")
		}
	}

	if m.jun_port_name == "" {
		return newAppError(errRsviewUnknownPort).log(nil, "Given Port doesn't have configuration matches!")
	}

	// parse jun name:
	for _, v := range globConfig.Base.Rsview.AllowRules.JunNames {
		if strings.Contains(rsResult[rsviewTableHostname], v) {
			if m.jun_name == "" {
				m.jun_name = rsResult[rsviewTableHostname]
				continue
			}

			globLogger.Warn().Msg("Something is wrong in (*basePort).getRsviewProperties(). Given Jun has two or more configuration matches!")
		}
	}

	if m.jun_name == "" {
		return newAppError(errRsviewUnknownJun).log(nil, "Given Jun doesn't have configuration matches!")
	}

	// parse lldp host:
	if rsResult[rsviewTableLldp] != "" {
		if buf := strings.SplitN(rsResult[rsviewTableLldp], ".", 2); len(buf) != 0 {
			m.lldp_host = buf[0]
		}
	}

	if m.lldp_host == "" {
		return newAppError(errRsviewUnknownLLDP).log(nil, "Given LLDP host does not valid!")
	}

	return m.updateRsviewProperties()
}

func (m *basePort) updateRsviewProperties() *appError {

	_, e := globSqlDB.Exec(
		"UPDATE macs SET jun_name = ?, jun_port_name = ?, jun_vlan = ? WHERE mac = ?",
		m.jun_name, m.jun_port_name, m.jun_vlan, m.mac.String())
	if e != nil {
		return newAppError(errInternalSqlError).log(nil, "Could not exec the database query!")
	}

	return nil
}

func (m *basePort) compareLLDPWithHost(hostname string) bool {

	if m.lldp_host == hostname {
		return true
	}

	return false
}

func (m *basePort) linkWithHost(hId string) *appError {

	if _, e := globSqlDB.Exec("UPDATE macs SET host = ? WHERE mac = ?", hId, m.mac.String()); e != nil {
		return newAppError(errInternalSqlError).log(e, "Could not save the mac into DB!")
	}

	return nil
}
