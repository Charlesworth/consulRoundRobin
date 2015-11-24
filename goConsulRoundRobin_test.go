package goConsulRoundRobin

import (
	"errors"
	"testing"
	"time"
)

type BehaviorTest struct {
	testName             string
	consulAgentEndpoints []string
	consulAgentError     error
	serviceName          string
	expectedEndpointOut  string
	sleep                bool
}

var testCases = []BehaviorTest{
	{testName: "new service (working consul request)",
		consulAgentEndpoints: []string{"test.com/test1", "test.com/test2"},
		consulAgentError:     nil,
		serviceName:          "testService1",
		expectedEndpointOut:  "test.com/test1"},
	{testName: "new service (error consul request)",
		consulAgentEndpoints: []string{""},
		consulAgentError:     errors.New("consul error"),
		serviceName:          "testService2",
		expectedEndpointOut:  ""},
	{testName: "refresh existing service (error consul request)",
		consulAgentEndpoints: []string{""},
		consulAgentError:     errors.New("consul error"),
		serviceName:          "testService1",
		expectedEndpointOut:  "",
		sleep:                true},
	{testName: "refresh existing service (working consul request)",
		consulAgentEndpoints: []string{"test.com/test1", "test.com/test2"},
		consulAgentError:     nil,
		serviceName:          "testService1",
		expectedEndpointOut:  "test.com/test2",
		sleep:                true},
	{testName: "new service (working consul request)",
		consulAgentEndpoints: []string{"test.com/test1", "test.com/test2"},
		consulAgentError:     nil,
		serviceName:          "testService3",
		expectedEndpointOut:  "test.com/test1"},
	{testName: "existing service, refresh timeout not triggered",
		consulAgentEndpoints: []string{"test.com/test1", "test.com/test2"},
		consulAgentError:     nil,
		serviceName:          "testService3",
		expectedEndpointOut:  "test.com/test2"},
}

func TestGetServiceEndpoint(t *testing.T) {
	for i := range testCases {
		t.Log(testCases[i].testName)

		//sleep if a refresh test to trigger timeout
		if testCases[i].sleep {
			time.Sleep(time.Millisecond * 10)
		}

		testAgent.endpoints = testCases[i].consulAgentEndpoints
		testAgent.err = testCases[i].consulAgentError

		endpoint, err := GetServiceEndpoint(testCases[i].serviceName)

		testsPassed := true
		if endpoint != testCases[i].expectedEndpointOut {
			t.Error("ERROR: endpoint expected:", testCases[i].expectedEndpointOut, " endpoint recieved:", endpoint)
			testsPassed = false
		}

		if err != testCases[i].consulAgentError {
			t.Error("ERROR: error expected: {", testCases[i].consulAgentError, "} error recieved: {", err, "}")
			testsPassed = false
		}

		if testsPassed {
			t.Log("PASS")
		}
	}
}

func TestServiceMapNewService(t *testing.T) {
	//test getHealthyEndpoints error
	testAgent.endpoints = []string{}
	testAgent.err = errors.New("get healthly endpoints failure")
	testServiceMap := make(serviceMap)
	err := testServiceMap.newService("testNewErrorCase")
	if err == nil {
		t.Error("no error reported despite getHealthyEndpoints returning an error")
	}
	_, present := testServiceMap["testNewErrorCase"]
	if present {
		t.Error("service 'testNewErrorCase' was added to testServiceMap map when getHealthyEndpoints returned an error")
	}

	//test valid test case
	testAgent.endpoints = []string{"test.com/test1", "test.com/test2", "test.com/test3"}
	testAgent.err = nil
	err = testServiceMap.newService("testNewValidCase")
	if err != nil {
		t.Error("adding a valid test case returned error:", err)
	}
	_, present = testServiceMap["testNewValidCase"]
	if !present {
		t.Error("service 'testNewValidCase' wasn't added to testServiceMap map")
	}
}
