package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kataras/iris"

	"github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/mdl"
	"github.com/zew/gorpx"
	"github.com/zew/irisx"
	"github.com/zew/logx"
	"github.com/zew/util"
)

func trafficHistory(c *iris.Context) {

	var err error
	reqSigned, _ := http.NewRequest("GET", Pref(), nil)
	display := ""
	respBytes := []byte{}

	start, _, _ := irisx.EffectiveParamInt(c, "Start", 1)
	count, _, _ := irisx.EffectiveParamInt(c, "Count", 5)

	sites := []mdl.Site{}

	if irisx.EffectiveParam(c, "submit", "none") != "none" {
		sql := `SELECT  
			site_id,
			domain_name,
			global_rank,
			country_rank,
			country_reach_permillion,
			country_pageviews_permillion,
			country_pageviews_peruser

		FROM 			` + gorpx.TableName(mdl.Site{}) + ` t1
		WHERE 			1=1
				AND		site_id >= :site_id_start
				AND		site_id <= :site_id_end
			`
		args := map[string]interface{}{
			"site_id_start": start,
			"site_id_end":   start + count,
		}
		_, err = gorpx.DBMap().Select(&sites, sql, args)
		util.CheckErr(err)

	}

	logx.Printf("sites are %v", sites)

	for _, site := range sites {

		myUrl := url.URL{}
		var ServiceHost2 = "awis.amazonaws.com"
		myUrl.Host = ServiceHost2
		myUrl.Scheme = "http"
		logx.Printf("host is %v", myUrl.String())

		vals := map[string]string{
			"Action":           "TrafficHistory",
			"AWSAccessKeyId":   util.EnvVar("AWS_ACCESS_KEY_ID"),
			"SignatureMethod":  "HmacSHA256",
			"SignatureVersion": "2",
			"Timestamp":        iso8601Timestamp(),
			// "Signature" : "will be added by awsauth.Sign2(req)"
			"ResponseGroup": "History",

			"Url":         site.Name,
			"CountryCode": irisx.EffectiveParam(c, "CountryCode", "DE"), // has no effect :(
			"Start":       irisx.EffectiveParam(c, "DateBegin", "20150101"),
			"Range":       "31",
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
		if err != nil {
			str := fmt.Sprintf("%v: Error reading body: %v\n\n", site.Name, err)
			logx.Print(str)
			display = str + display
			continue
		}

		trafHists := mdl.TrafHistories{}
		err = xml.Unmarshal(respBytes, &trafHists)
		if err != nil {
			str := fmt.Sprintf("Error unmarschalling bytes for %v - size -%v-   - error %v\n\n", site.Name, len(respBytes), err)
			logx.Print(str)
			display = str + display

			err = ioutil.WriteFile("./traffic-data-"+site.Name+".xml", respBytes, 0644)
			util.CheckErr(err)

			continue
		}

		for _, oneHist := range trafHists.TrafficHistories {
			oneHist.Site = site.Name
			err = gorpx.DBMap().Insert(&oneHist)
			util.CheckErr(err, "duplicate entry")
		}

		display += util.IndentedDump(trafHists) + "\n"
		// c.Text(200, display)
	}

	s := struct {
		HTMLTitle string
		Title     string
		FlashMsg  template.HTML

		URL string

		FormAction string

		ParamStart     string
		ParamCount     string
		ParamDateBegin string

		StructDump template.HTML
	}{
		HTMLTitle: AppName() + " - Traffic History",
		Title:     AppName() + " - Traffic History",
		FlashMsg:  template.HTML("Alexa Web Information Service"),

		URL: reqSigned.URL.String(),

		FormAction: TrafficHistory,

		ParamStart:     irisx.EffectiveParam(c, "Start", "1"),
		ParamCount:     irisx.EffectiveParam(c, "Count", "5"),
		ParamDateBegin: irisx.EffectiveParam(c, "DateBegin", "20150101"),

		StructDump: template.HTML(display),
	}

	err = c.Render("traffic-history.html", s)
	util.CheckErr(err)

}
