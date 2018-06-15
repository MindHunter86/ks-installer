package server

import "time"
import "encoding/hex"
import "crypto/sha256"
import "crypto/hmac"
import "bytes"
import "strings"
import "io/ioutil"
import "net/http"
import "encoding/json"
import "github.com/gorilla/mux"
import "github.com/gorilla/context"

// JSON response structs:
// Recomendations are taken from jsonapi.org:
type (
	// main module struct:
	apiController struct{}

	// JSON response structs:
	apiResponse struct {
		Data    *responseData    `json:"data,omitempty"`
		Errors  []*responseError `json:"errors,omitempty"`
		Meta    *responseMeta    `json:"meta,omitempty"`
		JsonApi *responseJsonApi `json:"jsonapi,omitempty"`
		Links   *responseLinks   `json:"links,omitempty"`
		Debug   *responseDebug   `json:"debug,omitempty"`
	}
	responseData struct {
		Type           string              `json:"type,omitempty"`
		Id             string              `json:"id,omitempty"`
		JobAttributes  *dataJobAttributes  `json:"attributes,omitempty"`  // BUG: json attributes duplicate!
		HostAttributes *dataHostAttributes `json:"attributes1,omitempty"` // TODO: refactor needed!
	}
	dataHostAttributes struct {
		Hostid string                 `json:"hostid,omitempty"`
		Ipmi   *hostAttributesIpmi    `json:"ipmi,omitempty"`
		Ports  []*hostAttributesPorts `json:"ports,omitempty"`
		//	Jobs []string
		Updated_At *time.Time `json:"updated_at,omitempty"`
		Created_At *time.Time `json:"created_at,omitempty"`
	}
	hostAttributesIpmi struct {
		Ptr_Name   string `json:"ptr_name,omitempty"`
		Ip_Address string `json:"ip_address,omitempty"`
	}
	hostAttributesPorts struct {
		Name       string     `json:"name,omitempty"`
		Jun        uint16     `json:"jun,omitempty"`
		Vlan       uint16     `json:"vlan,omitempty"`
		Mac        string     `json:"mac,omitempty"`
		Updated_At *time.Time `json:"updated_at,omitempty"`
	}
	hostAttributesJobs struct {
		Id         string       `json:"id,omitempty"`
		Payload    *jobsPayload `json:"payload,omitempty"`
		Updated_At string       `json:"updated_at,omitempty"`
		Created_At string       `json:"created_at,omitempty"`
	}
	dataJobAttributes struct {
		Job *jobAttributesJob `json:"job,omitempty"`
	}
	jobAttributesJob struct {
		Id         string       `json:"id,omitempty"`
		Payload    *jobsPayload `json:"payload,omitempty"`
		Updated_At string       `json:"updated_at,omitempty"`
		Created_At string       `json:"created_at,omitempty"`
	}
	jobsPayload struct {
		Action string `json:"action,omitempty"`
		State  string `json:"state,omitempty"`
		// TODO: add Errors
	}
	responseError struct {
		Id     string       `json:"id,omitempty"`
		Code   int          `json:"code,omitempty"`
		Status int          `json:"status,omitempty"`
		Title  string       `json:"title,omitempty"`
		Detail string       `json:"detail,omitempty"`
		Source *errorSource `json:"source,omitempty"`
	}
	errorSource struct {
		Parameter string `json:"parameter,omitempty"`
	}

	// JSON request structs:
	apiHostPostRequest struct {
		Data *hostRequestData `json:"data"`
	}
	hostRequestData struct {
		Type       string              `json:"type"`
		Attributes *dataHostAttributes `json:"attributes"`
	}

	// JSON meta information:
	responseMeta struct {
		ApiVersion string   `json:"api_version"`
		Copyright  string   `json:"copyright"`
		Authors    []string `json:"authors"`
	}

	// JSON links:
	responseLinks struct {
		Self string `json:"self"`
	}

	// JSON standart version:
	responseJsonApi struct {
		Version string `json:"version"`
	}

	// JSON debug:
	responseDebug struct {
		RequestId string `json:"request_id,omitempty"`
	}
)

