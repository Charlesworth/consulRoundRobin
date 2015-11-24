package goConsulRoundRobin

import (
	"errors"
	"testing"
	"time"
)

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
	testAgent.endpoints = []string{}
	testAgent.err = errors.New("get healthly endpoints failure")
	err := testServiceEndpoint.refresh()
	if err == nil {
		t.Error("no error returned despite getHealthyEndpoints returning an error")
	}

	//test index in range
	testServiceEndpoint = &serviceEndpoints{
		endpoints: []string{"test.com/test1", "test.com/test2"},
		index:     1,
	}
	testAgent.endpoints = []string{"test.com/test1", "test.com/test2"}
	testAgent.err = nil
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
	testAgent.endpoints = []string{"test.com/test1", "test.com/test2"}
	testAgent.err = nil
	err = testServiceEndpoint.refresh()
	if err != nil {
		t.Error("refreshing a valid test case returned error:", err)
	}
	if testServiceEndpoint.index == 5 {
		t.Error("refresh did not return a new index when endpoint list length decreased")
	}

	//test check endpoint have changed with what is being returned from consul
	testServiceEndpoint = &serviceEndpoints{
		endpoints: []string{"test.com/oldName", "test.com/oldName"},
		index:     0,
	}
	testAgent.endpoints = []string{"test.com/newName", "test.com/newName"}
	testAgent.err = nil
	err = testServiceEndpoint.refresh()
	if err != nil {
		t.Error("refreshing a valid test case returned error:", err)
	}
	if testServiceEndpoint.endpoints[0] != "test.com/newName" {
		t.Error("refresh did not insert the new endpoint list returned from the consul agent",
			"expected: ", testAgent.endpoints, ", got: ", testServiceEndpoint.endpoints)
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
