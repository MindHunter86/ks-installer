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
	m.errors = append(m.errors, aErr)
	return aErr
}

// TODO: 2DELETE !!!
func (m *httpRequest) newError(e uint8) (err *appError) {
	err = newAppError(e)
	m.errors = append(m.errors, err)
	return err
}

// TODO REFACTOR
func (m *httpRequest) respondApiErrors() []*responseError {
	var rspErrors []*responseError

	for _, v := range m.errors {
		rspErrors = append(rspErrors, &responseError{
			Id:     v.id,
			Code:   int(v.code),
			Status: apiErrorsStatus[v.code],
			Title:  apiErrorsTitle[v.code],
			Detail: apiErrorsDetail[v.code]})
		//	Source: &errorSource{
		//		Parameter: v.srcParam}})

		if apiErrorsStatus[v.code] > m.status {
			m.status = apiErrorsStatus[v.code]
		}
	}

	return rspErrors
}

func (m *httpRequest) saveErrors() *httpRequest { // TODO REFACTOR THIS SHIT. Use (*appErr).save() method!
	stmt, e := globSqlDB.Prepare("INSERT INTO errors (id,request_id,internal_code,displayed_title,displayed_detail) VALUES (?,?,?,?,?)")
	if e != nil {
		globLogger.Error().Err(e).Msg("[REQUEST]: Could not prepare DB statement!")
		m.newError(errInternalSqlError)
		return m
	}
	defer stmt.Close()

	for _, v := range m.errors {
		globLogger.Info().Str("request_link", m.link).Int("http_code", apiErrorsStatus[v.code]).Str("error_id", v.id).Str("error_title", apiErrorsTitle[v.code]).Msg("[REQUEST]:")

		_, e = stmt.Exec(v.id, m.id, v.code, apiErrorsTitle[v.code], apiErrorsDetail[v.code])
		if e != nil {
			globLogger.Error().Err(e).Str("error_id", v.id).Msg("[REQUEST]: Could not write error report!")
		}
	}

	return m
}
