package goConsulRoundRobin

import (
	"sync"
	"time"
)

type serviceMap map[string]*serviceEndpoints

var services = make(serviceMap)
var requestLock = &sync.Mutex{}

//GetServiceEndpoint returns a healthy, round robbined service endpoint
func GetServiceEndpoint(service string) (endpoint string, err error) {
	//requestLock makes all requests synchronus as maps are not thread safe
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
		err := services[service].refresh()
		if err != nil {
			return "", err
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
