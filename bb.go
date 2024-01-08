package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type timestamp struct {
	Hour string `json:"format_12_hour"`
}

type session struct {
	StartsAt timestamp `json:"starts_at"`
	Spaces   int       `json:"spaces"`
}

var errRedirect = errors.New("redirected")

func getSessions(date string) ([]session, error) {
	reqURL := fmt.Sprintf("%s/api/activities/venue/%s/activity/%s/times?date=%s",
		conf.Gym.URL,
		conf.Gym.Location,
		conf.Gym.Activity,
		date)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("better: could not create request: %w", err)
	}

	req.Header.Add("origin", conf.Gym.Origin)

	client := &http.Client{}
	// Return an error upon redirection.
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errRedirect
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("better: client.Do: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("better: Status: %s", res.Status)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("better: io.ReadAll: %w", err)
	}

	var response struct {
		Sessions []session `json:"data"`
	}

	err = json.Unmarshal(resBody, &response)
	if err != nil {
		return nil, fmt.Errorf("better: json.Unmarshal: %w", err)
	}

	return response.Sessions, nil
}