func NewApiController() *mux.Router {

	globApi = new(apiController)

	var r = mux.NewRouter()
	r.Host(globConfig.Base.Http.Host)
	r.Use(globApi.httpMiddlewareRequestLog)

	s := r.PathPrefix("/v1").Headers("Content-Type", "application/vnd.api+json").Subrouter()
	s.Use(globApi.httpMiddlewareAPIAuthentication)

	s.HandleFunc("/", globApi.httpHandlerRootV1).Methods("GET")

	s.HandleFunc("/host/{mac:(?:[0-9A-Fa-f]{2}[:-]){5}(?:[0-9A-Fa-f]{2})}", globApi.httpHandlerHostGet).Methods("GET")
	s.HandleFunc("/host", globApi.httpHandlerHostCreate).Methods("POST")

	s.HandleFunc("/job/{id:(?:[0-9a-f]{8}-)(?:[0-9a-f]{4}-){3}(?:[0-9a-f]{12})}", globApi.httpHandlerJobGet).Methods("GET")

	// TODO: reload the job if it does not work (failed
	// /v1/job/UUID?try_again=1

	return r
}

func (m *apiController) httpMiddlewareRequestLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req, e := new(httpRequest).createAndSave(r)
		if e != nil {
			globLogger.Error().Err(e).Msg("[API]: Could not save request in the database!")
			req.newError(errInternalCommonError)
		}
		context.Set(r, "internal_request", req)

		h.ServeHTTP(w, r)

		req.updateAndSave()
	})
}

func (m *apiController) httpMiddlewareAPIAuthentication(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var req = context.Get(r, "internal_request").(*httpRequest)

		var bodyBuf bytes.Buffer
		bufSize, e := bodyBuf.ReadFrom(r.Body)
		if !m.errorHandler(w, e, req) {
			return
		}
		r.Body.Close()

		mac := hmac.New(sha256.New, []byte(globConfig.Base.Api.Sign_Secret))
		macSize, e := mac.Write(bodyBuf.Bytes())
		if !m.errorHandler(w, e, req) {
			return
		}

		if r.ContentLength != bufSize || r.ContentLength != int64(macSize) {
			globLogger.Warn().Msg("[API]: Different sizes in request, buffer and mac!")
		}

		expectedMAC := mac.Sum(nil)
		receivedMAC, e := hex.DecodeString(strings.Split(
			r.Header.Get("Authorization"), " ")[1])
		if !m.errorHandler(w, e, req) {
			return
		}

		globLogger.Debug().Str("mac_expected", hex.EncodeToString(expectedMAC)).Msg("[API]: HMAC sings comparison")
		globLogger.Debug().Str("mac_received", strings.Split(r.Header.Get("Authorization"), " ")[1]).Msg("[API]: HMAC signs comparison")

		if !hmac.Equal(expectedMAC, receivedMAC) {
			req.newError(errApiNotAuthorized)
			m.respondJSON(w, req, nil, 0)
			return
		}

		r.Body = ioutil.NopCloser(bytes.NewReader(bodyBuf.Bytes()))
		h.ServeHTTP(w, r)
	})
}

func (m *apiController) httpHandlerRootV1(w http.ResponseWriter, r *http.Request) {}

func (m *apiController) httpHandlerJobGet(w http.ResponseWriter, r *http.Request) {

	var req = context.Get(r, "internal_request").(*httpRequest)
	var vars = mux.Vars(r)

	if vars["id"] == "" {
		req.appendAppError(newAppError(errApiUnknownApiFormat))
		m.respondJSON(w, req, nil, 0)
		return
	}

	jb,err := getJobById(vars["id"]); if err != nil {
		req.appendAppError(err)
		m.respondJSON(w, req, nil, 0)
		return
	}

	m.respondJSON(w, req, &responseData{
		Type: "job",
		Id:   req.id,
		JobAttributes: &dataJobAttributes{
			Job: &jobAttributesJob{
				Id: jb.id,
				Payload: &jobsPayload{
					Action: jobActHumanDetail[jb.action],
					State:  jobStatusHumanDetail[jb.state],
				},
				Updated_At: jb.updated_at.String(),
				Created_At: jb.created_at.String(),
			},
		}}, http.StatusOK)
}

func (m *apiController) httpHandlerHostGet(w http.ResponseWriter, r *http.Request) {
	var req = context.Get(r, "internal_request").(*httpRequest)
	vars := mux.Vars(r)

	if vars["mac"] == "" {
		req.newError(errApiUnknownApiFormat)
		m.respondJSON(w, req, nil, 0)
		return
	}

	// TODO: get host by MAC from host.go
}

