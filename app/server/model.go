package server

import "bytes"
import "net/http"
import "github.com/gorilla/context"

type hostModel struct {
	id      string
	request *httpRequest
}

func newHostModel(r *http.Request) *hostModel {
	return &hostModel{
		request: context.Get(r, "internal_request").(*httpRequest)}
}

func (m *hostModel) handleError(e error, err uint8, msg string) {
	m.request.newError(err).log(e, msg)
}

func (m *hostModel) checkMacExistence(mac string) bool {
	stmt, e := globSqlDB.Prepare("SELECT 1 FROM ports WHERE mac = ? LIMIT 2")
	if e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not prepare DB statement!")
		return false
	}
	defer stmt.Close()

	rows, e := stmt.Query(mac)
	if e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not get result from DB!")
		return false
	}
	defer rows.Close()

	// BUG: bad if for rows.Next (check queue.go for fix it)
	if !rows.Next() {
		m.handleError(rows.Err(), errInternalSqlError, "[HOST]: Could not exec rows.Next method!")
		return false
	}

	if rows.Next() {
		m.handleError(nil, errInternalSqlError, "[HOST]: Rows is not equal to 1. The DB has broken!")
		return false
	}

	return true
}

func (m *hostModel) getHostByMac(mac string) *baseHost {
	stmt, e := globSqlDB.Prepare(`SELECT hosts.id,hosts.hostname,hosts.ipmi_address,hosts.updated_at
															FROM hosts
															INNER JOIN ports
															ON hosts.id=ports.host
															WHERE ports.mac=? LIMIT 2`)
	if e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not prepare DB statement!")
		return nil
	}
	defer stmt.Close()

	rows, e := stmt.Query(mac)
	if e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not get result from DB!")
		return nil
	}
	defer rows.Close()

	// BUG: bad if for rows.Next (check queue.go for fix it)
	if !rows.Next() {
		m.handleError(rows.Err(), errInternalSqlError, "[HOST]: Could not exec rows.Next method!")
		return nil
	}

	var host = new(baseHost)
	if e = rows.Scan(&host.id, &host.hostname, &host.ipmi_address, &host.updated_at); e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not scan the result from DB!")
		return nil
	}

	if rows.Next() {
		m.handleError(nil, errInternalSqlError, "[HOST]: Rows is not equal to 1. The DB has broken!")
		return nil
	}

	return host
}

func (m *hostModel) checkHostExistence(hostname string) bool {

	stmt, e := globSqlDB.Prepare("SELECT 1 FROM ports WHERE hostname = ? LIMIT 2")
	if e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not prepare DB statement!")
		return false
	}
	defer stmt.Close()

	rows, e := stmt.Query(hostname)
	if e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not get result from DB!")
		return false
	}
	defer rows.Close()

	// BUG: bad if for rows.Next (check queue.go for fix it)
	if !rows.Next() {
		m.handleError(rows.Err(), errInternalSqlError, "[HOST]: Could not exec rows.Next method!")
		return false
	}

	if rows.Next() {
		m.handleError(nil, errInternalSqlError, "[HOST]: Rows is not equal to 1. The DB has broken!")
		return false
	}

	return true
}

// TODO: REMOVE THIS SHIT
func (m *hostModel) createNewHost(ipmiIp *string) *baseHost {

	var host = newHost()

	if err := host.parseIpmiAddress(ipmiIp); err != nil {
		return nil
	}

	var buf bytes.Buffer
	//if _, e := buf.WriteString(host.resolveIpmiHostname()); e != nil {
		//m.handleError(e, errInternalCommonError, "[HOST]: Bytes buffer - could not write the given string!")
		//return nil
	//}

	if buf.Len() == 0 {
		return nil
	}

	var bufBytes []byte
	bufBytes, e := buf.ReadBytes(byte('.'))
	if e != nil {
		m.handleError(e, errInternalCommonError, "[HOST] Bytes buffer - could not read from buffer!")
		return nil
	}

	if !bytes.Equal([]byte(globConfig.Base.Ipmi.Hostname_Tld), buf.Bytes()) {
		m.handleError(nil, errHostsIpmiTldMismatch, "[HOST]: Top-level domain of the resolved IPMI hostname does not match the configuration!")
		return nil
	}

	host.hostname = string(bufBytes[:len(bufBytes)-1])
	globLogger.Info().Str("id", host.id).Str("hostname", host.hostname).Msg("[HOST]: New host has been created!")

	// TODO: Host save

	/*


		save or update host in DB
		if host is exist in DB check it's UUID in ports table
		...
		some logic
	*/

	return host
}

/*
	stmt,e := globSqlDB.Prepare("INSERT INTO hosts (id,hostname,ipmi_address) VALUES (?,?,?)")
	if e != nil {
		req.newError(errInternalSqlError).log(e, "[HOST]: Could not prepare DB statement!"); return nil }
	defer stmt.Close()

	if _,e = stmt.Exec(nil,nil,nil) {}
*/
