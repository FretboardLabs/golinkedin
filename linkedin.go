/*
 * A Go wrapper for the LinkedIn API.
 */

package golinkedin

import (
  "math/rand"
  "net/url"
  "net/http"
  "strings"
  "errors"
)

type LinkedInAPI struct {
  apiKey      string
  apiSecret   string
  accessToken string
  callbackURL *url.URL
  scope       []string
}

var api *LinkedInAPI = nil

/***********************
 * Exports
 **********************/
func Init(apiKey string, apiSecret string, callbackURL string, scope []string) (err error) {
  parsedURL, err := url.Parse(callbackURL)
  if err != nil {
    return err
  }

  api = new(LinkedInAPI)
  api.apiKey = apiKey
  api.apiSecret = apiSecret
  api.callbackURL = parsedURL
  api.scope = scope

  return nil
}

func Auth(w http.ResponseWriter, r *http.Request) (err error) {
  if api == nil {
    return errors.New("API has not been initialized.")
  }

  state := randString(16)
  http.Redirect(w, r, "https://www.linkedin.com/uas/oauth2/authorization?response_type=code&client_id=" +
    api.apiKey + "&redirect_uri=http://localhost:7000&state=" +
    state + "&scope=" + strings.Join(api.scope, "%20"), http.StatusFound)

  return nil
}

/***********************
 * Helpers
 **********************/

var letterNumbers = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(length int) (str string) {
  b := make([]rune, length)
  for i := range b {
      b[i] = letterNumbers[rand.Intn(len(letterNumbers))]
  }
  return string(b)
}


