package main

import (
	"html/template"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kataras/iris"
	"github.com/zew/awis/logx"
	"github.com/zew/awis/util"
)

var ServiceHost1 = "ats.amazonaws.com"
var ServiceHost2 = "awis.amazonaws.com"

var awsSess *session.Session

func init() {
	awsSess = session.New(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
}

// get a http client
func client() *http.Client {
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

// Builds current ISO8601 timestamp.
// eg 2007-08-31T16:47:05.0000Z
func getTimestamp() string {
	t := time.Now()
	ts := t.Format("2006-01-02T15:04:05")
	// ts += ".0000Z"
	ts += "Z"
	// gmdate("Y-m-d\TH:i:s.\\0\\0\\0\\Z", time())
	logx.Printf("ts is %v", ts)
	return ts
}

func index(c *iris.Context) {

	var err error
	s := struct {
		HTMLTitle string
		Title     string
		FlashMsg  template.HTML

		ParamUrl         string
		ParamStart       string
		ParamCount       string
		ParamCountryCode string

		URL  string
		JSON template.HTML
	}{
		HTMLTitle: AppName() + " main",
		Title:     AppName() + " main",
		FlashMsg:  template.HTML("Alexa Web Information Service"),
		// JSON:      template.HTML(target),
		// URL:       reqSigned.URL.String(),
		ParamUrl:         util.EffectiveParam(c, "Url", "www.zew.de"),
		ParamStart:       util.EffectiveParam(c, "Start", "0"),
		ParamCount:       util.EffectiveParam(c, "Count", "5"),
		ParamCountryCode: util.EffectiveParam(c, "CountryCode", "DE"),
	}

	err = c.Render("index.html", s)
	util.CheckErr(err)

}
