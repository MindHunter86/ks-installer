package server

import "time"
import "net"
import "github.com/satori/go.uuid"


type (
	baseHost struct {
		req *httpRequest
		model *hostModel

		id string
		hostname string
		ipmi_address *net.IP
		updated_at *time.Time
	}
)


func newHost(r *httpRequest) *baseHost {
	return &baseHost{
		id: uuid.NewV4().String(),
		req: r,
	}
}

func (m *baseHost) handleError(e error, err uint8, msg string) {
	m.req.newError(err).log(e, msg)
}

func (m *baseHost) parseIpmiAddress(ipmiIp *string) bool {

	var ipmiAddr = net.ParseIP(*ipmiIp)
	if ipmiAddr == nil {
		m.handleError(nil, errHostsAbnormalIp, "[HOST]: Could not parse the given IP address!")
		return false }

	_,ipmiBlock,e := net.ParseCIDR(globConfig.Base.Ipmi.CIDR_Block); if e != nil {
		m.handleError(e, errInternalCommonError, "[HOST]: Could not parse configured ipmi CIDR block!")
		return false }

	if ! ipmiBlock.Contains(ipmiAddr) {
		m.handleError(nil, errHostsIpmiCidrMismatch, "[HOST] The configured CIDR does not include this IP address!")
		return false }

	m.ipmi_address = &ipmiAddr
	return true
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
