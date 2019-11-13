package gosportmonks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"

	"bitbucket.org/softwarehouseio/victory/victory-frontend/libs/gosportmonks/context"
	"github.com/google/go-querystring/query"
)

/*
Golang Lib to access the sportmonks.com API.

- This Lib is completely based on https://github.com/digitalocean/godo.git and changed to
  work with sportmonks.com.

A football season is a stretch of time usually in two years.

- 2017/2018

It's divided into stages, at least 1 stage but usually 2

- 1st Phase / Stage
- 2nd Phase / Stage

In each stage has rounds, commonly 2 days on which competitions are held.

- 2018-04-22/2018-04-23

The results of each round move the standings forward.
*/

const (
	libraryVersion = "0.1"
	defaultBaseURL = "https://soccer.sportmonks.com/api/"
	userAgent      = "gosportmonks/" + libraryVersion
	mediaType      = "application/json"
	tokenEnvValue  = "SPMKTOKEN"

	headerRateLimit     = "RateLimit-Limit"
	headerRateRemaining = "RateLimit-Remaining"
	headerRateReset     = "RateLimit-Reset"
)

// Client manages communication with DigitalOcean V2 API.
type Client struct {
	// HTTP client used to communicate with the DO API.
	client *http.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// User agent for client
	UserAgent string

	// Rate contains the current rate limit for the client as determined by the most recent
	// API call.
	Rate Rate

	// Services used for communicating with the API
	//Account           AccountService
	Leagues    LeaguesService
	Teams      TeamsService
	Seasons    SeasonsService
	Fixtures   FixturesService
	Topscorers TopscorersService
	Standings  StandingsService
	Players    PlayersService
	Livescores LivescoresService

	// Optional function called after every successful request made to the DO APIs
	onRequestCompleted RequestCompletionCallback
}

// RequestCompletionCallback defines the type of the request callback function
type RequestCompletionCallback func(*http.Request, *http.Response)

// TokenOption specifies the token parameter required by the Sportmonks API
type TokenOption struct {
	Token string `url:"api_token"`
}

// ListOptions specifies the optional parameters to various List methods that
// support pagination.
type ListOptions struct {
	Include string `url:"include,omitempty"`
	// For paginated result sets, page of results to retrieve.
	Page int `url:"page,omitempty"`

	// For paginated result sets, the number of results to include per page.
	PerPage int `url:"per_page,omitempty"`

	// For fixtures - leagues a comma separated list of leagues (only supported on fixtures endpoint)
	Leagues string `url:"leagues,omitempty"`
}

// Response is a DigitalOcean response. This wraps the standard http.Response returned from DigitalOcean.
type Response struct {
	*http.Response

	// Links that were returned with the response. These are parsed from
	// request body and not the header.
	//Links *Links

	// Monitoring URI
	Monitor string

	Rate
}

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string `json:"message"`

	// Code returned from the API, HTTP STATUS CODE
	Code int `json:"code"`
}

// Rate contains the rate limit for the current client.
type Rate struct {
	// The number of request per hour the client is currently limited to.
	Limit int `json:"limit"`

	// The number of remaining requests the client can make this hour.
	Remaining int `json:"remaining"`

	// The time at which the current rate limit will reset.
	Reset Timestamp `json:"reset"`
}

// Declaration

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	for k, v := range newValues {
		origValues[k] = v
	}

	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}

func AddOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	for k, v := range newValues {
		origValues[k] = v
	}

	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}

func addOptionsURL(originalURL *url.URL, opt interface{}) (*url.URL, error) {
	origURL := originalURL

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return originalURL, err
	}

	for k, v := range newValues {
		origValues[k] = v
	}

	origURL.RawQuery = origValues.Encode()
	return origURL, nil
}

// NewClient returns a new DigitalOcean API client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	token := os.Getenv(tokenEnvValue)

	baseURLWithApiToken, err := addOptions(defaultBaseURL, TokenOption{
		Token: token,
	})
	if err != nil {
		log.Printf("added default api_token failed: %v", err)
		return nil
	}

	baseURL, _ := url.Parse(baseURLWithApiToken)
	log.Printf("GSM.NewClient: baseURL: %v", baseURL)

	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}
	//c.Account = &AccountServiceOp{client: c}
	c.Leagues = &LeaguesServiceOp{client: c}
	c.Teams = &TeamsServiceOp{client: c}
	c.Seasons = &SeasonsServiceOp{client: c}
	c.Fixtures = &FixturesServiceOp{client: c}
	c.Topscorers = &TopscorersServiceOp{client: c}
	c.Standings = &StandingsServiceOp{client: c}
	c.Players = &PlayersServiceOp{client: c}
	c.Livescores = &LivescoresServiceOp{client: c}

	return c
}

