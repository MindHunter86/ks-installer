package server

import "time"

import "io"
import "crypto/tls"

import "net"
import "net/url"
import "net/http"

// import "net/http/httputil"

import "bufio"
import "strings"
import "bytes"

import "golang.org/x/net/html"
import "golang.org/x/net/html/atom"

const (
	rsviewTableRescan = uint8(iota)
	rsviewTableHostname
	rsviewTablePort
	rsviewTableBoundle
	rsviewTableForceUp
	rsviewTableVlans
	rsviewTableAdmOpStatus
	rsviewTableLldp
	rsviewTableMacList
	rsviewTableLinkIp
	rsviewTableRipList
	rsviewTableLinkIpByName
	rsviewTableRipIpByName
	rsviewTableRack
	rsviewTableZoneName
	rsviewTableDcName
	rsviewTablePortFlapped
	rsviewTableLastScan
)

var (
	rsviewTableHuman = map[uint8]string{
		rsviewTableRescan:       "Rescan link",
		rsviewTableHostname:     "Hostname",
		rsviewTablePort:         "Port",
		rsviewTableBoundle:      "Boundle",
		rsviewTableForceUp:      "Force-UP",
		rsviewTableVlans:        "Vlans",
		rsviewTableAdmOpStatus:  "Admin/Oper status",
		rsviewTableLldp:         "LLDP",
		rsviewTableMacList:      "Mac list",
		rsviewTableLinkIp:       "Link IP",
		rsviewTableRipList:      "RIP list",
		rsviewTableLinkIpByName: "Link IP by name",
		rsviewTableRipIpByName:  "RIP IP by name",
		rsviewTableRack:         "Rack",
		rsviewTableZoneName:     "Zone name",
		rsviewTableDcName:       "DC name",
		rsviewTablePortFlapped:  "Port flapped",
		rsviewTableLastScan:     "Last scan",
	}
)

type rsviewClient struct {
	httpClient *http.Client

	httpAuthHeader string
}

func newRsviewClient() (*rsviewClient, *appError) {

	var rcl = &rsviewClient{
		httpClient: &http.Client{
			Timeout: globConfig.GetDuration("base.rsview.client.timeout") * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: globConfig.GetBool("base.rsview.client.insecure_skip_verify")},
			},
		},
	}

	rq, e := http.NewRequest("GET", globConfig.Base.Rsview.Url, nil)
	if e != nil {
		return nil, newAppError(errInternalCommonError).log(e, "Could not create new httpRequest!")
	}

	rq.SetBasicAuth(
		globConfig.GetString("base.rsview.authentication.login"),
		globConfig.GetString("base.rsview.authentication.password"))

	//dump,e := httputil.DumpRequest(rq, true); if e != nil {
	//	newApiError(errInternalCommonError).log(e, "Request dump error!")
	//	return nil,errInternalCommonError }
	//globLogger.Debug().Bytes("request", dump).Msg("")

	rsp, e := rcl.httpClient.Do(rq)
	if e != nil {
		return nil, newAppError(errRsviewGenericError).log(e, "Could not do the request!")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		if rsp.StatusCode == http.StatusUnauthorized {
			return nil, newAppError(errRsviewAuthError).log(nil, "Authentication failed in rsview!")
		}

		globLogger.Warn().Int("response_code", rsp.StatusCode).Msg("[RSVIEW]: Abnormal response!")
		ae := newAppError(errRsviewGenericError)
		return nil, ae.log(nil, "Response code is not 200!", ae.glCtx().Int("http_code", rsp.StatusCode))
	}

	rcl.httpAuthHeader = rsp.Request.Header.Get("Authorization")
	globLogger.Debug().Str("auth_header", rcl.httpAuthHeader).Msg("Rsview HTTP Basic session")

	return rcl, rcl.testRsviewClient(rsp.Body)
}

func (m *rsviewClient) testRsviewClient(rBody io.ReadCloser) *appError {

	var buf = bufio.NewScanner(rBody)

	for buf.Scan() {
		if strings.Contains(buf.Text(), globConfig.Base.Rsview.Authentication.Test_String) {
			return nil
		}
	}

	if e := buf.Err(); e != nil {
		return newAppError(errInternalCommonError).log(e, "Could not test rsview client because of bufio error!")
	}

	return newAppError(errRsviewAuthTestFail).log(nil, "Client test failed!")
}

