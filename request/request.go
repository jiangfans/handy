package request

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/jiangfans/handy/monitor"
	"github.com/jiangfans/handy/utils"

	log "github.com/sirupsen/logrus"
)

const (
	ContentTypeJson       = "application/json"
	ContentTypeUrlencoded = "application/x-www-form-urlencoded"
)

type (
	Request struct {
		ctx             context.Context
		client          *http.Client
		internalRequest *InternalRequest
		basicAuth       *BasicAuth
		method          string
		url             string
		queryParams     url.Values
		headers         map[string]string
		bodyBytes       []byte
		err             error
		prom            *Prom
	}

	InternalRequest struct {
		HmacKey string
		StoreId uint64
		Locale  string
	}

	BasicAuth struct {
		UserName string
		Password string
	}

	Prom struct {
		url string
	}
)

func New() *Request {
	r := &Request{
		headers:     make(map[string]string),
		queryParams: make(url.Values),
	}

	client := http.DefaultClient
	client.Timeout = 5 * time.Second
	r.client = client

	return r
}

func (r *Request) Method(method string) *Request {
	r.method = method
	return r
}

func (r *Request) Url(reqUrl string, params ...interface{}) *Request {
	reqUrl = fmt.Sprintf(reqUrl, params...)

	u, err := url.Parse(reqUrl)
	if err != nil {
		log.Error("parse url error: ", err.Error())
		r.err = errors.New("parse url error: " + err.Error())
		return r
	}

	if u.RawQuery != "" {
		values, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			log.Error("parse url query error: ", err.Error())
			r.err = errors.New("parse url query error: " + err.Error())
			return r
		}

		for key, _ := range values {
			r.queryParams.Set(key, values.Get(key))
		}

		r.url = u.String()
	} else {
		r.url = reqUrl
	}

	return r
}

func (r *Request) QueryParams(params map[string]interface{}) *Request {
	if len(params) == 0 {
		return r
	}

	for key, value := range params {
		if reflect.ValueOf(value).Kind() == reflect.Slice {
			s := reflect.ValueOf(value)

			key = key + "[]"
			for i := 0; i < s.Len(); i++ {
				r.queryParams.Add(key, fmt.Sprintf("%v", s.Index(i).Interface()))
			}
		}

		r.queryParams.Set(key, fmt.Sprintf("%v", value))
	}

	return r
}

func (r *Request) StructQueryParams(params interface{}) *Request {
	bs, err := json.Marshal(params)
	if err != nil {
		r.err = errors.New("incorrect struct query params format")
		return r
	}

	mapParams := make(map[string]interface{})
	err = json.Unmarshal(bs, &mapParams)
	if err != nil {
		r.err = errors.New("struct query params convert to map failed")
		return r
	}

	return r.QueryParams(mapParams)
}

func (r *Request) JsonBody(body interface{}) *Request {
	bs, err := json.Marshal(body)
	if err != nil {
		log.Error("incorrect body format: " + err.Error())
		r.err = errors.New("incorrect body format")
		return r
	}
	r.bodyBytes = bs

	r.ContentType(ContentTypeJson)
	return r
}

func (r *Request) UrlencodedFormatBody(body map[string]string) *Request {
	values := make(url.Values)
	for key, value := range body {
		values.Set(key, value)
	}

	if len(values) != 0 {
		bodyStr := values.Encode()
		r.bodyBytes = []byte(bodyStr)
	}

	r.ContentType(ContentTypeUrlencoded)
	return r
}

func (r *Request) BodyBytes(bs []byte) *Request {
	r.bodyBytes = bs
	return r
}

func (r *Request) AddHeader(key, value string) *Request {
	r.headers[key] = value
	return r
}

func (r *Request) AddHeaders(headers map[string]string) *Request {
	for key, value := range headers {
		r.AddHeader(key, value)
	}
	return r
}

func (r *Request) AccessToken(accessToken string) *Request {
	r.AddHeader("Access-Token", accessToken)
	return r
}

func (r *Request) ContentType(contentType string) *Request {
	r.AddHeader("Content-Type", contentType)
	return r
}

func (r *Request) InternalRequest(hmacKey string, storeId uint64, locale string) *Request {
	r.internalRequest = &InternalRequest{
		HmacKey: hmacKey,
		StoreId: storeId,
		Locale:  locale,
	}
	return r
}

