package audit

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/noedaka/go-url-shortener/internal/model"
)

type HTTPObserver struct {
    url    string
    client *http.Client
}

func NewHTTPObserver(url string) *HTTPObserver {
    return &HTTPObserver{
        url: url,
        client: &http.Client{
            Timeout: 5 * time.Second,
        },
    }
}

func (o *HTTPObserver) Notify(event model.AuditEvent) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    resp, err := o.client.Post(o.url, "application/json", bytes.NewReader(data))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return ErrHTTPRequestFailed
    }

    return nil
}

func (o *HTTPObserver) Close() error {
    // HTTP клиент не требует закрытия
    return nil
}

var ErrHTTPRequestFailed = errors.New("HTTP request failed")