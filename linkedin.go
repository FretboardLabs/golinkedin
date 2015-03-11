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
  "io/ioutil"
  "encoding/json"
)

type LinkedInAPI struct {
  apiKey      string
  apiSecret   string
  callbackURL string
  scope       []string
}

var api *LinkedInAPI = nil
var states = make(map[string]bool)

/***********************
 * Exports
 **********************/
func Init(apiKey string, apiSecret string, callbackURL string, scope []string) (err error) {
  _, err = url.Parse(callbackURL)
  if err != nil {
    return err
  }

  api = new(LinkedInAPI)
  api.apiKey = apiKey
  api.apiSecret = apiSecret
  api.callbackURL = callbackURL
  api.scope = scope

  return nil
}

func StartAuth(w http.ResponseWriter, r *http.Request) (err error) {
  if api == nil {
    return errors.New("API has not been initialized.")
  }

  state := randString(16)
  states[state] = true

  http.Redirect(w, r, "https://www.linkedin.com/uas/oauth2/authorization?response_type=code&client_id=" +
    api.apiKey + "&redirect_uri=" + api.callbackURL + "&state=" +
    state + "&scope=" + strings.Join(api.scope, "%20"), http.StatusFound)

  return nil
}

func CompleteAuth(w http.ResponseWriter, r *http.Request) (accessToken string, err error) {
  if api == nil {
    return "", errors.New("API has not been initialized.")
  }

  queryValues := r.URL.Query()
  if queryValues["state"] == nil || !states[queryValues["state"][0]] {
    return "", errors.New("State mismatch. Possible CSRF attack.")
  }
  states[queryValues["state"][0]] = false

  res, err := http.Post(
    "https://www.linkedin.com/uas/oauth2/accessToken?grant_type=authorization_code&code=" +
      queryValues["code"][0] + "&redirect_uri=" + api.callbackURL + "&client_id=" + api.apiKey +
      "&client_secret=" + api.apiSecret,
    "application/json",
    strings.NewReader(""))

  if err != nil {
    return "", err
  }

  body, err := ioutil.ReadAll(res.Body)
  var result map[string]interface{}
  err = json.Unmarshal(body, &result)
  accessToken, ok := result["access_token"].(string)
  if !ok {
    return "", errors.New("Unable to parse access_token from LinkedIn response.")
  }

  return accessToken, nil
}

func GetUser(w http.ResponseWriter, r *http.Request, accessToken string) (firstName string, lastName string, linkedinId string, err error) {
  resp, err := http.Get("https://api.linkedin.com/v1/people/~?oauth2_access_token=" + accessToken + "&format=json")
  if err != nil {
    return "", "", "", err
  }

  body, err := ioutil.ReadAll(resp.Body)
  var result map[string]interface{}
  err = json.Unmarshal(body, &result)
  if err != nil {
    return "", "", "", err
  }
  firstName, _ = result["firstName"].(string)
  lastName, _ = result["lastName"].(string)
  linkedinId, _ = result["id"].(string)
  if (len(firstName) == 0 || len(lastName) == 0 || len(linkedinId) == 0) {
    return "", "", "", errors.New("error missing fields in login response")
  } else {
    return firstName, lastName, linkedinId, nil
  }
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


