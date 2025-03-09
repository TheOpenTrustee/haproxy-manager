package dataplane_client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/haproxytech/client-native/v6/models"
)

// function that gets a list of frontends from the haproxy dataplane api via http at port 5555 v3
// use https://www.haproxy.com/documentation/dataplaneapi/community/?v=v3 as a reference
func GetFrontends() (models.Frontends, *models.Error, error) {
	// setup the connection to the api
	req, err := http.NewRequest("GET", "http://localhost:5555/v3/services/haproxy/configuration/frontends", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	b, e, err := doRequestHandlingErrors(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error handling request: %v", err)
	}

	if e != nil {
		return nil, e, nil
	}

	var frontends models.Frontends
	err = json.Unmarshal(b, &frontends)
	if err != nil {
		return nil, nil, fmt.Errorf("failed unmarshalling frontends: %v", e)
	}

	return frontends, nil, nil
}

func doRequestHandlingErrors(req *http.Request) ([]byte, *models.Error, error) {
	req.SetBasicAuth("static", "static")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var err models.Error
		e := err.UnmarshalJSON(b)
		if e != nil {
			return nil, nil, fmt.Errorf("failed unmarshalling error: %v", e)
		}
		return nil, &err, nil
	}

	return b, nil, nil
}