func (m *apiController) httpHandlerHostCreate(w http.ResponseWriter, r *http.Request) {

	var req = context.Get(r, "internal_request").(*httpRequest)

	var postRequest *apiHostPostRequest
	rspBody, e := ioutil.ReadAll(r.Body)
	if !m.errorHandler(w, e, req) {
		return
	}
	e = json.Unmarshal(rspBody, &postRequest)
	if !m.errorHandler(w, e, req) {
		return
	}

	switch {
	case postRequest.Data == nil:
		fallthrough
	case postRequest.Data.Type == "":
		fallthrough
	case postRequest.Data.Attributes == nil:
		fallthrough
	case postRequest.Data.Attributes.Ports == nil:
		fallthrough
	case len(postRequest.Data.Attributes.Ports) == 0:
		fallthrough
	case false: // something impossible
		req.newError(errApiUnknownApiFormat)
		m.respondJSON(w, req, nil, 0)
		return
	case postRequest.Data.Type != "host":
		req.newError(errApiUnknownType)
		m.respondJSON(w, req, nil, 0)
		return
	}

	// test given ipmi && mac addresses:
	var ipAddr *string
	if ipAddr = &postRequest.Data.Attributes.Ipmi.Ip_Address; *ipAddr == "" {
		req.appendAppError(newAppError(errHostsAbnormalIp))
		m.respondJSON(w, req, nil, 0)
		return
	}

	var macs []*string
	for _, v := range postRequest.Data.Attributes.Ports {
		if v.Mac == "" {
			req.appendAppError(newAppError(errPortsAbnormalMac))
			m.respondJSON(w, req, nil, 0)
			return
		}

		macs = append(macs, &v.Mac)
	}

	// parse given ipmi && mac addresses:
	var host = newHost()
	if e := host.parseIpmiAddress(ipAddr); e != nil {
		req.appendAppError(e)
		m.respondJSON(w, req, nil, 0)
		return
	}

	var port *basePort = newPort()
	if e := port.parseMacAddress(macs); e != nil {
		req.appendAppError(e)
		m.respondJSON(w, req, nil, 0)
		return
	}

	// add jobs and respond:
	job,err := newQueueJob(&req.id, jobActHostCreate); if err != nil {
		req.appendAppError(err)
		m.respondJSON(w, req, nil, 0)
		return
	}

	job.setPayload(&map[string]interface{}{
		"job_payload_host": host})

	job.addToQueue()

	m.respondJSON(w, req, &responseData{
		Type: "job",
		Id:   req.id,
		JobAttributes: &dataJobAttributes{
			Job: &jobAttributesJob{
				Id: job.id,
				Payload: &jobsPayload{ // TODO: rename Payload to Properties or Attributes or Meta(-Data)
					Action: jobActHumanDetail[job.action],
					State:  jobStatusHumanDetail[job.state],
				},
				Updated_At: job.updated_at.String(),
				Created_At: job.created_at.String(),
			},
		}}, http.StatusCreated)
}

func (m *apiController) errorHandler(w http.ResponseWriter, e error, req *httpRequest) bool {
	if e == nil {
		return true
	}

	req.newError(errInternalCommonError)
	globLogger.Error().Err(e).Msg("[API]: Abnormal function result!")

	m.respondJSON(w, req, nil, 0)
	return false
}

func (m *apiController) respondJSON(w http.ResponseWriter, req *httpRequest, payloadData *responseData, status int) {
	//
	var rspPayload = &apiResponse{
		Data: payloadData,
		Meta: &responseMeta{
			ApiVersion: appVersion,
			Authors: []string{
				"vadimka_kom"},
			Copyright: "Copyright 2018 Mindhunter and CO."},
		Links: &responseLinks{
			Self: req.link},
		JsonApi: &responseJsonApi{
			Version: "1.0"},
	}

	if globLogger.Debug().Enabled() {
		rspPayload.Debug = &responseDebug{
			RequestId: req.id}
	}

	globLogger.Debug().Int("http_pre_status", status).Msg("")

	if rspPayload.Errors = req.saveErrors().respondApiErrors(); req.status > status {
		status = req.status
		rspPayload.Data = nil
	}

	req.status = status // TODO: refactor

	globLogger.Debug().Int("http_post_status", status).Msg("")

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(rspPayload)
}
