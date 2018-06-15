package server

import "strings"
import "net/http"
import _ "github.com/go-sql-driver/mysql"
import "github.com/satori/go.uuid"

type httpRequest struct {
	id, link string
	status   int
	errors   []*appError
}

func (m *httpRequest) createAndSave(req *http.Request) (*httpRequest, error) {
	m.id = uuid.NewV4().String()
	m.link = req.RequestURI

	stmt, e := globSqlDB.Prepare("INSERT INTO requests (id,srcip,method,size,url,status,user_agent) VALUES (?,?,?,?,?,?,?)")
	if e != nil {
		return m, e
	}
	defer stmt.Close()

	if _, e = stmt.Exec(m.id, strings.Split(req.RemoteAddr, ":")[0], req.Method, req.ContentLength, m.link, m.status, req.UserAgent()); e != nil {
		return m, e
	}

	m.status = http.StatusOK
	return m, e
}

func (m *httpRequest) updateAndSave() {
	stmt, e := globSqlDB.Prepare("UPDATE requests SET status = ? WHERE id = ?")
	if e != nil {
		globLogger.Error().Err(e).Msg("[REQUEST]: Could not prepare DB statement!")
		return
	}
	defer stmt.Close()

	if _, e := stmt.Exec(m.status, m.id); e != nil {
		globLogger.Error().Err(e).Msg("[REQUEST]: Could not execute DB statement!")
		return
	}
}

func (m *httpRequest) appendAppError(aErr *appError) *appError {
	m.errors = append(m.errors, aErr.setRequestId(m.id))
	return aErr
}

// TODO: 2DELETE !!!
func (m *httpRequest) newError(e uint8) (err *appError) {
	err = newAppError(e)
	m.errors = append(m.errors, err.setRequestId(m.id))
	return err
}

func (m *httpRequest) respondApiErrors() []*responseError {

	var rspErrors []*responseError

	for _, v := range m.errors {
		rspErrors = append(rspErrors, &responseError{
			Id:     v.id,
			Code:   int(v.code),
			Status: v.getHttpStatusCode(),
			Title:  v.getErrorTitle(),
			Detail: v.getHumanDetails()})
		//	Source: &errorSource{
		//		Parameter: v.srcParam}})

		if v.getHttpStatusCode() > m.status {
			m.status = v.getHttpStatusCode()
		}
	}

	return rspErrors
}

func (m *httpRequest) saveErrors() *httpRequest {

	for _, v := range m.errors {
		if v.save() {
			globLogger.Debug().Str("request_link", m.link).Int("http_code", v.getHttpStatusCode()).Str("error_id", v.id).Str("error_title", v.getErrorTitle()).Msg("|SAVED|")
			continue
		}

		globLogger.Debug().Str("request_link", m.link).Int("http_code", v.getHttpStatusCode()).Str("error_id", v.id).Str("error_title", v.getErrorTitle()).Msg("|!NOT SAVED!|")
		m.appendAppError(newAppError(errInternalSqlError))
		break
	}

	return m
}
