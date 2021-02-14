package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/html"
	"golang.org/x/sync/errgroup"

	"github.com/AKovalevich/simplecrawler/internal/crawler"
)

const (
	DefaultServerPort = ":8080"
	DefaultLogLevel = "debug"
	DefaultParsingTimeout = 2 * time.Second
)

type App struct {
	logger *zap.Logger
	Config Config
}

type Config struct {
	Port 		string
	LogLevel 	string
}

type ParsingResult struct {
	URL     string `json:"url"`
	Result  string `json:"result"`
	Success bool   `json:"status"`
	Error   string `json:"error"`
}

func New(logger *zap.Logger, config Config) App {
	return App{
		logger: logger,
		Config: config,
	}
}

func (app *App) Run() {
	ctx, done := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		select {
		case sig := <-signalChannel:
			app.logger.Info("Received signal", zap.String("signal", sig.String()))
			done()
		case <-gctx.Done():
			app.logger.Info("Closing signal goroutine")
			return gctx.Err()
		}

		return nil
	})

	srv := &http.Server{
		Addr:    app.Config.Port,
		Handler: http.DefaultServeMux,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("httpServer error", zap.Error(err))
			done()
		}
	}()

	g.Go(func() error {
		<-ctx.Done()
		return srv.Shutdown(ctx)
	})

	http.HandleFunc("/crawler", crawlerHandler(app.logger))

	err := g.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			app.logger.Info("context was canceled")
		} else {
			app.logger.Error("received error", zap.String("error", err.Error()))
		}
	} else {
		app.logger.Info("Crawler stopped !")
	}
}

func crawlerHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urls, ok := r.URL.Query()["url"]
		if !ok || len(urls[0]) < 1 {
			_, err := fmt.Fprint(w, "Query param 'url' is missing")
			if err != nil {
				logger.Error(err.Error())
			}
			return
		}

		urls = filterUrls(urls)
		httpCrawler := crawler.New(logger)
		resultChan := make(chan ParsingResult, len(urls))

		wg := sync.WaitGroup{}
		for i, url := range urls {
			wg.Add(1)
			go func (url string, index int) {
				result, err := httpCrawler.Parse(url, parseFunction, DefaultParsingTimeout)
				errorMessage := ""
				if err != nil {
					errorMessage = err.Error()
					logger.Debug("parsing error", zap.String("URL", url), zap.Error(err))
				}

				parsingResult := ParsingResult{
					URL:     url,
					Result:  result,
					Success: err == nil,
					Error:   errorMessage,
				}

				resultChan <- parsingResult
				wg.Done()
			}(url, i)
		}

		var resultData []ParsingResult

	MainLoop:
		for {
			select {
			case parsingResult := <- resultChan:
				resultData = append(resultData, parsingResult)
				if len(urls) == len(resultData) {
					break MainLoop
				}
			case <-time.After(DefaultParsingTimeout):
				break MainLoop
			}
		}
		wg.Wait()
		close(resultChan)

		result, _ := json.Marshal(resultData)
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprint(w, string(result))
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

// Parsing function for url crawling.
func parseFunction(targetHTML []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(targetHTML))
	if err != nil {
		return "", err
	}

	foundedValue, _ := traverse(doc)
	return foundedValue, nil
}

func traverse(n *html.Node) (string, bool) {
	if n.Type == html.ElementNode && n.Data == "title" {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := traverse(c)
		if ok {
			return result, ok
		}
	}

	return "", false
}

// Ensures elements in a slice are unique.
func filterUrls(urls []string) []string {
	unique := make(map[string]bool, len(urls))
	us := make([]string, len(unique))
	for _, elem := range urls {
		if len(elem) != 0 {
			if !unique[elem] {
				us = append(us, elem)
				unique[elem] = true
			}
		}
	}

	return us
}
