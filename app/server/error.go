package server

import "errors"
import "net/http"
import "github.com/satori/go.uuid"

const (
	// api errors:
	errNotError = uint8(iota)
	errInternalCommonError
	errInternalSqlError
	errApiNotAuthorized
	errApiUnknownApiFormat
	errApiUnknownType
	errHostsAmbiguousResolver
	errHostsAbnormalMac
	errHostsAbnormalIp
	errHostsIpmiTldMismatch
	errHostsIpmiCidrMismatch
	errJobsJobNotFound
	errRsviewGenericError
	errRsviewAuthError
	errRsviewAuthTestFail

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
		errHostsAbnormalMac:       "Abnormal MAC address",
		errHostsAbnormalIp:        "Abnormal IP address",
		errHostsIpmiTldMismatch:   "Ipmi hostname tld mismatch",
		errHostsIpmiCidrMismatch:  "Ipmi CIDR mismatch",
		errJobsJobNotFound:        "Job not found",
		errRsviewGenericError:     "Rsview internal error",
		errRsviewAuthError:        "Rsview authorization error",
		errRsviewAuthTestFail:     "Rsview client test error",
	}
	apiErrorsDetail = map[uint8]string{
		errNotError:               "",
		errInternalCommonError:    "The current request could not processed! Please, try again later.",
		errInternalSqlError:       "The current request could not processed due to a database error. Please, try again later.",
		errApiNotAuthorized:       "The current request must be signed with a special key for correct authorization! Please, check your credentials.",
		errApiUnknownApiFormat:    "Could not parse request! Please read the documentation and try again!",
		errApiUnknownType:         "The current request has a type that was sent incorrectly!",
		errHostsAmbiguousResolver: "The given ip address has two or more PTR records! Fix DNS records and try again later.",
		errHostsAbnormalMac:       "The MAC address must be in the format \"ff:ff:ff:ff:ff:ff\"",
		errHostsAbnormalIp:        "The IP address must be in the format \"255.255.255.255\"",
		errHostsIpmiTldMismatch:   "The resolved top-level domain of the ipmi (TLD) does not match the configuration. Correct this discrepancy in the configuration file and try again.",
		errHostsIpmiCidrMismatch:  "The given ipmi address is not included to the configured ipmi CIDR block! Correct this discrepancy in the configuration file and try again.",
		errJobsJobNotFound:        "The requested job was not found in the database!",
		errRsviewGenericError:     "The job failed because of an rsview internal error!",
		errRsviewAuthError:        "The job failed because of an rsview authorization failure! Check the rsview credentials and try again.",
		errRsviewAuthTestFail:     "The job failed because of an rsview client test failure!",
	}
	apiErrorsStatus = map[uint8]int{
		errNotError:               0,
		errInternalCommonError:    http.StatusInternalServerError,
		errInternalSqlError:       http.StatusInternalServerError,
		errApiNotAuthorized:       http.StatusUnauthorized,
		errApiUnknownApiFormat:    http.StatusBadRequest,
		errApiUnknownType:         http.StatusBadRequest,
		errHostsAmbiguousResolver: http.StatusBadRequest,
		errHostsAbnormalMac:       http.StatusBadRequest,
		errHostsAbnormalIp:        http.StatusBadRequest,
		errHostsIpmiTldMismatch:   http.StatusBadRequest,
		errHostsIpmiCidrMismatch:  http.StatusBadRequest,
		errJobsJobNotFound:        http.StatusNotFound,
		errRsviewGenericError:     http.StatusInternalServerError,
		errRsviewAuthError:        http.StatusInternalServerError,
		errRsviewAuthTestFail:     http.StatusInternalServerError,
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

func newApiError(e uint8) *apiError {
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
