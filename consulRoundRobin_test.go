package consulRoundRobin

import (
	"errors"
	"testing"
	"time"
)

func TestServiceMapNewService(t *testing.T) {
	//test getHealthyEndpoints error
	agent.endpoints = []string{}
	agent.err = errors.New("get healthly endpoints failure")
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
	agent.endpoints = []string{"test.com/test1", "test.com/test2", "test.com/test3"}
	agent.err = nil
	err = testServiceMap.newService("testNewValidCase")
	if err != nil {
		t.Error("adding a valid test case returned error:", err)
	}
	_, present = testServiceMap["testNewValidCase"]
	if !present {
		t.Error("service 'testNewValidCase' wasn't added to testServiceMap map")
	}
}

func TestServiceEndpointsGetAndInc(t *testing.T) {
	testServiceEndpoint := &serviceEndpoints{
		endpoints: []string{"test.com/test1", "test.com/test2"},
		index:     0,
	}

	//test first increment on 2 endpoint service
	endpointFirst := testServiceEndpoint.getAndInc()
	if (endpointFirst != "test.com/test1") && (testServiceEndpoint.index != 1) {
		t.Error("getAndInc did not return the correct endpoint and increment after initializing")
	}

	//test second increment on 2 endpoint service
	endpointSecond := testServiceEndpoint.getAndInc()
	if (endpointSecond != "test.com/test2") && (testServiceEndpoint.index != 0) {
		t.Error("getAndInc did not return the correct endpoint and increment after a previos call")
	}

	//test third increment on 2 endpoint service, should loop back to first endpoint
	endpointBackToStart := testServiceEndpoint.getAndInc()
	if (endpointBackToStart != "test.com/test1") && (testServiceEndpoint.index != 1) {
		t.Error("getAndInc did not circle around to the first endpoint when at the emd on the endpoint slice")
	}

}

func TestServiceEndpointsRefresh(t *testing.T) {
	//test error from getHealthyEndpoints
	testServiceEndpoint := &serviceEndpoints{
		endpoints: []string{"test.com/test1", "test.com/test2"},
		index:     0,
	}
	agent.endpoints = []string{}
	agent.err = errors.New("get healthly endpoints failure")
	err := testServiceEndpoint.refresh()
	if err == nil {
		t.Error("no error returned despite getHealthyEndpoints returning an error")
	}

	//test index in range
	testServiceEndpoint = &serviceEndpoints{
		endpoints: []string{"test.com/test1", "test.com/test2"},
		index:     1,
	}
	agent.endpoints = []string{"test.com/test1", "test.com/test2"}
	agent.err = nil
	err = testServiceEndpoint.refresh()
	if err != nil {
		t.Error("refreshing a valid test case returned error:", err)
	}
	if testServiceEndpoint.index != 1 {
		t.Error("refresh did not return the same index when endpoint list length didn't change")
	}

	//test index out of range
	testServiceEndpoint = &serviceEndpoints{
		endpoints: []string{"test.com/test1", "test.com/test2"},
		index:     5,
	}
	agent.endpoints = []string{"test.com/test1", "test.com/test2"}
	agent.err = nil
	err = testServiceEndpoint.refresh()
	if err != nil {
		t.Error("refreshing a valid test case returned error:", err)
	}
	if testServiceEndpoint.index == 5 {
		t.Error("refresh did not return a new index when endpoint list length decreased")
	}
}

func TestServiceEndpointsTimedOut(t *testing.T) {
	//test timed out
	testServiceEndpoint := &serviceEndpoints{
		timeout: time.After(time.Millisecond),
	}
	time.Sleep(time.Millisecond * 5)
	if !testServiceEndpoint.timedOut() {
		t.Error("timeout should have returned true for a time out")
	}

	//test not timed out
	testServiceEndpoint = &serviceEndpoints{
		timeout: time.After(time.Second),
	}
	if testServiceEndpoint.timedOut() {
		t.Error("timeout should have returned false for a time out")
	}
}

type BehaviorTest struct {
	testName       string
	agentEndpoints []string
	agentError     error
	serviceName    string
	endpointOut    string
	sleep          bool
}

var testCases = []BehaviorTest{
	{testName: "new working",
		agentEndpoints: []string{"test.com/test1", "test.com/test2"},
		agentError:     nil,
		serviceName:    "test1",
		endpointOut:    "test.com/test1"},
	{testName: "new not working",
		agentEndpoints: []string{""},
		agentError:     errors.New("consul error"),
		serviceName:    "test2",
		endpointOut:    ""},
	{testName: "refresh not working",
		agentEndpoints: []string{""},
		agentError:     errors.New("consul error"),
		serviceName:    "test1",
		endpointOut:    "",
		sleep:          true},
	{testName: "refresh working",
		agentEndpoints: []string{"test.com/test1", "test.com/test2"},
		agentError:     nil,
		serviceName:    "test1",
		endpointOut:    "test.com/test2",
		sleep:          true},
	{testName: "new working",
		agentEndpoints: []string{"test.com/test1", "test.com/test2"},
		agentError:     nil,
		serviceName:    "test3",
		endpointOut:    "test.com/test1"},
	{testName: "don't refresh, working",
		agentEndpoints: []string{"test.com/test1", "test.com/test2"},
		agentError:     nil,
		serviceName:    "test3",
		endpointOut:    "test.com/test2"},
}

func TestGetServiceEndpoint(t *testing.T) {
	for i := range testCases {
		if testCases[i].sleep {
			time.Sleep(time.Second * 2)
		}
		agent.endpoints = testCases[i].agentEndpoints
		agent.err = testCases[i].agentError

		endpoint, err := GetServiceEndpoint(testCases[i].serviceName)

		if endpoint != testCases[i].endpointOut {
			t.Error(testCases[i].testName, " endpoint expected:", testCases[i].endpointOut, " endpoint recieved:", endpoint)
		}

		if err != testCases[i].agentError {
			t.Error(testCases[i].testName, err, testCases[i].agentError)
		}
	}
}
