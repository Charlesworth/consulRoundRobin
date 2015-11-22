package consulRoundRobin

/*TODO
- set global timeout
- make the health agent type
- sort out test agent init
- make channel input output
*/

import (
	"time"
)

type serviceEndpoints struct {
	name      string
	endpoints []string
	index     int
	timeout   <-chan time.Time
}

// var services map[string]*serviceEndpoints
type serviceMap map[string]*serviceEndpoints

var services = make(serviceMap)
var endpointTimeOut = time.Minute

//possible error that serviceMap has no *
func (s serviceMap) newService(service string) error {
	endpoints, err := getHealthyEndpoints(service)
	if err != nil {
		return err
	}

	s[service] = &serviceEndpoints{
		name:      service,
		endpoints: endpoints,
		index:     0,
		timeout:   time.After(endpointTimeOut),
	}

	return nil
}

func (s *serviceEndpoints) refresh() error {
	endpoints, err := getHealthyEndpoints(s.name)
	if err != nil {
		return err
	}

	if s.index > (len(endpoints) - 1) {
		s.index = 0
	}

	s = &serviceEndpoints{
		name:      s.name,
		endpoints: endpoints,
		index:     s.index,
		timeout:   time.After(endpointTimeOut),
	}

	return nil
}

func (s *serviceEndpoints) getAndInc() (endpoint string) {
	endpoint = s.endpoints[s.index]
	s.index = (s.index + 1) % len(s.endpoints)
	return endpoint
}

func (s *serviceEndpoints) timedOut() bool {
	select {
	case <-s.timeout:
		return true
	default:
		return false
	}
}

//GetServiceEndpoint returns a healthy, round robbined service endpoint
func GetServiceEndpoint(service string) (endpoint string, err error) {
	//if new service request
	if _, present := services[service]; !present {
		//make new service and return endpoint
		err = services.newService(service)
		if err != nil {
			return "", err
		}
		endpoint = services[service].getAndInc()
		return endpoint, nil
	}

	//if timeout
	if services[service].timedOut() {
		//refresh endpoints
		err := services[service].refresh()
		if err != nil {
			return "", err
		}
	}

	//return endpoint
	endpoint = services[service].getAndInc()
	return endpoint, nil
}

type agentStruct struct {
	endpoints []string
	err       error
}

var agent agentStruct

func getHealthyEndpoints(service string) (endpoints []string, err error) {
	return agent.endpoints, agent.err
}
