package proxy

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

var proxyURLRaw string
var proxyToken string

// Init configures the proxy for use elsewhere
func Init(rawURL, token string) error {
	proxyURLRaw = rawURL
	if proxyURLRaw == "" {
		log.Fatal("rawURL for proxy is not set")
		os.Exit(1)
	}

	var err error
	_, err = url.Parse(proxyURLRaw)
	if err != nil {
		return fmt.Errorf("failed to parse proxy url (%s): %s", proxyURLRaw, err)
	}

	proxyToken = token
	if proxyToken == "" {
		return fmt.Errorf("token for proxy is not set")
	}

	return nil
}

// GetURLViaProxy will make a get request using a simple-proxy endpoint
// and forward the result
func GetURLViaProxy(requestURL string, headers map[string]string) (int, []byte, error) {
	log.Printf("fetching via proxy: %s", requestURL)
	response := &http.Response{}

	// need a fresh copy of the url since we change the RawQuery each request
	proxyURL, err := url.Parse(proxyURLRaw)
	if err != nil {
		return 0, []byte{}, errors.Wrap(err, "failed to parse raw proxy url")
	}

	q := proxyURL.Query()
	q.Add("url", requestURL)
	proxyURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", proxyURL.String(), nil)
	if err != nil {
		return 0, []byte{}, errors.Wrap(err, "failed to create proxy request")
	}

	client := &http.Client{}

	req.Header.Add("Authorization", "bearer "+proxyToken)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	response, err = client.Do(req)
	if err != nil {
		return 0, []byte{}, errors.Wrap(err, "failed to get via proxy")
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	return response.StatusCode, body, nil
}
