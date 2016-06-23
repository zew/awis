package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kataras/iris"

	"github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/gorpx"
	"github.com/zew/awis/logx"
	"github.com/zew/awis/mdl"
	"github.com/zew/awis/util"
)

var awsSess *session.Session // not needed

func init() {
	awsSess = session.New(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
}

// Builds current ISO8601 timestamp.
// eg 2007-08-31T16:47:05.0000Z
func iso8601Timestamp() string {
	t := time.Now()
	ts := t.Format("2006-01-02T15:04:05")
	// ts += ".0000Z"
	ts += "Z"
	// gmdate("Y-m-d\TH:i:s.\\0\\0\\0\\Z", time())
	logx.Printf("ts is %v", ts)
	return ts
}

// Get a http client
func httpClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	return netClient
}

func ParseIntoSite(dat []byte) ([]mdl.Site, error) {
	type Result struct {
		// Sites []Site `xml:"TopSitesResponse>Response>TopSitesResult>Alexa>TopSites>Country>Sites>Site"`
		Sites []mdl.Site `xml:"Response>TopSitesResult>Alexa>TopSites>Country>Sites>Site"` // omit the outmost tag name TopSitesResponse
	}
	result := Result{}
	err := xml.Unmarshal(dat, &result)
	if err != nil {
		return nil, err
	}
	return result.Sites, nil
}

func awisTopSites(c *iris.Context) {

	var err error

	var ServiceHost1 = "ats.amazonaws.com"

	myUrl := url.URL{}
	myUrl.Host = ServiceHost1
	myUrl.Scheme = "http"
	logx.Printf("host is %v", myUrl.String())

	vals := map[string]string{
		"Action":           "TopSites",
		"AWSAccessKeyId":   util.EnvVar("AWS_ACCESS_KEY_ID"),
		"SignatureMethod":  "HmacSHA256",
		"SignatureVersion": "2",
		"Timestamp":        iso8601Timestamp(),
		// "Signature" : "will be added by awsauth.Sign2(req)"
		"ResponseGroup": "Country",
		"Url":           util.EffectiveParam(c, "Url", "wwww.zew.de"),
		"CountryCode":   util.EffectiveParam(c, "CountryCode", "DE"),
		"Start":         util.EffectiveParam(c, "Start", "0"),
		"Count":         util.EffectiveParam(c, "Count", "5"),
	}

	queryStr := ""
	for k, v := range vals {
		queryStr += fmt.Sprintf("%v=%v&", k, v)
	}
	logx.Printf("queryStr is %v", queryStr)

	strUrl := myUrl.String() + "/?" + queryStr
	req, err := http.NewRequest("GET", strUrl, nil)
	util.CheckErr(err)
	// logx.Printf("req is %v", req)

	// Explicit or implicit -
	// At every rate - we need to call Sign2(),
	// because awsauth does not know about awis
	if false {
		awsauth.Sign2(req, awsauth.Credentials{
			AccessKeyID:     util.EnvVar("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: util.EnvVar("AWS_SECRET_ACCESS_KEY"),
			// SecurityToken:   "Security Token", // STS (optional)
		})
	} else {
		awsauth.Sign2(req)
	}
	reqSigned := req

	resp, err := httpClient().Do(reqSigned)
	util.CheckErr(err)
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	util.CheckErr(err)
	// target := html.EscapeString(string(respBytes))

	sites, err := ParseIntoSite(respBytes)
	if err != nil {
		c.Text(200, err.Error())
		return
	}

	for _, v := range sites {
		gorpx.DBMap().Insert(&v)
	}

	// c.Text(200, "xml parsed into structs")
	display := util.IndentedDump(sites)
	// c.Text(200, display)

	s := struct {
		HTMLTitle string
		Title     string
		FlashMsg  template.HTML

		FormAction       string
		ParamUrl         string
		ParamStart       string
		ParamCount       string
		ParamCountryCode string

		URL        string
		StructDump template.HTML
	}{
		HTMLTitle:        AppName() + " result",
		Title:            AppName() + " result",
		FlashMsg:         template.HTML("Alexa Web Information Service"),
		StructDump:       template.HTML(display),
		URL:              reqSigned.URL.String(),
		FormAction:       PathTopSites,
		ParamUrl:         util.EffectiveParam(c, "Url", "www.zew.de"),
		ParamStart:       util.EffectiveParam(c, "Start", "0"),
		ParamCount:       util.EffectiveParam(c, "Count", "5"),
		ParamCountryCode: util.EffectiveParam(c, "CountryCode", "DE"),
	}

	err = c.Render("index.html", s)
	util.CheckErr(err)

}
