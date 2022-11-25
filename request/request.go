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
	"handy/utils"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

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
		promUrl         string
		queryParams     url.Values
		headers         map[string]string
		bodyBytes       []byte
		err             error
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
	r.promUrl = reqUrl
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

		u.RawQuery = ""
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

func (r *Request) Body(body interface{}) *Request {
	bs, err := json.Marshal(body)
	if err != nil {
		log.Error("incorrect body format: " + err.Error())
		r.err = errors.New("incorrect body format")
		return r
	}
	r.bodyBytes = bs

	return r
}

func (r *Request) AddHeader(key, value string) *Request {
	r.headers[key] = value
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

func (r *Request) InternalRequest(hmacKey string, storeID uint64, locale string) *Request {
	r.internalRequest = &InternalRequest{
		HmacKey: hmacKey,
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
	err = r.err
	if err != nil {
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
		log.Error("new request failed: " + err.Error())
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

	log.Infof("request %s, method: %s", r.url, r.method)
	log.Debug("request body: ", string(r.bodyBytes))

	resp, err := r.client.Do(req)
	if err != nil {
		log.Error("send request error: " + err.Error())
		return
	}

	statusCode = resp.StatusCode
	log.Infof("request %s, return code: %d", r.url, statusCode)

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
		log.Error("read body content error: " + err.Error())
		return
	}

	log.Debug("resp: ", string(respBs))

	return
}