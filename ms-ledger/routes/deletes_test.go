// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

func TestLedgerDelete(t *testing.T) {
	// Use community-recommended shorthand (known name clash)
	require := require.New(t)
	assert := assert.New(t)

	// Accounts slice
	accountLedgers := getDefaultAccountLedgers()
	// Default variables
	defaultAccountID := "1"
	InvalidAccountID := "10"
	defaultTransactionID := "1579215712984890248"
	InvalidTransactionID := "1579215712984890249"

	tests := []struct {
		Name               string
		InvalidLedger      bool
		AccountID          string
		TransactionID      string
		TransactionDeleted bool
		ExpectedStatusCode int
	}{
		{"Valid AccountID and TransactionID", false, defaultAccountID, defaultTransactionID, true, http.StatusOK},
		{"Bad data AccountID", false, "badformat", defaultTransactionID, false, http.StatusBadRequest},
		{"Nonexistent AccountID", false, InvalidAccountID, defaultTransactionID, false, http.StatusBadRequest},
		{"Bad data TransactionID", false, defaultAccountID, "badformat", false, http.StatusBadRequest},
		{"Nonexistent TransactionID", false, defaultAccountID, InvalidTransactionID, false, http.StatusBadRequest},
		{"Invalid Ledger Endpoint", true, defaultAccountID, defaultTransactionID, false, http.StatusInternalServerError},
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

			req := httptest.NewRequest("DELETE", "http://localhost:48093/ledger", bytes.NewBuffer([]byte(currentTest.AccountID+"/"+currentTest.TransactionID)))
			w := httptest.NewRecorder()

			URLVars := map[string]string{
				"accountid": currentTest.AccountID,
				"tid":       currentTest.TransactionID,
			}

			req = mux.SetURLVars(req, URLVars)
			req.Header.Set("Content-Type", "application/json")
			LedgerDelete(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")

			if !currentTest.InvalidLedger {
				// run GetAllLedgers and get the result as JSON
				accountsFromFile, err := GetAllLedgers()
				require.NoError(err)

				if currentTest.TransactionDeleted {
					assert.NotEqual(accountLedgers, accountsFromFile, "Ledgers should not match")
				} else {
					assert.Equal(accountLedgers, accountsFromFile, "Ledgers should match")
				}
			}
		})
	}
}
