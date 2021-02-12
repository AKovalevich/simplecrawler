package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Crawler struct {
	logger *zap.Logger
}

type ParsingFunc func([]byte) (string, error)

type TargetURL struct {
	URL string
	// Allows you to override the HTML page parsing method for each URL.
	ParsingFunc ParsingFunc
}

func New(logger *zap.Logger) Crawler {
	return Crawler{
		logger: logger,
	}
}

// For start URL parsing. For each URL, with the given way to parse the HTML body.
// Will executed synchronously, and is blocked by I/O. For concurrent execution, it can be performed in a separate goroutine.
func (crawler *Crawler) Parse(url string, parsingFunc ParsingFunc, timeout time.Duration) (string, error) {
	body, err := crawler.fetch(url, timeout)
	if err != nil {
		return "", err
	}

	return crawler.parse(body, parsingFunc)
}

func (crawler *Crawler) fetch(url string, timeout time.Duration) ([]byte, error) {
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error with status %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	return body, nil
}

func (crawler *Crawler) parse(body []byte, parsingFunc ParsingFunc) (string, error) {
	return parsingFunc(body)
}
