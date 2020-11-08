package ch7465lg

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

const (
	downstreamsFunc = 10
	loginFunc       = 15
	logoutFunc      = 16
)

// Downstream represents a bonded downstream channel.
type Downstream struct {
	Frequency    int64   `xml:"freq"`
	Power        int     `xml:"pow"`
	SNR          int     `xml:"snr"`
	Modulation   string  `xml:"mod"`
	ChannelID    int     `xml:"chid"`
	RxMER        float64 `xml:"RxMER"`
	PreRSErrs    int64   `xml:"PreRs"`
	PostRSErrs   int64   `xml:"PostRs"`
	IsQamLocked  int     `xml:"IsQamLocked"`
	IsFECLocked  int     `xml:"IsFECLocked"`
	IsMpegLocked int     `xml:"IsMpegLocked"`
}

// Login performs a login attempt.
func (c *Client) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	formValues := FormValues{}
	formValues.Add("Username", "NULL")
	formValues.Add("Password", c.password)

	// Load homepage to get initial session token cookie (is stored in cookie jar).
	resp, err := c.c.Get(c.baseURL.String())
	if err != nil {
		return err
	}
	// TODO: Check if we're already logged in.
	resp.Body.Close()

	resp, err = c.Set(loginFunc, formValues)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	matches := sidRe.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return ErrLoginFailed
	}
	sid := matches[1]

	sidCookie := &http.Cookie{
		Domain: c.baseURL.Host,
		Name:   "SID",
		Value:  sid,
	}

	// Store SID cookie in cookie jar.
	updated := false
	cookies := resp.Cookies()
	for i, cookie := range cookies {
		if cookie.Name == "SID" {
			cookies[i] = sidCookie
			updated = true
			break
		}
	}
	if !updated {
		cookies = append(cookies, sidCookie)
	}
	c.c.Jar.SetCookies(c.baseURL, cookies)

	return nil
}

// Logout ends the current session.
func (c *Client) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.Set(logoutFunc, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Response cookies won't contain `SID`, so we can just overwrite the current
	// cookie jar.
	cookies := resp.Cookies()
	c.c.Jar.SetCookies(c.baseURL, cookies)

	return nil
}

// Downstreams fetches statistics for bonded downstream channels.
func (c *Client) Downstreams() ([]Downstream, error) {
	type dto struct {
		XMLName     xml.Name     `xml:"downstream_table"`
		Downstreams []Downstream `xml:"downstream"`
	}
	resp, err := c.Get(downstreamsFunc)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var d dto
	if err := xml.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}

	return d.Downstreams, nil
}
