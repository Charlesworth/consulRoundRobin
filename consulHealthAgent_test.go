package goConsulRoundRobin

import (
	"testing"

	consulapi "github.com/hashicorp/consul/api"
)

func newAgentService(port int, address string) *consulapi.AgentService {
	return &consulapi.AgentService{
		ID:      "",
		Service: "",
		Tags:    []string{""},
		Port:    port,
		Address: address,
	}
}

func TestGetEndpointsFromServiceEntries(t *testing.T) {
	testAgentService1 := newAgentService(8080, "bluemix.fake.net")
	testAgentService2 := newAgentService(8081, "bluemix.fake.net")
	testServiceEntry1 := &consulapi.ServiceEntry{Service: testAgentService1}
	testServiceEntry2 := &consulapi.ServiceEntry{Service: testAgentService2}
	testServiceEntryArray := []*consulapi.ServiceEntry{testServiceEntry1, testServiceEntry2}

	endpoints := getEndpointsFromServiceEntries(testServiceEntryArray)
	if (endpoints[0] != "bluemix.fake.net:8080") || (endpoints[1] != "bluemix.fake.net:8081") {
		t.Error("getEndpointsFromServiceEntries did not return the correct endpoints")
	}
}
