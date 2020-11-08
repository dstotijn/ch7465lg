package ch7465lg

import (
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/publicsuffix"
)

const getterPath = "/xml/getter.xml"
const setterPath = "/xml/setter.xml"

// ErrLoginFailed is returned when login failed, e.g. invalid credentials.
var ErrLoginFailed = errors.New("login failed")

var sidRe = regexp.MustCompile(`^successful;SID=(\d+)$`)

// Client represents a client for a CH7465LG modem.
type Client struct {
	baseURL  *url.URL
	password string
	c        *http.Client
	mu       sync.Mutex
}

// FormValue is used for submitting form values.
type FormValue struct {
	Key   string
	Value string
}

// FormValues is used for submitting forms. Order matters, thus we cannot use a
// map (e.g. url.Values).
type FormValues []FormValue

// NewClient returns a new Client.
func NewClient(addr, password string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	if httpClient.Jar == nil {
		jar, err := cookiejar.New(&cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		})
		if err != nil {
			return nil, err
		}
		httpClient.Jar = jar
	}

	u, err := url.Parse("http://" + addr)
	if err != nil {
		log.Fatal(err)
	}

	client := &Client{
		baseURL:  u,
		password: password,
		c:        httpClient,
	}

	return client, nil
}

// Get retrieves settings.
func (c *Client) Get(funcNum int) (*http.Response, error) {
	sessionToken, err := c.sessionToken()
	if err != nil {
		return nil, err
	}

	formValues := FormValues{}
	formValues.Add("token", sessionToken)
	formValues.Add("fun", strconv.Itoa(funcNum))
	body := formValues.Encode()

	resp, err := c.c.Post(c.baseURL.String()+getterPath,
		"application/x-www-form-urlencoded", strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Set sends a mutation request.
func (c *Client) Set(funcNum int, formValues FormValues) (*http.Response, error) {
	sessionToken, err := c.sessionToken()
	if err != nil {
		return nil, err
	}

	if formValues == nil {
		formValues = FormValues{}
	}

	formValues = append(FormValues{
		FormValue{Key: "token", Value: sessionToken},
		FormValue{Key: "fun", Value: strconv.Itoa(funcNum)},
	}, formValues...)

	body := formValues.Encode()

	resp, err := c.c.Post(c.baseURL.String()+setterPath,
		"application/x-www-form-urlencoded", strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) sessionToken() (string, error) {
	cookies := c.c.Jar.Cookies(c.baseURL)
	if len(cookies) == 0 || cookies[0].Name != "sessionToken" {
		return "", errors.New("sessionToken cookie not found")
	}
	return cookies[0].Value, nil

}

// Add adds a form value.
func (v *FormValues) Add(key, value string) {
	*v = append(*v, FormValue{Key: key, Value: value})
}

// Encode encodes the form values into ``URL encoded'' form ("bar=baz&foo=quux").
func (v FormValues) Encode() string {
	if v == nil {
		return ""
	}

	var buf strings.Builder

	for _, vs := range v {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(vs.Key))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(vs.Value))
	}

	return buf.String()
}