func (m *rsviewClient) getPortAttributes(mac net.HardwareAddr) ([]string, *appError) {

	rqUrl, e := url.Parse(globConfig.Base.Rsview.Url)

	// mask our url:

	var urlArgs = rqUrl.Query()

	urlArgs.Set("hostname", "")
	urlArgs.Set("dc_name", "")
	urlArgs.Set("vlan", "")
	urlArgs.Set("lldp_neighbour", "")
	urlArgs.Set("boundle", "")
	urlArgs.Set("port", "")
	urlArgs.Set("dns_link", "")
	urlArgs.Set("link_ip", "")
	urlArgs.Set("rip_ip", "")
	urlArgs.Set("mac", mac.String())
	urlArgs.Set("dns_rip", "")
	urlArgs.Set("zone_name", "")
	urlArgs.Set("rack", "")
	urlArgs.Set("search", "Proceed")
	urlArgs.Set("proceed", "proceed")

	rqUrl.RawQuery = urlArgs.Encode()

	rq, e := http.NewRequest("GET", rqUrl.String(), nil)
	if e != nil {
		return nil, newAppError(errInternalCommonError).log(e, "Could not create new httpRequest!")
	}

	// mask our request
	rq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:58.0) Gecko/20100101 Firefox/58.0")
	rq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	rq.Header.Set("Accept-Language", "en,ru;q=0.5")
	rq.Header.Set("Referer", "https://nw.net.mail.ru/rsview/rsview.py")
	rq.Header.Set("Pragma", "no-cache")
	rq.Header.Set("Cache-Control", "no-cache")

	rq.Header.Set("Authorization", m.httpAuthHeader)

	//	dump, e := httputil.DumpRequest(rq, true)
	//	if e != nil {
	//		return nil,newAppError(errInternalCommonError).log(e, "Request dump error!")
	//	}
	//	globLogger.Debug().Bytes("request", dump).Msg("")

	rsp, e := m.httpClient.Do(rq)
	if e != nil {
		return nil, newAppError(errRsviewGenericError).log(e, "Could not do the request!")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		ae := newAppError(errRsviewGenericError)
		return nil, ae.log(nil, "Response code is not 200!", ae.glCtx().Int("http_code", rsp.StatusCode))
	}

	// TODO: XXX: Do we need JUN rescan before page parse ?

	return m.parseResponseHTML(rsp.Body)
}

func (m *rsviewClient) parseResponseHTML(rBody io.ReadCloser) ([]string, *appError) {

	var buf []string
	var trResultCount int
	var tdClassResult, tdTextReaded bool

	var z *html.Tokenizer = html.NewTokenizer(rBody)

LOOP:
	for {
		switch z.Next() {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return buf, newAppError(errInternalCommonError).log(z.Err(), "Tokenizer generic error!")
			}
			break LOOP

		case html.StartTagToken:
			tkn := z.Token()

			switch tkn.DataAtom {
			case atom.Tr:
				for _, attr := range tkn.Attr {
					if attr.Key == "class" && attr.Val == "result" {
						if trResultCount++; trResultCount > 3 {
							break
						}
						continue
					}
				}
			case atom.Td:
				if len(tkn.Attr) == 0 {
					continue
				}
				if tkn.Attr[0].Key != "class" || (tkn.Attr[0].Val != "result_table2" && tkn.Attr[0].Val != "popup") {
					continue
				}

				tdClassResult = true
			}

		case html.EndTagToken:
			tkn := z.Token()
			if tkn.DataAtom != atom.Td || !tdClassResult {
				continue
			}
			if !tdTextReaded {
				buf = append(buf, "NULL")
			}
			tdClassResult = false
			tdTextReaded = false

		case html.TextToken:
			_ = z.Token() //- test it, please (XXX: 2DELETE)
			if !tdClassResult {
				continue
			}
			bData := bytes.Replace(bytes.ToLower(bytes.TrimSpace(z.Raw())), []byte("none"), []byte(""), -1)
			if bytes.Compare(bData, []byte(" ")) != 0 && len(bData) != 0 {
				if tdTextReaded {
					lastTest := len(buf) - 1
					buf[lastTest] = buf[lastTest] + " " + string(bData)
					continue
				}
				buf = append(buf, string(bData))
				tdTextReaded = true
			}
		}
	}

	if len(buf) == 0 {
		return buf, newAppError(errRsviewMacNotFound).log(nil, "Buffer is empty after parsing (seems mac was not found)!")
	}

	if len(buf) != len(rsviewTableHuman) {
		return buf, newAppError(errRsviewUnknownApi).log(nil, "Comparison of buffer and template failed! Check rsview site layout!")
	}

	for k, v := range buf {
		globLogger.Debug().Str("human", rsviewTableHuman[uint8(k)]).Str("value", v).Msg("parsed value")
	}

	return buf, nil
}
