package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kataras/iris"

	"github.com/zew/gorpx"
	"github.com/zew/irisx"

	"github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/mdl"
	"github.com/zew/logx"
	"github.com/zew/util"
)

func ParseIntoContact(dat []byte) (mdl.Meta, []mdl.Rank, []mdl.Category, error) {
	type Result struct {
		Contact mdl.Meta `xml:"Response>UrlInfoResult>Alexa"` // omit the outmost tag name TopSitesResponse
	}
	type ResRanks struct {
		Ranks []mdl.Rank `xml:"Response>UrlInfoResult>Alexa>TrafficData>RankByCountry>Country"`
	}
	type ResCats struct {
		Categories []mdl.Category `xml:"Response>UrlInfoResult>Alexa>Related>Categories>CategoryData"`
	}
	res1 := Result{}
	err := xml.Unmarshal(dat, &res1)
	if err != nil {
		return res1.Contact, nil, nil, err
	}
	res2 := ResRanks{}
	err = xml.Unmarshal(dat, &res2)
	if err != nil {
		return res1.Contact, nil, nil, err
	}
	res3 := ResCats{}
	err = xml.Unmarshal(dat, &res3)
	if err != nil {
		return res1.Contact, nil, nil, err
	}
	return res1.Contact, res2.Ranks, res3.Categories, nil
}

func awisDomainInfo(c *iris.Context) {

	var err error
	reqSigned, _ := http.NewRequest("GET", Pref(), nil)
	display := ""
	respBytes := []byte{}

	startFl, _, _ := irisx.EffectiveParamFloat(c, "Start", 1.0)
	startCn, _, _ := irisx.EffectiveParamFloat(c, "Count", 5.0)
	start := int(startFl)
	count := int(startCn)
	sites := []string{}
	for i := start; i < start+count; i++ {

		if irisx.EffectiveParam(c, "submit", "none") == "none" {
			continue
		}

		site := mdl.Site{}
		sql := `SELECT 
		      site_id
			, domain_name
			, global_rank
			, country_rank
		FROM 			` + gorpx.TableName(mdl.Site{}) + ` t1
		WHERE 			1=1
				AND		site_id = :site_id
			`
		args := map[string]interface{}{
			"site_id": i,
		}
		err = gorpx.DBMap().SelectOne(&site, sql, args)
		util.CheckErr(err)
		// c.Text(200, fmt.Sprintf("%v - %+v\n\n", i, site))
		sites = append(sites, site.Name)

	}

	logx.Printf("sites are %v", sites)

	// return

	for _, site := range sites {

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

			"Url": site,
			// "Url":           irisx.EffectiveParamIsSet(c, "Url", "wwww.zew.de"),
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
		reqSigned = req

		resp, err := util.HttpClient().Do(reqSigned)
		util.CheckErr(err)
		defer resp.Body.Close()

		respBytes, err = ioutil.ReadAll(resp.Body)
		util.CheckErr(err)
		// target := html.EscapeString(string(respBytes))

		contact, ranks, categories, err := ParseIntoContact(respBytes)
		if err != nil {
			c.Text(200, err.Error())
			return
		}

		err = gorpx.DBMap().Insert(&contact)
		if err != nil {
			c.Text(200, err.Error())
		} else {
			for rankRecordIdx, rank := range ranks {
				rank.Name = contact.Name
				err = gorpx.DBMap().Insert(&rank)
				if err != nil {
					c.Text(200, err.Error())
					break
				}
				if rankRecordIdx > 20 {
					break // max five ranks from top to bottom
				}
			}
			for catRecordIdx, cat := range categories {
				cat.Name = contact.Name
				err = gorpx.DBMap().Insert(&cat)
				if err != nil {
					c.Text(200, err.Error())
					break
				}
				if catRecordIdx > 4 {
					break // max five cats
				}
			}
		}

		display = util.IndentedDump(contact) + "\n" + util.IndentedDump(ranks) + "\n" + util.IndentedDump(categories)
		// c.Text(200, display)
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
		StructDump  template.HTML
		StructDump2 template.HTML
	}{
		HTMLTitle:  AppName() + " url infos",
		Title:      AppName() + " url infos",
		FlashMsg:   template.HTML("Alexa Web Information Service"),
		URL:        reqSigned.URL.String(),
		FormAction: PathDomainInfo,
		ParamUrl:   irisx.EffectiveParam(c, "Url", "www.zew.de"),
		ParamStart: irisx.EffectiveParam(c, "Start", "1"),
		ParamCount: irisx.EffectiveParam(c, "Count", "5"),

		StructDump:  template.HTML(string(respBytes)),
		StructDump2: template.HTML(display),
	}

	err = c.Render("form.html", s)
	util.CheckErr(err)

}
