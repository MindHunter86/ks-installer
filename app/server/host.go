package server

import "strings"

import "time"
import "net"
import "github.com/satori/go.uuid"

type (
	baseHost struct {
		id           string
		hostname     string
		ipmi_address *net.IP
		created_by string
		updated_at time.Time
	}
)

func newHost() *baseHost {
	return &baseHost{
		id:  uuid.NewV4().String(),
	}
}

func (m *baseHost) parseIpmiAddress(ipmiIp *string) *appError {

	var ipmiAddr = net.ParseIP(*ipmiIp)
	if ipmiAddr == nil {
		return newAppError(errHostsAbnormalIp).log(nil, "Could not parse the given IP address!")
	}

	_, ipmiBlock, e := net.ParseCIDR(globConfig.Base.Ipmi.CIDR_Block)
	if e != nil {
		return newAppError(errInternalCommonError).log(nil, "Could not parse configured ipmi CIDR block!")
	}

	if !ipmiBlock.Contains(ipmiAddr) {
		return newAppError(errHostsIpmiCidrMismatch).log(nil, "The configured CIDR does not include this IP address!")
	}

	m.ipmi_address = &ipmiAddr
	return nil
}

func (m *baseHost) resolveIpmiHostname() *appError {

	hostnames, e := net.LookupAddr(m.ipmi_address.String())
	if e != nil {
		return newAppError(errInternalCommonError).log(e, "Net lookup error!")
	}

	if len(hostnames) != 1 {
		return newAppError(errHostsAmbiguousResolver).log(nil, "The resolver returned two or more hostnames!")
	}

	m.hostname = strings.SplitN(hostnames[0], ".", 2)[0]
	return nil
}

func (m *baseHost) updateOrCreate(jobId string) *appError {

	m.created_by = jobId

	ok,e := m.foundProperties(); if e != nil {
		return e
	}

	if ok {
		return m.updateProperties()
	}

	return m.createProperties()
}

func (m *baseHost) foundProperties() (bool, *appError) {

	rws,e := globSqlDB.Query("SELECT id,ipmi_address,updated_at,created_by WHERE hostname = ? LIMIT 2", m.hostname)
	if e != nil {
		return false,newAppError(errInternalSqlError).log(e, "Could not get result from DB!")
	}
	defer rws.Close()

	if ! rws.Next() {
		if rws.Err() != nil {
			return false,newAppError(errInternalSqlError).log(rws.Err(), "Could not exec rows.Next method!")
		}
		return false,nil
	}

	var oldJobId string
	var foundIpmiAddr string
	if e = rws.Scan(&oldJobId, &foundIpmiAddr, &m.updated_at); e != nil {
		return false,newAppError(errInternalSqlError).log(e, "Could not scan the result from DB!")
	}

	if m.ipmi_address.String() != foundIpmiAddr {
		globLogger.Warn().Str("hostname", m.hostname).Str("given_ipmi", m.ipmi_address.String()).Str("found_ipmi", foundIpmiAddr).Str("job_id", oldJobId).
			Time("last_update", m.updated_at).Msg("Found a conflict in the DB! The current App policy will overwrite the server!")
	}

	if err := m.parseIpmiAddress(&foundIpmiAddr); err != nil {
		return false,err
	}

	if rws.Next() {
		return false,newAppError(errInternalCommonError).log(nil, "Rows is not equal to 1. The DB has broken!")
	}

	return true,nil
}

func (m *baseHost) updateProperties() *appError {

	_,e := globSqlDB.Exec(
		"UPDATE hosts SET id = ?, ipmi_address = ?, created_by = ? WHERE hostname = ?",
		m.id, m.ipmi_address.String(), m.created_by, m.hostname)
	if e != nil {
		return newAppError(errInternalSqlError).log(e, "Could not exec the database query!")
	}

	return nil
}

func (m *baseHost) createProperties() *appError {

	_,e := globSqlDB.Exec(
		"INSERT INTO hosts (id, hostname, impi_address, created_by)",
		m.id, m.hostname, m.ipmi_address.String(), m.created_by)
	if e != nil {
		return newAppError(errInternalSqlError).log(e, "Could not exec the database query!")
	}

	return nil
}
