package server

import "errors"
import "runtime"
import "strings"
import "net/http"
import "database/sql"
import "github.com/satori/go.uuid"
import "github.com/rs/zerolog"

const (
	// api errors:
	errNotError = uint8(iota)
	errInternalCommonError
	errInternalSqlError
	errApiNotAuthorized
	errApiUnknownApiFormat
	errApiUnknownType
	errHostsAmbiguousResolver
	errHostsAbnormalIp
	errHostsIpmiTldMismatch
	errHostsIpmiCidrMismatch
	errPortsAbnormalMac
	errJobsJobNotFound
	errRsviewGenericError
	errRsviewAuthError
	errRsviewAuthTestFail
	errRsviewParseError
	errRsviewUnknownApi

	// telegram errors:
	errTgUnknownCommand = uint8(iota)
)

var (
	// common errors:
	errApiCommonTypeInvalid = errors.New("The request type and the link are not the same!")

	// api errors:
	apiErrorsTitle = map[uint8]string{
		errNotError:               "",
		errInternalCommonError:    "Internal error",
		errInternalSqlError:       "Internal database error",
		errApiNotAuthorized:       "Authorization failed",
		errApiUnknownApiFormat:    "Unknown API request format",
		errApiUnknownType:         "Unknown request type",
		errHostsAmbiguousResolver: "Ambiguous resolver answer",
		errHostsAbnormalIp:        "Abnormal IP address",
		errHostsIpmiTldMismatch:   "Ipmi hostname tld mismatch",
		errHostsIpmiCidrMismatch:  "Ipmi CIDR mismatch",
		errPortsAbnormalMac:       "Abnormal MAC address",
		errJobsJobNotFound:        "Job not found",
		errRsviewGenericError:     "Rsview internal error",
		errRsviewAuthError:        "Rsview authorization error",
		errRsviewAuthTestFail:     "Rsview client test error",
		errRsviewParseError: "Rsview parse generic error",
		errRsviewUnknownApi: "Rsview parse error",
	}
	apiErrorsDetail = map[uint8]string{
		errNotError:               "",
		errInternalCommonError:    "The current request could not processed! Please, try again later.",
		errInternalSqlError:       "The current request could not processed due to a database error. Please, try again later.",
		errApiNotAuthorized:       "The current request must be signed with a special key for correct authorization! Please, check your credentials.",
		errApiUnknownApiFormat:    "Could not parse request! Please read the documentation and try again!",
		errApiUnknownType:         "The current request has a type that was sent incorrectly!",
		errHostsAmbiguousResolver: "The given ip address has two or more PTR records! Fix DNS records and try again later.",
		errHostsAbnormalIp:        "The IP address must be in the format \"255.255.255.255\"",
		errHostsIpmiTldMismatch:   "The resolved top-level domain of the ipmi (TLD) does not match the configuration. Correct this discrepancy in the configuration file and try again.",
		errHostsIpmiCidrMismatch:  "The given ipmi address is not included to the configured ipmi CIDR block! Correct this discrepancy in the configuration file and try again.",
		errPortsAbnormalMac:       "The MAC address must be in the format \"ff:ff:ff:ff:ff:ff\"",
		errJobsJobNotFound:        "The requested job was not found in the database!",
		errRsviewGenericError:     "The job failed because of an rsview internal error!",
		errRsviewAuthError:        "The job failed because of an rsview authorization failure! Check the rsview credentials and try again.",
		errRsviewAuthTestFail:     "The job failed because of an rsview client test failure!",
		errRsviewParseError: "The job failed because of rsview parse failure!",
		errRsviewUnknownApi: "The job failed because of rsview parse failure! It's possible that site layout is not the same as before.",
	}
	apiErrorsStatus = map[uint8]int{
		errNotError:               http.StatusOK,
		errInternalCommonError:    http.StatusInternalServerError,
		errInternalSqlError:       http.StatusInternalServerError,
		errApiNotAuthorized:       http.StatusUnauthorized,
		errApiUnknownApiFormat:    http.StatusBadRequest,
		errApiUnknownType:         http.StatusBadRequest,
		errHostsAmbiguousResolver: http.StatusBadRequest,
		errHostsAbnormalIp:        http.StatusBadRequest,
		errHostsIpmiTldMismatch:   http.StatusBadRequest,
		errHostsIpmiCidrMismatch:  http.StatusBadRequest,
		errPortsAbnormalMac:       http.StatusBadRequest,
		errJobsJobNotFound:        http.StatusNotFound,
		errRsviewGenericError:     http.StatusInternalServerError,
		errRsviewAuthError:        http.StatusInternalServerError,
		errRsviewAuthTestFail:     http.StatusInternalServerError,
		errRsviewParseError: http.StatusInternalServerError,
		errRsviewUnknownApi: http.StatusInternalServerError,
	}

	// telegram errors:
	tgErrorsDetail = map[uint8]string{
		errTgUnknownCommand: "The requested command was not found!",
	}
)

