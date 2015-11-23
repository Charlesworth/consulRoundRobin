package consulRoundRobin

/*TODO
- set global timeout
- make the health agent type
- sort out test agent init
- make channel input output
*/

import (
	"log"
	"os"
	"strconv"
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

// var consulRefreshRate = time.Minute

var consulIP string
var consulRefreshRate time.Duration

func init() {

	//***** if consul ip empty then use test client *****
	consulIP = os.Getenv("CONSUL_IP")
	consulRefreshRateString := os.Getenv("CONSUL_REFRESH_RATE")

	var err error
	consulRefreshRateInt, err := strconv.Atoi(consulRefreshRateString)
	if err != nil {
		log.Fatal(err)
	}
	consulRefreshRate = time.Second * time.Duration(consulRefreshRateInt)
}

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
		errr := services[service].refresh()
		if errr != nil {
			return "", errr
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
