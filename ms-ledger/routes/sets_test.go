// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

func getDefaultProduct() Product {
	return Product{
		CreatedAt:          1567787309,
		IsActive:           true,
		ItemPrice:          1.99,
		MaxRestockingLevel: 24,
		MinRestockingLevel: 0,
		ProductName:        "Sprite (Lemon-Lime) - 16.9 oz",
		SKU:                "4900002470",
		UnitsOnHand:        0,
		UpdatedAt:          1567787309,
	}
}

func newInventoryTestServer(t *testing.T) *httptest.Server {

	inventoryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		expectedResponse := utilities.HTTPResponse{
			Content:     "",
			ContentType: "",
			StatusCode:  200,
			Error:       false,
		}

		// vars
		defaultProduct := getDefaultProduct()
		sku := r.RequestURI

		if sku == "/"+defaultProduct.SKU {
			w.WriteHeader(http.StatusOK)
			jsonProduct, _ := json.Marshal(defaultProduct)
			expectedResponse.Content = string(jsonProduct)
			jsonResponse, _ := json.Marshal(expectedResponse)
			_, err := w.Write(jsonResponse)
			if err != nil {
				t.Fatal(err.Error())
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("Could not find product for SKU"))
			if err != nil {
				t.Fatal(err.Error())
			}
		}
	}))

	return inventoryServer
}

func TestLedgerAddTransaction(t *testing.T) {
	// Use community-recommended shorthand (known name clash)
	require := require.New(t)

	// Accounts slice
	accountLedgers := getDefaultAccountLedgers()

	appSettings := make(map[string]string)

	inventoryServer := newInventoryTestServer(t)

	appSettings["InventoryEndpoint"] = inventoryServer.URL

	tests := []struct {
		Name               string
		InvalidLedger      bool
		UpdateLedger       string
		ExpectedStatusCode int
	}{
		{"Valid SKU and accountID", false, `{"accountId":2,"deltaSKUs":[{"sku":"4900002470","delta":-1}]}`, http.StatusOK},
		{"Incorrect type for accountID", false, `{"accountId":"2","deltaSKUs":[{"sku":"4900002470","delta":-1}]}`, http.StatusBadRequest},
		{"Nonexistent accountID", false, `{"accountId":10,"deltaSKUs":[{"sku":"4900002470","delta":-1}]}`, http.StatusBadRequest},
		{"bad data for SKU", false, `{"accountId":2,"deltaSKUs":[{"sku":"badSKU","delta":-1}]}`, http.StatusBadRequest},
		{"Nonexistent SKU in inventory", false, `{"accountId":2,"deltaSKUs":[{"sku":"4900002479","delta":-1}]}`, http.StatusBadRequest},
		{"Invalid Ledger", true, `{"accountId":2,"deltaSKUs":[{"sku":"4900002470","delta":-1}]}`, http.StatusInternalServerError},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := DeleteAllLedgers()
			require.NoError(err)
			if currentTest.InvalidLedger {
				err = ioutil.WriteFile(LedgerFileName, []byte("invalid json test"), 0644)
			} else {
				err = utilities.WriteToJSONFile(LedgerFileName, &accountLedgers, 0644)
			}
			require.NoError(err)

			req := httptest.NewRequest("POST", "http://localhost:48093/ledger", bytes.NewBuffer([]byte(currentTest.UpdateLedger)))
			w := httptest.NewRecorder()
			req.Header.Set("Content-Type", "application/json")
			ctx := req.Context()
			ctx = context.WithValue(ctx, AppSettingsKey, appSettings)
			req = req.WithContext(ctx)
			req.Context()
			LedgerAddTransaction(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}

func TestGetInventoryItemInfo(t *testing.T) {

	appSettings := make(map[string]string)
	// Default variables
	defaultProduct := getDefaultProduct()
	defaultSKU := "4900002470"

	inventoryServer := newInventoryTestServer(t)

	tests := []struct {
		Name              string
		MissingAppSetting bool
		InventoryEndpoint string
		SKU               string
		ProductMatch      bool
		Error             bool
	}{
		{"Valid SKU", false, inventoryServer.URL, defaultSKU, true, false},
		{"Nonexistent SKU", false, inventoryServer.URL, "123", false, true},
		{"Missing AppSetting", true, inventoryServer.URL, defaultSKU, false, true},
		{"Invalid InventoryEndpoint", false, "badURL", defaultSKU, false, true},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			if currentTest.MissingAppSetting {
				badAppSettings := make(map[string]string)
				_, err := getInventoryItemInfo(badAppSettings, currentTest.SKU)
				require.Error(t, err)
				return
			}

			appSettings["InventoryEndpoint"] = currentTest.InventoryEndpoint
			inventoryItem, err := getInventoryItemInfo(appSettings, currentTest.SKU)
			if currentTest.Error {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if currentTest.ProductMatch {
				assert.Equal(t, defaultProduct, inventoryItem, "Products should match")
			}
		})
	}
}

func TestSetPaymentStatus(t *testing.T) {
	// Use community-recommended shorthand (known name clash)
	require := require.New(t)

	// Accounts slice
	accountLedgers := getDefaultAccountLedgers()

	tests := []struct {
		Name               string
		InvalidLedger      bool
		PaymentInfo        string
		ExpectedStatusCode int
	}{
		{"Valid Payment Info", false, `{"accountId":1,"transactionID":"1579215712984890248","isPaid": true }`, http.StatusOK},
		{"Nonexistent accountID", false, `{"accountId":10,"transactionID":"1579215712984890248","isPaid": true }`, http.StatusBadRequest},
		{"Nonexistent transactionID", false, `{"accountId":1,"transactionID":"1579215712984890249","isPaid": true }`, http.StatusBadRequest},
		{"Bad data in Payment Info", false, `{"accountId":1,"transactionID":"improperFormat","isPaid": true }`, http.StatusBadRequest},
		{"Invalid ledger", true, `{"accountId":1,"transactionID":"1579215712984890248","isPaid": true }`, http.StatusInternalServerError},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := DeleteAllLedgers()
			require.NoError(err)
			if currentTest.InvalidLedger {
				err = ioutil.WriteFile(LedgerFileName, []byte("invalid json test"), 0644)
			} else {
				err = utilities.WriteToJSONFile(LedgerFileName, &accountLedgers, 0644)
			}
			require.NoError(err)

			req := httptest.NewRequest("POST", "http://localhost:48093/ledger/ledgerPaymentUpdate", bytes.NewBuffer([]byte(currentTest.PaymentInfo)))
			w := httptest.NewRecorder()
			SetPaymentStatus(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}
