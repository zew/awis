package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/kataras/iris"

	"github.com/zew/gorpx"
	"github.com/zew/irisx"

	"github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/mdl"
	"github.com/zew/logx"
	"github.com/zew/util"
)

func ParseIntoStructs(dat []byte) (mdl.Meta, []mdl.Rank, []mdl.Category, error) {
	type Result struct {
		Meta mdl.Meta `xml:"Response>UrlInfoResult>Alexa"` // omit the outmost tag name TopSitesResponse
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
		return res1.Meta, nil, nil, err
	}
	res2 := ResRanks{}
	err = xml.Unmarshal(dat, &res2)
	if err != nil {
		return res1.Meta, nil, nil, err
	}
	res3 := ResCats{}
	err = xml.Unmarshal(dat, &res3)
	if err != nil {
		return res1.Meta, nil, nil, err
	}
	return res1.Meta, res2.Ranks, res3.Categories, nil
}

func ParseDeltas(dat []byte) ([]mdl.Delta, error) {
	type ResDeltas struct {
		Deltas []mdl.Delta `xml:"Response>UrlInfoResult>Alexa>TrafficData>UsageStatistics>UsageStatistic"`
	}
	res4 := ResDeltas{}
	err := xml.Unmarshal(dat, &res4)
	if err != nil {
		return nil, err
	}
	return res4.Deltas, nil
}

func awisDomainInfo(c *iris.Context) {

	var err error
	reqSigned, _ := http.NewRequest("GET", Pref(), nil)
	display := ""
	errors := ""
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

		site := mdl.Domain{}
		sql := `SELECT 
		      domain_id
			, domain_name
			, global_rank
			, country_rank
		FROM 			` + gorpx.TableName(mdl.Domain{}) + ` t1
		WHERE 			1=1
				AND		domain_id = :domain_id
			`
		args := map[string]interface{}{
			"domain_id": i,
		}
		err = gorpx.DBMap().SelectOne(&site, sql, args)
		util.CheckErr(err)
		// c.Text(200, fmt.Sprintf("%v - %+v\n\n", i, site))
		sites = append(sites, site.Name)

	}

	logx.Printf("sites are %v", sites)

	// return

	ts := unixDayStamp()

	for _, site := range sites {

		myUrl := url.URL{}
		var ServiceHost2 = "awis.amazonaws.com"
		myUrl.Host = ServiceHost2
		myUrl.Scheme = "http"
		// logx.Printf("host is %v", myUrl.String())

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

		meta, ranks, categories, err := ParseIntoStructs(respBytes)
		if err != nil {
			c.Text(200, err.Error())
			return
		}

		if meta.Name == "" {
			log.Fatalf("meta name is empty; should contain domain name. %+v", meta)
		}

		if false {
			// <aws:Value>+-3,244.2</aws:Value>
			r1, err := regexp.Compile(`<aws:Value>[0-9\.\-\+]+([,])[0-9\.\-\+]+</aws:Value>`)
			util.CheckErr(err)
			matches := r1.FindAllSubmatchIndex(respBytes, -1)
			for idx, match := range matches {
				fmt.Printf("%2v result is %v\n", idx, match)
				if len(match) > 3 {
					p1 := match[2]
					p2 := match[3]
					subMatch := respBytes[p1:p2]
					_ = subMatch
					fmt.Printf("    -%s-\n", subMatch)
				}
			}
			if err != nil {
				c.Text(200, err.Error())
				return
			}

			flushOutCommata := func(r rune) rune {
				switch {
				case r == ',':
					return rune(-1) // drop
				}
				return r
			}
			respBytes = bytes.Map(flushOutCommata, respBytes)
		}

		respBytes = bytes.Replace(respBytes, []byte(","), []byte(""), -1)
		deltas, err := ParseDeltas(respBytes)

		meta.LastUpdated = ts
		err = gorpx.DBMap().Insert(&meta)
		if err != nil {
			errors += fmt.Sprintf("meta: %v\n", err)
		}

		for _, rank := range ranks {
			rank.Name = meta.Name
			rank.LastUpdated = ts
			err = gorpx.DBMap().Insert(&rank)
			if err != nil {
				errors += fmt.Sprintf("rank: %v\n", err)
				break
			}
		}
		for catRecordIdx, cat := range categories {
			cat.Name = meta.Name
			cat.LastUpdated = ts

			err = gorpx.DBMap().Insert(&cat)
			if err != nil {
				errors += fmt.Sprintf("cat: %v\n", err)
				break
			}
			if catRecordIdx > 4 {
				break // max five cats
			}
		}
		for _, delta := range deltas {
			delta.Name = meta.Name
			delta.LastUpdated = ts
			// logx.Printf("--------------delta is %+v", delta)

			err = gorpx.DBMap().Insert(&delta)
			if err != nil {
				errors += fmt.Sprintf("delta: %v\n", err)
			}
		}

		display += util.IndentedDump(meta) + "\n"
		display += util.IndentedDump(ranks) + "\n"
		display += util.IndentedDump(categories) + "\n"
		display += util.IndentedDump(deltas)
		// c.Text(200, display)
	}

	display = errors + "\n\n" + display

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
		HTMLTitle:  AppName() + " url infos",
		Title:      AppName() + " url infos",
		FlashMsg:   template.HTML("Alexa Web Information Service"),
		URL:        reqSigned.URL.String(),
		FormAction: PathDomainInfo,
		ParamUrl:   irisx.EffectiveParam(c, "Url", "www.zew.de"),
		ParamStart: irisx.EffectiveParam(c, "Start", "1"),
		ParamCount: irisx.EffectiveParam(c, "Count", "5"),

		StructDump1: template.HTML(string(respBytes)),
		StructDump2: template.HTML(display),
	}

	err = c.Render("form.html", s)
	util.CheckErr(err)

}
