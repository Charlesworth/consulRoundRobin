package consulRoundRobin

import (
	"log"
	"os"
	"strconv"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

type consulHealthAgent interface {
	Service(service, tag string, passingOnly bool, q *consulapi.QueryOptions) ([]*consulapi.ServiceEntry, *consulapi.QueryMeta, error)
}

type agentStruct struct {
	endpoints []string
	err       error
}

var consulIP string
var consulRefreshRate time.Duration
var healthClient consulHealthAgent
var testAgent agentStruct

func init() {
	consulRefreshRateString := os.Getenv("CONSUL_REFRESH_RATE")

	var err error
	consulRefreshRateInt, err := strconv.Atoi(consulRefreshRateString)
	if err != nil {
		log.Fatal(err)
	}
	consulRefreshRate = time.Second * time.Duration(consulRefreshRateInt)

	consulIP = os.Getenv("CONSUL_IP")
	//***** if consul ip empty then use test client *****
	if consulIP == "test" {
		//set test agent here
		consulRefreshRate = time.Millisecond
	}

	healthClient, err = newConsulHealthClient(consulIP, ":8500")
	if err != nil {
		log.Fatal(err)
	}
}

func getHealthyEndpoints(service string) (serviceEndpoints []string, err error) {
	if consulIP == "test" {
		return testAgent.endpoints, testAgent.err
	}
	agent := healthClient

	healthyServiceEntries, _, err := agent.Service(service, "", true, &consulapi.QueryOptions{})
	if err != nil {
		return []string{}, err
	}

	serviceEndpoints = getEndpointsFromServiceEntries(healthyServiceEntries)

	return serviceEndpoints, nil
}

// how to test, need a mocked consul server?
func newConsulHealthClient(consulIP string, consulPort string) (*consulapi.Health, error) {
	config := consulapi.DefaultConfig()
	config.Address = consulIP + consulPort

	consul, err := consulapi.NewClient(config)
	if err != nil {
		return &consulapi.Health{}, err
	}

	consulHealthAgent := consul.Health()
	return consulHealthAgent, nil
}

func getEndpointsFromServiceEntries(serviceEntries []*consulapi.ServiceEntry) (serviceEndpoints []string) {
	for i := 0; i < len(serviceEntries); i++ {
		serviceAddress := serviceEntries[i].Service.Address
		servicePortInt := serviceEntries[i].Service.Port
		serviceEndpoint := serviceAddress + ":" + strconv.Itoa(servicePortInt)
		serviceEndpoints = append(serviceEndpoints, serviceEndpoint)
	}
	return serviceEndpoints
}

func checkErrFatal(err error) {
	if err != nil {
		log.Fatalln("consulRoundRobin Fatal Error: ", err)
	}
}
