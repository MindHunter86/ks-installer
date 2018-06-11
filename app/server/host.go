package server

import "bytes"
import "time"
import "net"
import "net/http"
import "github.com/gorilla/context"
import "github.com/satori/go.uuid"


type (
	hostModel struct {
		id string
		request *httpRequest
	}
	baseHost struct {
		req *httpRequest
		model *hostModel

		id string
		hostname string
		ipmi_address *net.IP
		updated_at *time.Time
	}
	basePort struct {
		name string
		jun uint16
		vlan uint16
		mac string
		updated_at *time.Time
	}
)

func newHostModel(r *http.Request) *hostModel {
	return &hostModel{
		request: context.Get(r, "internal_request").(*httpRequest) }
}

func (m *hostModel) handleError(e error, err uint8, msg string) {
	m.request.newError(err).log(e, msg)
}

func (m *hostModel) checkMacExistence(mac string) bool {
	stmt,e := globSqlDB.Prepare("SELECT 1 FROM ports WHERE mac = ? LIMIT 2")
	if e != nil { m.handleError(e, errInternalSqlError, "[HOST]: Could not prepare DB statement!"); return false }
	defer stmt.Close()

	rows,e := stmt.Query(mac)
	if e != nil { m.handleError(e, errInternalSqlError, "[HOST]: Could not get result from DB!"); return false }
	defer rows.Close()

	if ! rows.Next() {
		m.handleError(rows.Err(), errInternalSqlError, "[HOST]: Could not exec rows.Next method!"); return false}

	if rows.Next() {
		m.handleError(nil, errInternalSqlError, "[HOST]: Rows is not equal to 1. The DB has broken!"); return false }

	return true
}

func (m *hostModel) getHostByMac(mac string) *baseHost {
	stmt,e := globSqlDB.Prepare(`SELECT hosts.id,hosts.hostname,hosts.ipmi_address,hosts.updated_at
															FROM hosts
															INNER JOIN ports
															ON hosts.id=ports.host
															WHERE ports.mac=? LIMIT 2`)
	if e != nil { m.handleError(e, errInternalSqlError, "[HOST]: Could not prepare DB statement!"); return nil }
	defer stmt.Close()

	rows,e := stmt.Query(mac)
	if e != nil { m.handleError(e, errInternalSqlError, "[HOST]: Could not get result from DB!"); return nil }
	defer rows.Close()

	if ! rows.Next() {
		m.handleError(rows.Err(), errInternalSqlError, "[HOST]: Could not exec rows.Next method!"); return nil}

	var host = new(baseHost)
	if e = rows.Scan(&host.id, &host.hostname, &host.ipmi_address, &host.updated_at); e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not scan the result from DB!"); return nil }

	if rows.Next() {
		m.handleError(nil, errInternalSqlError, "[HOST]: Rows is not equal to 1. The DB has broken!"); return nil }

	return host
}

func (m *hostModel) checkHostExistence(hostname string) bool {

	stmt,e := globSqlDB.Prepare("SELECT 1 FROM ports WHERE hostname = ? LIMIT 2")
	if e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not prepare DB statement!")
		return false }
	defer stmt.Close()

	rows,e := stmt.Query(hostname)
	if e != nil {
		m.handleError(e, errInternalSqlError, "[HOST]: Could not get result from DB!")
		return false }
	defer rows.Close()

	if ! rows.Next() {
		m.handleError(rows.Err(), errInternalSqlError, "[HOST]: Could not exec rows.Next method!")
		return false }

	if rows.Next() {
		m.handleError(nil, errInternalSqlError, "[HOST]: Rows is not equal to 1. The DB has broken!")
		return false }

	return true
}

func (m *hostModel) createNewHost(ipmiIp *string) *baseHost {

	var host = newHost(m.request)

	if host.ipmi_address = host.parseIpmiAddress(ipmiIp); host.ipmi_address == nil {
		return nil }

	var buf bytes.Buffer
	if _,e := buf.WriteString(host.resolveIpmiHostname()); e != nil {
		m.handleError(e, errInternalCommonError, "[HOST]: Bytes buffer - could not write the given string!")
		return nil }

	if buf.Len() == 0 { return nil }

	var bufBytes []byte
	bufBytes,e := buf.ReadBytes(byte('.')); if e != nil {
		m.handleError(e,errInternalCommonError, "[HOST] Bytes buffer - could not read from buffer!")
		return nil }

	if ! bytes.Equal([]byte(globConfig.Base.Ipmi.Hostname_Tld), buf.Bytes()) {
		m.handleError(nil, errHostsIpmiTldMismatch, "[HOST]: Top-level domain of the resolved IPMI hostname does not match the configuration!")
		return nil }

	host.hostname = string(bufBytes[:len(bufBytes)-1])
	host.id = host.genId()
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

func newHost(r *httpRequest) *baseHost {
	return &baseHost{
		req: r,
	}
}

func (m *baseHost) handleError(e error, err uint8, msg string) {
	m.req.newError(err).log(e, msg)
}

func (m *baseHost) genId() string {
	if m.id == "" {
		m.id = uuid.NewV4().String() }
	return m.id
}

func (m *baseHost) parseIpmiAddress(ipmiIp *string) *net.IP {

	var ipmiAddr = net.ParseIP(*ipmiIp)
	if ipmiAddr == nil {
		m.handleError(nil, errHostsAbnormalIp, "[HOST]: Could not parse the given IP address!")
		return nil }

	_,ipmiBlock,e := net.ParseCIDR(globConfig.Base.Ipmi.CIDR_Block); if e != nil {
		m.handleError(e, errInternalCommonError, "[HOST]: Could not parse configured ipmi CIDR block!")
		return nil }

	if ! ipmiBlock.Contains(ipmiAddr) {
		m.handleError(nil, errHostsIpmiCidrMismatch, "[HOST] The configured CIDR does not include this IP address!")
		return nil }

	return &ipmiAddr
}

func (m *baseHost) resolveIpmiHostname() string {

	hostnames,e := net.LookupAddr(m.ipmi_address.String()); if e != nil {
		m.handleError(e, errInternalCommonError, "[HOST]: Net lookup error!"); return "" }

	if len(hostnames) != 1 {
		m.handleError(nil, errHostsAmbiguousResolver, "[HOST]: The resolver returned two or more hostnames!"); return "" }

	return hostnames[1]
}

/*
func (m *baseUser) createAndSave(phone string, chatId int64) error {
	stmt,e := globSqlDB.Prepare("INSERT INTO users (phone,chat_id,registered) VALUES (?,?,?)")
	if e != nil { return e }
	defer stmt.Close()

	if _,e := stmt.Exec(phone, chatId, true); e != nil { return e }

	m.phone = phone
	m.chatId = chatId

	return nil
}
*/
