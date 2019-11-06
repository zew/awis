package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kataras/iris/v12"

	awsauth "github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/mdl"
	"github.com/zew/gorpx"
	"github.com/zew/logx"
	"github.com/zew/util"
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
	// logx.Printf("ts is %v", ts)
	return ts
}

func unixDayStamp() int {
	ts := int(int32(time.Now().Unix()))
	ts = int(ts/(24*3600))*24*3600 + 9*3600 // norm it towards a single day; 8 in the morning
	return ts
}

func ParseIntoDomains(dat []byte) ([]mdl.Domain, error) {
	type Result struct {
		// Sites []Site `xml:"TopSitesResponse>Response>TopSitesResult>Alexa>TopSites>Country>Sites>Site"`
		Sites []mdl.Domain `xml:"Response>TopSitesResult>Alexa>TopSites>Country>Sites>Site"` // omit the outmost tag name TopSitesResponse
	}
	result := Result{}
	err := xml.Unmarshal(dat, &result)
	if err != nil {
		return nil, err
	}
	return result.Sites, nil
}

func topSites(c iris.Context) {

	var err error
	reqSigned, _ := http.NewRequest("GET", Pref(), nil)
	display := ""
	errors := ""
	respBytes := []byte{}

	ts := unixDayStamp()

	if EffectiveParam(c, "submit", "none") != "none" {

		var ServiceHost1 = "ats.amazonaws.com"

		myURL := url.URL{}
		myURL.Host = ServiceHost1
		myURL.Scheme = "http"
		// logx.Printf("host is %v", myURL.String())

		awsAccessKeyID, _ := util.EnvVar("AWS_ACCESS_KEY_ID")

		vals := map[string]string{
			"Action":           "TopSites",
			"AWSAccessKeyId":   awsAccessKeyID,
			"SignatureMethod":  "HmacSHA256",
			"SignatureVersion": "2",
			"Timestamp":        iso8601Timestamp(),
			// "Signature" : "will be added by awsauth.Sign2(req)"
			"ResponseGroup": "Country",
			"Url":           EffectiveParam(c, "Url", "wwww.zew.de"),
			"CountryCode":   EffectiveParam(c, "CountryCode", "DE"),
			"Start":         EffectiveParam(c, "Start", "0"),
			"Count":         EffectiveParam(c, "Count", "5"),
		}

		queryStr := ""
		for k, v := range vals {
			queryStr += fmt.Sprintf("%v=%v&", k, v)
		}
		logx.Printf("queryStr is %v", queryStr)

		strURL := myURL.String() + "/?" + queryStr
		req, err := http.NewRequest("GET", strURL, nil)
		util.CheckErr(err)
		// logx.Printf("req is %v", req)

		// Explicit or implicit -
		// At every rate - we need to call Sign2(),
		// because awsauth does not know about awis
		if false {
			awsSecretAccessKey, _ := util.EnvVar("AWS_SECRET_ACCESS_KEY")

			awsauth.Sign2(req, awsauth.Credentials{
				AccessKeyID:     awsAccessKeyID,
				SecretAccessKey: awsSecretAccessKey,
				// SecurityToken:   "Security Token", // STS (optional)
			})
		} else {
			awsauth.Sign2(req)
		}
		reqSigned = req

		resp, err := util.HttpClient().Do(reqSigned)
		util.CheckErr(err)
		defer resp.Body.Close()

		respBytes, err = ioutil.ReadAll(resp.Body)
		util.CheckErr(err)
		// target := html.EscapeString(string(respBytes))

		domains, err := ParseIntoDomains(respBytes)
		if err != nil {
			errors += fmt.Sprintf("xml parsing failded: %v\n", err)
		}

		for _, domain := range domains {
			domain.LastUpdated = ts
			err := gorpx.DbMap().Insert(&domain)
			if err != nil {
				errors += fmt.Sprintf("domain: %v\n", err)
			}
		}

		display += util.IndentedDump(domains)

		display = errors + "\n\n" + display

	}

	s := struct {
		HTMLTitle string
		Title     string
		FlashMsg  template.HTML

		FormAction       string
		ParamUrl         string
		ParamStart       string
		ParamCount       string
		ParamCountryCode string

		URL         string
		StructDump1 template.HTML
		StructDump2 template.HTML
	}{
		HTMLTitle:        AppName() + " top sites",
		Title:            AppName() + " top sites",
		FlashMsg:         template.HTML("Alexa Web Information Service"),
		StructDump2:      template.HTML(display),
		URL:              reqSigned.URL.String(),
		FormAction:       PathTopSites,
		ParamUrl:         EffectiveParam(c, "Url", "www.zew.de"),
		ParamStart:       EffectiveParam(c, "Start", "0"),
		ParamCount:       EffectiveParam(c, "Count", "5"),
		ParamCountryCode: EffectiveParam(c, "CountryCode", "DE"),
	}

	err = c.View("form.html", s)
	util.CheckErr(err)
}
