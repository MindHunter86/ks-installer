package server

import "net/http"
import "regexp"

type (
	puppetClient struct {
		htClient *http.Client
		projects map[string]*puppetProject
	}
	puppetProject struct {
		hostRegexp *regexp.Regexp
		apiEndpoints []*projectEndpoints
	}
	projectEndpoints struct {
		vlan, endpoint string
	}
)


func newPuppetClient() *puppetClient {
	return &puppetClient{
		htClient: &http.Client{},
	}
}

func (m *puppetClient) parseEndpoints() error {

	if len(globConfig.Base.Puppet.Projects) == 0 {
		return errPuppetConfigInvalid
	}

	if len(globConfig.Base.Puppet.Endpoints) == 0 {
		return errPuppetConfigInvalid
	}

	m.projects = make(map[string]*puppetProject)

	for k,v := range globConfig.Base.Puppet.Projects {

		rexp,e := regexp.Compile(v)
		if e != nil {
			return e
		}

		m.projects[k] = &puppetProject{
			hostRegexp: rexp,
		}
	}

//	globLogger.Debug().Interface("config", globConfig.Base.Puppet.Endpoints).Msg("")


	for k,v := range globConfig.Base.Puppet.Endpoints {

		if _,ok := globConfig.Base.Puppet.Projects[k]; !ok {
			return errPuppetConfigUnknownProject
		}

		for k2,v2 := range v {
			for _,v3 := range globConfig.Base.Rsview.Access.Vlans {
				if v3 == k2 {
					m.projects[k].apiEndpoint = append(m.projects[k].apiEndpoint, &projectEndpoints{
						vlan: v3,
						endpoint: v2,
					})
				}
			}
		}

		if m.projects[k].apiEndpoint == "" {
			return errPuppetConfigUnknownVlan
		}
	}

/*
	for k,_ := range m.projects {
		if k2,ok := globConfig.Base.Puppet.Endpoints[k]; !ok && k2 != "" {
			return errPuppetConfigUnknownProject
		}

		for _,v := range globConfig.Base.Rsview.Access.Vlans {
			if k3,ok2 := globConfig.Base.Puppet.Endpoints[k][v]; !ok2 && k3 != "" {
				return errPuppetConfigUnknownVlan
			} else {
				m.projects[k].apiEndpoint = k3
			}
		}
	}
*/
	// debug:
	for k,v := range m.projects {
		globLogger.Debug().Str("project", k).Str("regexp", v.hostRegexp.String()).Str("endpoint", v.apiEndpoint).Msg("")
	}

	return nil
}
