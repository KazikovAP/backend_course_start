package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	defaultTimeout  = 15
	exitCodeTimeout = 228
)

const helpText = `Usage: hedgedcurl [OPTIONS] URL [URL...]

CLI утилита многопоточного curl'а с хеджированием запросов.

Options:
  -t, --timeout SECONDS  Таймаут для HTTP запросов в секундах (по умолчанию: 15)
  -h, --help             Показать эту справку

Examples:
  hedgedcurl https://example.com https://example.org
  hedgedcurl -t 5 https://httpbin.org/delay/1 https://httpbin.org/delay/10
  hedgedcurl --help`

// config хранит параметры запуска утилиты.
type config struct {
	timeout time.Duration
	urls    []string
}

// requestResult содержит результат одного HTTP запроса.
type requestResult struct {
	resp *http.Response
	err  error
	url  string
}

// timeoutError — sentinel-тип для ошибки таймаута.
type timeoutError struct {
	cause error
}

func (e *timeoutError) Error() string {
	return fmt.Sprintf("все запросы завершились по таймауту: %v", e.cause)
}

func (e *timeoutError) Unwrap() error {
	return e.cause
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		fmt.Fprintln(os.Stderr, helpText)
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		var timeoutErr *timeoutError
		if errors.As(err, &timeoutErr) {
			fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
			os.Exit(exitCodeTimeout)
		}

		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

// parseArgs разбирает аргументы командной строки и возвращает конфиг.
func parseArgs(args []string) (*config, error) {
	cfg := &config{
		timeout: defaultTimeout * time.Second,
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			fmt.Println(helpText)
			os.Exit(0)

		case "-t", "--timeout":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("флаг %s требует значения", args[i])
			}
			i++

			seconds, err := strconv.Atoi(args[i])
			if err != nil || seconds <= 0 {
				return nil, fmt.Errorf("таймаут должен быть положительным числом, получено: %q", args[i])
			}

			cfg.timeout = time.Duration(seconds) * time.Second

		default:
			if _, err := url.ParseRequestURI(args[i]); err != nil {
				return nil, fmt.Errorf("некорректный URL %q: %w", args[i], err)
			}

			cfg.urls = append(cfg.urls, args[i])
		}
	}

	if len(cfg.urls) == 0 {
		return nil, fmt.Errorf("необходимо указать хотя бы один URL")
	}

	return cfg, nil
}

// run выполняет хеджированные запросы и печатает первый успешный ответ.
func run(cfg *config) error {
	client := &http.Client{Timeout: cfg.timeout}
	results := make(chan requestResult, len(cfg.urls))

	for _, u := range cfg.urls {
		go func(u string) {
			resp, err := client.Get(u)
			results <- requestResult{resp: resp, err: err, url: u}
		}(u)
	}

	var (
		lastErr   error
		timedOut  bool
		failCount int
	)

	for range cfg.urls {
		res := <-results

		if res.err != nil {
			failCount++
			lastErr = res.err
			if os.IsTimeout(res.err) {
				timedOut = true
			}

			if failCount == len(cfg.urls) {
				if timedOut {
					return &timeoutError{lastErr}
				}

				return fmt.Errorf("все запросы завершились с ошибкой, последняя (%s): %w", res.url, lastErr)
			}

			continue
		}

		return printResponse(res.resp)
	}

	return nil
}

// printResponse выводит заголовки и тело HTTP ответа.
func printResponse(resp *http.Response) error {
	defer resp.Body.Close()

	fmt.Printf("HTTP/1.1 %s\n", resp.Status)
	for key, values := range resp.Header {
		for _, v := range values {
			fmt.Printf("%s: %s\n", key, v)
		}
	}
	fmt.Println()

	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		return fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}

	return nil
}
