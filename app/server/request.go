package server

import "strings"
import "net/http"
import _ "github.com/go-sql-driver/mysql"
import "github.com/satori/go.uuid"

type httpRequest struct {
	id, link string
	status int
	errors []*apiError
}

func (m *httpRequest) createAndSave(req *http.Request) (*httpRequest, error) {
	m.id = uuid.NewV4().String()
	m.link = req.RequestURI

	stmt,e := globSqlDB.Prepare("INSERT INTO requests (id,srcip,method,size,url,status,user_agent) VALUES (?,?,?,?,?,?,?)"); if e != nil { return m,e }
	defer stmt.Close()

	if _,e = stmt.Exec(m.id, strings.Split(req.RemoteAddr, ":")[0], req.Method, req.ContentLength, m.link, m.status, req.UserAgent()); e != nil { return m,e }

	return m,e
}

func (m *httpRequest) updateAndSave() {
	stmt,e := globSqlDB.Prepare("UPDATE requests SET status = ? WHERE id = ?"); if e != nil {
		globLogger.Error().Err(e).Msg("[REQUEST]: Could not prepare DB statement!"); return }
	defer stmt.Close()

	if _,e := stmt.Exec(m.status, m.id); e != nil {
		globLogger.Error().Err(e).Msg("[REQUEST]: Could not execute DB statement!"); return }
}

func (m *httpRequest) newError(e uint8) *apiError {
	var err *apiError = new(apiError).setError(e)
	m.errors = append(m.errors, err)
	return err
}

func (m *httpRequest) respondApiErrors() ([]*responseError, int) {
	var rspErrors []*responseError

	for _,v := range m.errors {
		rspErrors = append(rspErrors, &responseError{
			Id: v.getId(),
			Code: int(v.e),
			Status: apiErrorsStatus[v.e],
			Title: apiErrorsTitle[v.e],
			Detail: apiErrorsDetail[v.e],
			Source: &errorSource{
				Parameter: v.srcParam },
			Links: &dataLinks{
				Self: m.link } })

		if apiErrorsStatus[v.e] > m.status {
		 m.status = apiErrorsStatus[v.e] }
	}

	return rspErrors,m.status
}

func (m *httpRequest) saveErrors() *httpRequest {
	stmt,e := globSqlDB.Prepare("INSERT INTO errors (id,request_id,internal_code,displayed_title,displayed_detail) VALUES (?,?,?,?,?)"); if e != nil {
		globLogger.Error().Err(e).Msg("[REQUEST]: Could not prepare DB statement!")
		m.newError(errInternalSqlError)
		return m }
	defer stmt.Close()

	for _,v := range m.errors {
			globLogger.Info().Str("request_link", m.link).Int("http_code", apiErrorsStatus[v.e]).Str("error_id", v.getId()).Str("error_title", apiErrorsTitle[v.e]).Msg("[REQUEST]:")

		_,e = stmt.Exec(v.getId(), m.id, v.e, apiErrorsTitle[v.e], apiErrorsDetail[v.e]); if e != nil {
			globLogger.Error().Err(e).Str("error_id", v.getId()).Msg("[REQUEST]: Could not write error report!") }
	}

	return m
}
