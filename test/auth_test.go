/*
 * Tests for golinkedin.
 */

package golinkedin_test

import (
  "net/http"
  "net/http/httptest"
  "testing"
  "strings"
  "fmt"
  "io/ioutil"
  "github.com/stretchr/testify/suite"
  "github.com/stretchr/testify/assert"
  "github.com/jarcoal/httpmock"
  "github.com/FretboardLabs/golinkedin"
)

type GoLinkedInTestSuite struct {
  suite.Suite
  APIKey string
  APISecret string
  RedirectURL string
  Scope []string
  AccessToken string
}

/************
 * Tests
 ***********/

// Setup
func (suite *GoLinkedInTestSuite) SetupTest() {
  suite.APIKey = "api_key"
  suite.APISecret = "api_secret"
  suite.RedirectURL = "http://localhost"
  suite.Scope = []string{"scope_1", "scope_2"}
  suite.AccessToken = "abcd1234"

  golinkedin.Init(suite.APIKey, suite.APISecret, suite.RedirectURL, suite.Scope)
}

// Make sure starting the auth process results in a
// redirect to the LinkedIn auth page
func (suite *GoLinkedInTestSuite) TestStartAuth() {
  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    err := golinkedin.StartAuth(w, r)
    if err != nil {
      http.Error(w, err.Error(), 500)
    }
  }))
  defer ts.Close()

  req, err := http.NewRequest("GET", ts.URL, nil)
  assert.Equal(suite.T(), err, nil)

  transport := http.Transport{}
  res, err := transport.RoundTrip(req)
  assert.Equal(suite.T(), err, nil)

  // Make sure we received a redirect
  assert.Equal(suite.T(), res.StatusCode, 302)

  redirectLocation, err := res.Location()
  assert.Equal(suite.T(), err, nil)

  // Make sure we were redirected to the right place
  assert.Equal(suite.T(), redirectLocation.Host, "www.linkedin.com")
  assert.Equal(suite.T(), redirectLocation.Path, "/uas/oauth2/authorization")

  // Make sure the redirect URL had the right query parameters
  queryValues := redirectLocation.Query()
  assert.Equal(suite.T(), queryValues["response_type"][0], "code")
  assert.Equal(suite.T(), queryValues["client_id"][0], suite.APIKey)
  assert.Equal(suite.T(), queryValues["redirect_uri"][0], suite.RedirectURL)
  assert.Equal(suite.T(), queryValues["scope"][0], strings.Join(suite.Scope, " "))
}

// Test exchanging an authorization token for an access token works
func (suite *GoLinkedInTestSuite) TestCompleteAuth() {

  // Activate HTTP mocking, allowing non-mocked requests to go
  // through as usual
  httpmock.Activate()
  httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
  defer httpmock.DeactivateAndReset()

  // Mock our request to LinkedIn's access token endpoint
  jsonResponse, err := httpmock.NewJsonResponder(200, map[string]string{"access_token": suite.AccessToken})
  assert.Equal(suite.T(), err, nil)

  httpmock.RegisterResponder("POST", "https://www.linkedin.com/uas/oauth2/accessToken", jsonResponse)

  // Server for starting auth
  ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    golinkedin.StartAuth(w, r)
  }))
  defer ts1.Close()

  // Server for completing auth
  ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    res, err := golinkedin.CompleteAuth(w, r)
    if err != nil {
      http.Error(w, err.Error(), 500)
    }
    fmt.Fprintf(w, res)
  }))
  defer ts2.Close()

  // Hit the first server, so we can match the state given to us by LinkedIn
  // when the auth process is initiated
  req, _ := http.NewRequest("GET", ts1.URL, nil)
  transport := http.Transport{}
  res, _ := transport.RoundTrip(req)
  redirectLocation, _ := res.Location()
  queryValues := redirectLocation.Query()

  // Hit the second server to test the retrieval of an access token. Note that
  // this retrieval will go through our mocked LinkedIn endpoint.
  res, err = http.Get(ts2.URL + "?code=blah&state=" + queryValues["state"][0])
  assert.Equal(suite.T(), err, nil)
  assert.Equal(suite.T(), res.StatusCode, 200)

  body, err := ioutil.ReadAll(res.Body)
  assert.Equal(suite.T(), err, nil)
  assert.Equal(suite.T(), string(body), suite.AccessToken)
}

// Kick off all tests
func TestGoLinkedInTestSuite(t *testing.T) {
  suite.Run(t, new(GoLinkedInTestSuite))
}

/************
 * Helpers
 ***********/
func printf(suite *GoLinkedInTestSuite, format string, args ...interface{}) {
  suite.T().Logf(format, args)
}
