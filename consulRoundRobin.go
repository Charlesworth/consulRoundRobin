package consulRoundRobin

import (
	"sync"
	"time"
)

type serviceEndpoints struct {
	name      string
	endpoints []string
	index     int
	timeout   <-chan time.Time
}

type serviceMap map[string]*serviceEndpoints

var services = make(serviceMap)
var requestLock = &sync.Mutex{}

//GetServiceEndpoint returns a healthy, round robbined service endpoint
func GetServiceEndpoint(service string) (endpoint string, err error) {
	//requestLock makes all requests synchronus as maps are not concurrent access safe
	requestLock.Lock()
	defer requestLock.Unlock()

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
		errr := services[service].refresh()
		if errr != nil {
			return "", errr
		}
	}

	//return endpoint
	endpoint = services[service].getAndInc()
	return endpoint, nil
}

//possible problem that serviceMap has no *
func (s serviceMap) newService(service string) error {
	endpoints, err := getHealthyEndpoints(service)
	if err != nil {
		return err
	}

	s[service] = &serviceEndpoints{
		name:      service,
		endpoints: endpoints,
		index:     0,
		timeout:   time.After(consulRefreshRate),
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
		timeout:   time.After(consulRefreshRate),
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
