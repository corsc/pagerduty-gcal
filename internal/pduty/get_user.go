package pduty

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// UserAPI contains the functions to call the User API
type UserAPI struct{}

// GetUsers will returns the mapping between PD user id and email
func (u *UserAPI) GetUsers(apiKey string, entries []*ScheduleEntry) (map[string]string, error) {
	out := map[string]string{}

	for _, entry := range entries {
		result, err := u.getUserEmail(apiKey, entry.User.ID)
		if err != nil {
			return nil, err
		}

		out[entry.User.ID] = result
	}

	return out, nil
}

func (u *UserAPI) getUserEmail(apiKey string, userID string) (string, error) {
	req, err := u.buildRequest(apiKey, userID)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	decoder := json.NewDecoder(resp.Body)

	apiResp := &apiResponse{}
	err = decoder.Decode(apiResp)
	if err != nil {
		return "", err
	}

	return apiResp.UserOuter.Email, nil
}

func (u *UserAPI) buildRequest(apiKey string, userID string) (*http.Request, error) {
	params := &url.Values{}
	params.Set("id", userID)

	req, err := http.NewRequest("GET", apiBaseURL+"/users/"+userID+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token token="+apiKey)
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	return req, nil
}

type userOuter struct {
	Email string
}
