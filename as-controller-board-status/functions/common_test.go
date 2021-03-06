// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

// These constants are used to simplify code repetition when working with the
// config map/struct
const (
	AverageTemperatureMeasurementDuration     = "AverageTemperatureMeasurementDuration"
	DeviceName                                = "DeviceName"
	MaxTemperatureThreshold                   = "MaxTemperatureThreshold"
	MinTemperatureThreshold                   = "MinTemperatureThreshold"
	MQTTEndpoint                              = "MQTTEndpoint"
	NotificationCategory                      = "NotificationCategory"
	NotificationEmailAddresses                = "NotificationEmailAddresses"
	NotificationHost                          = "NotificationHost"
	NotificationLabels                        = "NotificationLabels"
	NotificationReceiver                      = "NotificationReceiver"
	NotificationSender                        = "NotificationSender"
	NotificationSeverity                      = "NotificationSeverity"
	NotificationSlug                          = "NotificationSlug"
	NotificationSlugPrefix                    = "NotificationSlugPrefix"
	NotificationSubscriptionMaxRESTRetries    = "NotificationSubscriptionMaxRESTRetries"
	NotificationSubscriptionRESTRetryInterval = "NotificationSubscriptionRESTRetryInterval"
	NotificationThrottleDuration              = "NotificationThrottleDuration"
	RESTCommandTimeout                        = "RESTCommandTimeout"
	SubscriptionHost                          = "SubscriptionHost"
	VendingEndpoint                           = "VendingEndpoint"
)

// GetCommonSuccessConfig is used in test cases to quickly build out
// an example of a successful ControllerBoardStatusAppSettings configuration
func GetCommonSuccessConfig() ControllerBoardStatusAppSettings {
	return ControllerBoardStatusAppSettings{
		AverageTemperatureMeasurementDuration:     -15 * time.Second,
		DeviceName:                                "ds-controller-board",
		MaxTemperatureThreshold:                   83.0,
		MinTemperatureThreshold:                   10.0,
		MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
		NotificationCategory:                      "HW_HEALTH",
		NotificationEmailAddresses:                []string{"test@site.com", "test@site.com"},
		NotificationHost:                          "http://localhost:48060/api/v1/notification",
		NotificationLabels:                        []string{"HW_HEALTH"},
		NotificationReceiver:                      "System Administrator",
		NotificationSender:                        "Automated Checkout Maintenance Notification",
		NotificationSeverity:                      "CRITICAL",
		NotificationSlug:                          "sys-admin",
		NotificationSlugPrefix:                    "maintenance-notification",
		NotificationSubscriptionMaxRESTRetries:    10,
		NotificationSubscriptionRESTRetryInterval: 10 * time.Second,
		NotificationThrottleDuration:              1 * time.Minute,
		RESTCommandTimeout:                        15 * time.Second,
		SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
		VendingEndpoint:                           "http://localhost:48099/boardStatus",
	}
}

// GetHTTPTestServer returns a basic HTTP test server that does nothing more than respond with
// a desired status code
func GetHTTPTestServer(statusCodeResponse int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCodeResponse)
		_, err := w.Write([]byte(response))
		if err != nil {
			panic(err)
		}

	}))
}

// GetErrorHTTPTestServer returns a basic HTTP test server that produces a guaranteed error condition
// by simply closing client connections
func GetErrorHTTPTestServer() *httptest.Server {
	var testServerThrowError *httptest.Server
	testServerThrowError = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testServerThrowError.CloseClientConnections()
	}))
	return testServerThrowError
}

type testTableRESTCommandJSONStruct struct {
	TestCaseName    string
	Config          ControllerBoardStatusAppSettings
	InputRESTMethod string
	InputInterface  interface{}
	HTTPTestServer  *httptest.Server
	Output          error
}

func prepRESTCommandJSONTest() ([]testTableRESTCommandJSONStruct, []*httptest.Server) {
	output := []testTableRESTCommandJSONStruct{}

	// This server returns 200 OK
	testServerStatusOK := GetHTTPTestServer(http.StatusOK, "")

	// This server throws HTTP 500 as part of a non-error response
	testServer500 := GetHTTPTestServer(http.StatusInternalServerError, "test response body")

	// This server throws errors when it receives a connection
	testServerThrowError := GetErrorHTTPTestServer()

	edgexconfig := ControllerBoardStatusAppSettings{
		RESTCommandTimeout: time.Second * 15,
	}

	invalidRestMethod := "invalid rest method"

	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:    "Success GET",
			Config:          edgexconfig,
			InputRESTMethod: RESTGet,
			InputInterface:  "",
			HTTPTestServer:  testServerStatusOK,
			Output:          nil,
		})

	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:    "Success POST",
			Config:          edgexconfig,
			InputRESTMethod: RESTPost,
			InputInterface:  "simple test string",
			HTTPTestServer:  testServerStatusOK,
			Output:          nil,
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:    "Success PUT",
			Config:          edgexconfig,
			InputRESTMethod: RESTPut,
			InputInterface:  "simple test string",
			HTTPTestServer:  testServerStatusOK,
			Output:          nil,
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:    "Unsuccessful GET due to undesired status code",
			Config:          edgexconfig,
			InputRESTMethod: RESTGet,
			InputInterface:  "",
			HTTPTestServer:  testServer500,
			Output:          fmt.Errorf("Did not receive an HTTP 200 status OK response from %v, instead got a response code of %v, and the response body was: %v", testServer500.URL, http.StatusInternalServerError, "test response body"),
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:    "Unsuccessful GET due to connection closure",
			Config:          edgexconfig,
			InputRESTMethod: RESTGet,
			InputInterface:  "",
			HTTPTestServer:  testServerThrowError,
			Output:          fmt.Errorf("Failed to submit REST %v request due to error: %v \"%v\": %v", RESTGet, "Get", testServerThrowError.URL, "EOF"),
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:    "Unsuccessful GET due to unserializable JSON input",
			Config:          edgexconfig,
			InputRESTMethod: RESTGet,
			InputInterface: map[string](chan bool){
				"test": make(chan bool),
			},
			HTTPTestServer: testServerStatusOK,
			Output:         fmt.Errorf("Failed to serialize the input interface as JSON: Failed to marshal into JSON string: json: unsupported type: chan bool"),
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:    "Unsuccessful call due to invalid REST Method",
			Config:          edgexconfig,
			InputRESTMethod: invalidRestMethod,
			InputInterface:  "",
			HTTPTestServer:  testServerStatusOK,
			Output:          fmt.Errorf("Failed to build the REST %v request for the URL %v due to error: net/http: invalid method \"%v\"", invalidRestMethod, testServerStatusOK.URL, invalidRestMethod), // https://github.com/golang/go/blob/7d2473dc81c659fba3f3b83bc6e93ca5fe37a898/src/net/http/request.go#L846
		})
	return output, []*httptest.Server{
		testServerStatusOK,
		testServer500,
		testServerThrowError,
	}
}

// TestRESTCommandJSON validates that the RESTCommandJSON works in all
// possible error conditions & success conditions
func TestRESTCommandJSON(t *testing.T) {
	testTable, testServers := prepRESTCommandJSONTest()
	// We are responsible for closing the test servers
	for _, testServer := range testServers {
		defer testServer.Close()
	}

	for _, testCase := range testTable {
		ct := testCase // pinning to avoid concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			assert := assert.New(t)
			err := testCase.Config.RESTCommandJSON(testCase.HTTPTestServer.URL, testCase.InputRESTMethod, testCase.InputInterface)
			assert.Equal(ct.Output, err, "Expected output to be the same")
		})
	}

}