// ClientOpt are options for New.
type ClientOpt func(*Client) error

// New returns a new DigitalOcean API client instance.
func New(httpClient *http.Client, opts ...ClientOpt) (*Client, error) {
	c := NewClient(httpClient)
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// SetTokenParam is a client option for setting the Sportmonks token
// as query param
func SetTokenParam(token string) ClientOpt {
	return func(c *Client) error {
		newBaseURL, err := addOptionsURL(c.BaseURL, &TokenOption{
			Token: token,
		})
		log.Printf("GoSportMonks.SetTokenParam %v", newBaseURL.String())
		if err != nil {
			return err
		}
		c.BaseURL = newBaseURL
		return nil
	}
}

// SetBaseURL is a client option for setting the base URL.
func SetBaseURL(bu string) ClientOpt {
	return func(c *Client) error {
		u, err := url.Parse(bu)
		if err != nil {
			return err
		}

		c.BaseURL = u
		return nil
	}
}

// SetUserAgent is a client option for setting the user agent.
func SetUserAgent(ua string) ClientOpt {
	return func(c *Client) error {
		c.UserAgent = fmt.Sprintf("%s %s", ua, c.UserAgent)
		return nil
	}
}

// NewRequest creates an API request. A relative URL can be provided in urlStr, which will be resolved to the
// BaseURL of the Client. Relative URLS should always be specified without a preceding slash. If specified, the
// value pointed to by body is JSON encoded and included in as the request body.
func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	log.Printf("GSM.Client.NewRequest - urlStr %v rel %v", urlStr, rel)

	u := c.BaseURL.ResolveReference(rel)

	token := os.Getenv(tokenEnvValue)

	u, err = addOptionsURL(u, TokenOption{
		Token: token,
	})
	if err != nil {
		log.Printf("GSM.Client.NewRequest: added default api_token failed: %v", err)
		return nil, err
	}

	log.Printf("GSM.Client.NewRequest: Resolved ref: %v", u.String())

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// OnRequestCompleted sets the DO API request completion callback
func (c *Client) OnRequestCompleted(rc RequestCompletionCallback) {
	c.onRequestCompleted = rc
}

// newResponse creates a new Response for the provided http.Response
func newResponse(r *http.Response) *Response {
	response := Response{Response: r}
	response.populateRate()

	return &response
}

// populateRate parses the rate related headers and populates the response Rate.
func (r *Response) populateRate() {
	if limit := r.Header.Get(headerRateLimit); limit != "" {
		r.Rate.Limit, _ = strconv.Atoi(limit)
	}
	if remaining := r.Header.Get(headerRateRemaining); remaining != "" {
		r.Rate.Remaining, _ = strconv.Atoi(remaining)
	}
	if reset := r.Header.Get(headerRateReset); reset != "" {
		if v, _ := strconv.ParseInt(reset, 10, 64); v != 0 {
			r.Rate.Reset = Timestamp{time.Unix(v, 0)}
		}
	}
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	log.Printf("GSM.Client.Do URL: %v", req.URL.String())
	resp, err := context.DoRequestWithClient(ctx, c.client, req)
	if err != nil {
		return nil, err
	}
	if c.onRequestCompleted != nil {
		c.onRequestCompleted(req, resp)
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	response := newResponse(resp)
	c.Rate = response.Rate

	err = CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {

		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, err
}
func (r *ErrorResponse) Error() string {
	if r.Code != 0 {
		return fmt.Sprintf("%v %v: %d (code %d) %v",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Code, r.Message)
	}
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

// CheckResponse checks the API response for errors, and returns them if present. A response is considered an
// error if it has a status code outside the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = string(data)
		}
	}

	return errorResponse
}

func (r Rate) String() string {
	return Stringify(r)
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

// Int is a helper routine that allocates a new int32 value
// to store v and returns a pointer to it, but unlike Int32
// its argument value is an int.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// StreamToString converts a reader to a string
func StreamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(stream)
	return buf.String()
}
