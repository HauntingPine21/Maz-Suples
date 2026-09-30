package tidb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Caller interface {
	Call(context.Context, string, string, map[string]any) (Response, error)
}

type Client struct {
	baseURL, appID, publicKey, privateKey string
	http                                  *http.Client
}

type Response struct {
	Type string `json:"type"`
	Data Data   `json:"data"`
}
type Data struct {
	Columns []Column         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Result  Result           `json:"result"`
}
type Column struct {
	Col, DataType string
	Nullable      bool
}
type Result struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RowCount  int    `json:"row_count"`
	RowAffect int    `json:"row_affect"`
	Limit     int    `json:"limit"`
}
type Error struct {
	HTTPStatus, Code int
	Message          string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Data Service: http=%d code=%d", e.HTTPStatus, e.Code)
}

func New(baseURL, appID, publicKey, privateKey string) *Client {
	return &Client{strings.TrimRight(baseURL, "/"), appID, publicKey, privateKey, &http.Client{Timeout: 20 * time.Second}}
}

func (c *Client) Call(ctx context.Context, method, endpoint string, params map[string]any) (Response, error) {
	if method == "" || endpoint == "" {
		return Response{}, errors.New("método y endpoint son obligatorios")
	}
	u := fmt.Sprintf("%s/api/v1beta/app/%s/endpoint/%s", c.baseURL, url.PathEscape(c.appID), strings.TrimLeft(endpoint, "/"))
	var body io.Reader
	if method == http.MethodGet || method == http.MethodDelete {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, fmt.Sprint(v))
		}
		if encoded := q.Encode(); encoded != "" {
			u += "?" + encoded
		}
	} else if params != nil {
		payload, err := json.Marshal(params)
		if err != nil {
			return Response{}, err
		}
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return Response{}, err
	}
	req.SetBasicAuth(c.publicKey, c.privateKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("Data Service no disponible: %w", err)
	}
	defer resp.Body.Close()
	limited := io.LimitReader(resp.Body, 4<<20)
	var result Response
	if err := json.NewDecoder(limited).Decode(&result); err != nil {
		return Response{}, fmt.Errorf("respuesta inválida de Data Service: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || result.Data.Result.Code != 200 {
		return Response{}, &Error{HTTPStatus: resp.StatusCode, Code: result.Data.Result.Code, Message: result.Data.Result.Message}
	}
	return result, nil
}
