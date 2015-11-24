package goConsulRoundRobin

import "time"

type serviceEndpoints struct {
	name      string
	endpoints []string
	index     int
	timeout   <-chan time.Time
}

func (s *serviceEndpoints) refresh() error {
	endpoints, err := getHealthyEndpoints(s.name)
	if err != nil {
		return err
	}

	if s.index > (len(endpoints) - 1) {
		s.index = 0
	}

	s.endpoints = endpoints
	s.timeout = time.After(consulRefreshRate)

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
