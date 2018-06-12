package server

import "io"

import "net/http"
import "bufio"
import "strings"


type rsviewClient struct {
	httpClient *http.Client
}


func newRsviewClient() (*rsviewClient, uint8) {

	var rcl = &rsviewClient{
		httpClient: new(http.Client) }

	rq,e := http.NewRequest("GET", globConfig.Base.Rsview.Url, nil); if e != nil {
		newApiError(errInternalCommonError).log(e, "[RSVIEW]: Could not create new httpRequest!")
		return nil,errInternalCommonError }

	rq.SetBasicAuth(
		globConfig.Base.Rsview.Authentication.Login,
		globConfig.Base.Rsview.Authentication.Password)

	rsp,e := rcl.httpClient.Do(rq); if e != nil {
		newApiError(errRsviewAuthError).log(e, "[RSVIEW]: Authentication failed in rsview!")
		return nil,errRsviewAuthError }
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		globLogger.Warn().Int("response_code", rsp.StatusCode).Msg("[RSVIEW]: Abnormal response!")
		newApiError(errRsviewGenericError).log(e, "[RSVIEW]: Response code is not 200")
		return nil,errRsviewGenericError }

	return rcl,rcl.testRsviewClient(rsp.Body)
}

func (m *rsviewClient) testRsviewClient(rBody io.ReadCloser) uint8 {

	var buf = bufio.NewScanner(rBody)

	for buf.Scan() {
		if strings.Contains(buf.Text(), globConfig.Base.Rsview.Authentication.Test_String) {
		 return errNotError } }

	if e := buf.Err(); e != nil {
		newApiError(errInternalCommonError).log(e, "[RSVIEW]: Could not test rsview client because of bufio error!")
		return errInternalCommonError }

	newApiError(errRsviewAuthTestFail).log(nil, "[RSVIEW]: Client test failed!")
	return errRsviewAuthTestFail
}

func (m *rsviewClient) parseMacInfo() {}