type apiError struct {
	e        uint8
	eId      string
	srcParam string
}

type appError struct {
	id string
	code uint8
	prefix string
	jobId string
	requestId string
}

func newAppError(e uint8) *appError {

	var rawFuncName string = "unknown"
	var pcs []uintptr = make([]uintptr, 1)

	if n := runtime.Callers(2, pcs); n != 0 {
		if fun := runtime.FuncForPC(pcs[0] - 1); fun != nil {
			rawFuncName = fun.Name()
		}
	}

	fName := strings.Split(rawFuncName, "/")

	return &appError {
		id: uuid.NewV4().String(),
		prefix: fName[len(fName)-1:][0],
		code: e,
	}
}

func (m *appError) setJobId(jId string) *appError {
	m.jobId = jId
	return m
}

func (m *appError) setRequestId(rId string) *appError {
	m.requestId = rId
	return m
}

func (m *appError) save() bool {

	_,e := globSqlDB.Exec(
		"INSERT INTO errors (id,job_id,request_id,internal_code,displayed_title,displayed_detail) VALUES (?,?,?,?,?,?)",
		m.id, getSqlString(m.jobId), getSqlString(m.requestId), m.code, m.getErrorTitle(), m.getHumanDetails())

	if e != nil {
		globLogger.Error().Err(e).Uint8("errCode", m.code).Str("errTitle", m.getErrorTitle()).Msg("Could not save the error!!!")
		return false
	}

	return true
}

func (m *appError) glCtx() zerolog.Context {
	return globLogger.With()
}

func (m *appError) log(e error, msg string, ctx ...zerolog.Context) *appError { // TODO: try to REFACTOR

	var globLoggerCtx zerolog.Context

	if len(ctx) == 0 {
		globLoggerCtx = m.glCtx()
	} else {
		globLoggerCtx = ctx[0]
	}

	patchedLogger := globLoggerCtx.Err(e).Logger()
	patchedLogger.Error().Msgf("[%s]: " + msg, m.prefix)

	return m
}

func (m *appError) getHttpStatusCode() int {
	return apiErrorsStatus[m.code]
}

func (m *appError) getErrorTitle() string {
	return apiErrorsTitle[m.code]
}

func (m *appError) getHumanDetails() string {
	return apiErrorsDetail[m.code]
}

func newApiError2(e uint8) *apiError {
	return &apiError{
		e: e,
	}
}

func (m *apiError) setError(e uint8) *apiError      { m.e = e; return m }
func (m *apiError) setParameter(p string) *apiError { m.srcParam = p; return m }

func (m *apiError) log(e error, msg string) *apiError {
	globLogger.Error().Err(e).Msg(msg)
	return m
}

func (m *apiError) getId() string {
	if len(m.eId) == 0 {
		m.eId = uuid.NewV4().String()
	}
	return m.eId
}

func getSqlString(in string) (out sql.NullString) {

	out.String = in

	if in == "" {
		return
	}

	out.Valid = true
	return
}
