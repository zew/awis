package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kataras/iris"

	"github.com/zew/awis/gorpx"

	"github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/logx"
	"github.com/zew/awis/mdl"
	"github.com/zew/awis/util"
)

func ParseIntoContact(dat []byte) (mdl.Meta, []mdl.Rank, error) {
	type Result struct {
		Contact mdl.Meta `xml:"Response>UrlInfoResult>Alexa"` // omit the outmost tag name TopSitesResponse
	}
	type ResRanks struct {
		Ranks []mdl.Rank `xml:"Response>UrlInfoResult>Alexa>TrafficData>RankByCountry>Country"`
	}
	res1 := Result{}
	err := xml.Unmarshal(dat, &res1)
	if err != nil {
		return res1.Contact, nil, err
	}
	res2 := ResRanks{}
	err = xml.Unmarshal(dat, &res2)
	if err != nil {
		return res1.Contact, nil, err
	}
	return res1.Contact, res2.Ranks, nil
}

func awisDomainInfo(c *iris.Context) {

	var err error

	myUrl := url.URL{}
	var ServiceHost2 = "awis.amazonaws.com"
	myUrl.Host = ServiceHost2
	myUrl.Scheme = "http"
	logx.Printf("host is %v", myUrl.String())

	vals := map[string]string{
		"Action":           "UrlInfo",
		"AWSAccessKeyId":   util.EnvVar("AWS_ACCESS_KEY_ID"),
		"SignatureMethod":  "HmacSHA256",
		"SignatureVersion": "2",
		"Timestamp":        iso8601Timestamp(),
		// "Signature" : "will be added by awsauth.Sign2(req)"
		"ResponseGroup": "RelatedLinks,Categories,RankByCountry,UsageStats,AdultContent,Speed,Language,OwnedDomains,LinksInCount,SiteData,ContactInfo",
		"Url":           util.EffectiveParam(c, "Url", "wwww.zew.de"),
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

	awsauth.Sign2(req)
	reqSigned := req

	resp, err := httpClient().Do(reqSigned)
	util.CheckErr(err)
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	util.CheckErr(err)
	// target := html.EscapeString(string(respBytes))

	contact, ranks, err := ParseIntoContact(respBytes)
	if err != nil {
		c.Text(200, err.Error())
		return
	}

	err = gorpx.DBMap().Insert(&contact)
	if err != nil {
		c.Text(200, err.Error())
	} else {
		for _, rank := range ranks {
			rank.Name = contact.Name
			err = gorpx.DBMap().Insert(&rank)
			if err != nil {
				c.Text(200, err.Error())
				break
			}
		}
	}

	display := util.IndentedDump(contact)
	display = display + "\n" + util.IndentedDump(ranks)
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

		URL         string
		StructDump  template.HTML
		StructDump2 template.HTML
	}{
		HTMLTitle:  AppName() + " result",
		Title:      AppName() + " result",
		FlashMsg:   template.HTML("Alexa Web Information Service"),
		URL:        reqSigned.URL.String(),
		FormAction: PathDomainInfo,
		ParamUrl:   util.EffectiveParam(c, "Url", "www.zew.de"),

		StructDump:  template.HTML(string(respBytes)),
		StructDump2: template.HTML(display),
	}

	err = c.Render("index.html", s)
	util.CheckErr(err)

}