func (r *Request) addInternalRequestHeaders() *Request {
	if r.internalRequest != nil {
		timestamp := time.Now().Unix()
		r.AddHeader("X-TimeStamp", strconv.FormatInt(timestamp, 10))

		if r.internalRequest.StoreId != 0 {
			r.AddHeader("store-id", strconv.FormatUint(r.internalRequest.StoreId, 10))
		}

		if r.internalRequest.Locale != "" {
			r.AddHeader("X-LOCALE", r.internalRequest.Locale)
		} else {
			r.AddHeader("X-LOCALE", utils.LanguageZhCN)
		}

		var (
			rawQuery string
			body     string
		)

		if len(r.queryParams) != 0 {
			rawQuery = r.queryParams.Encode()
		}
		if len(r.bodyBytes) != 0 {
			body = string(r.bodyBytes)
		}

		h := hmac.New(sha256.New, []byte(r.internalRequest.HmacKey))
		_, err := h.Write([]byte(fmt.Sprintf("timestamp:%d\n%s\n%s", timestamp, rawQuery, body)))
		if err != nil {
			log.Error("hmac count failed: " + err.Error())
			r.err = err
			return r
		}
		sum := h.Sum(nil)
		token := base64.StdEncoding.EncodeToString(sum)
		r.AddHeader("X-TOKEN", token)
	}

	return r
}

func (r *Request) Timeout(t time.Duration) *Request {
	if t != 0 {
		r.client.Timeout = t
	}

	return r
}

func (r *Request) BasicAuth(userName, password string) *Request {
	r.basicAuth = &BasicAuth{
		UserName: userName,
		Password: password,
	}
	return r
}

func (r *Request) EnableProm(reqUrl string) *Request {
	r.prom = &Prom{
		url: reqUrl,
	}
	return r
}

func (r *Request) Get(ctx ...context.Context) (respBs []byte, statusCode int, err error) {
	return r.Method(http.MethodGet).Do(ctx...)
}

func (r *Request) Post(ctx ...context.Context) (respBs []byte, statusCode int, err error) {
	return r.Method(http.MethodPost).Do(ctx...)
}

func (r *Request) Put(ctx ...context.Context) (respBs []byte, statusCode int, err error) {
	return r.Method(http.MethodPut).Do(ctx...)
}

func (r *Request) Patch(ctx ...context.Context) (respBs []byte, statusCode int, err error) {
	return r.Method(http.MethodPatch).Do(ctx...)
}

func (r *Request) Delete(ctx ...context.Context) (respBs []byte, statusCode int, err error) {
	return r.Method(http.MethodDelete).Do(ctx...)
}

func (r *Request) Do(ctx ...context.Context) (respBs []byte, statusCode int, err error) {
	var elapsed int
	var requestStart *time.Time

	defer func() {
		log.WithFields(log.Fields{
			"method":      r.method,
			"url":         r.url,
			"status_code": statusCode,
			"elapsed":     elapsed,
			"error":       err,
		}).Info()

		log.WithFields(log.Fields{"request_body": string(r.bodyBytes), "resp_data": string(respBs)}).Debug()

		if r.prom != nil && r.prom.url != "" {
			if err != nil {
				monitor.RequestErrorProm.Inc(r.prom.url, r.method)
			} else {
				monitor.RequestProm.Inc(r.prom.url, r.method, strconv.Itoa(statusCode))
				if requestStart != nil {
					monitor.RequestProm.HandleTime(*requestStart, r.prom.url, r.method)
				}
			}
		}
	}()

	err = r.err
	if err != nil {
		log.Error(err.Error())
		return
	}

	if len(r.queryParams) != 0 {
		r.url += fmt.Sprintf("?%s", r.queryParams.Encode())
	}

	var body io.Reader
	if len(r.bodyBytes) != 0 {
		body = bytes.NewReader(r.bodyBytes)
	}

	if len(ctx) != 0 {
		r.ctx = ctx[0]
	} else {
		r.ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(r.ctx, r.method, r.url, body)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if len(r.headers) != 0 {
		for key, value := range r.headers {
			req.Header.Set(key, value)
		}
	}

	if _, ok := r.headers["Content-Type"]; !ok {
		if r.method == http.MethodPost || r.method == http.MethodPut || r.method == http.MethodPatch {
			req.Header.Set("Content-Type", ContentTypeJson)
		}
	}

	if r.internalRequest != nil {
		r.addInternalRequestHeaders()
	}

	if r.basicAuth != nil {
		req.SetBasicAuth(r.basicAuth.UserName, r.basicAuth.Password)
	}

	timeStart := time.Now()
	requestStart = &timeStart

	resp, err := r.client.Do(req)
	if err != nil {
		log.WithError(err).Error()
		return
	}

	elapsed = time.Now().Nanosecond()/1e6 - timeStart.Nanosecond()/1e6

	statusCode = resp.StatusCode

	defer func() {
		if resp.Body != nil {
			err = resp.Body.Close()
			if err != nil {
				log.Error(err.Error())
			}
		}
	}()

	respBs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return
	}

	return
}
