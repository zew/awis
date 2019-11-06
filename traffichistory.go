package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kataras/iris/v12"
	awsauth "github.com/smartystreets/go-aws-auth"

	"github.com/zew/awis/mdl"
	"github.com/zew/gorpx"
	"github.com/zew/logx"
	"github.com/zew/util"
)

func trafficHistory(c iris.Context) {

	var err error
	reqSigned, _ := http.NewRequest("GET", Pref(), nil)
	display := ""
	respBytes := []byte{}

	start := EffectiveParamInt(c, "Start", 1)
	count := EffectiveParamInt(c, "Count", 5)

	sites := []mdl.Domain{}

	if EffectiveParam(c, "submit", "none") != "none" {
		sql := `SELECT  
			site_id,
			domain_name,
			global_rank,
			country_rank,
			country_reach_permillion,
			country_pageviews_permillion,
			country_pageviews_peruser

		FROM 			` + gorpx.DbTableName(mdl.Domain{}) + ` t1
		WHERE 			1=1
				AND		site_id >= :site_id_start
				AND		site_id <= :site_id_end
			`
		args := map[string]interface{}{
			"site_id_start": start,
			"site_id_end":   start + count,
		}
		_, err = gorpx.DbMap().Select(&sites, sql, args)
		util.CheckErr(err)

	}

	logx.Printf("sites are %v", sites)
	awsAccessKeyID, _ := util.EnvVar("AWS_ACCESS_KEY_ID")
	for _, site := range sites {

		myUrl := url.URL{}
		var ServiceHost2 = "awis.amazonaws.com"
		myUrl.Host = ServiceHost2
		myUrl.Scheme = "http"
		logx.Printf("host is %v", myUrl.String())

		vals := map[string]string{
			"Action":           "TrafficHistory",
			"AWSAccessKeyId":   awsAccessKeyID,
			"SignatureMethod":  "HmacSHA256",
			"SignatureVersion": "2",
			"Timestamp":        iso8601Timestamp(),
			// "Signature" : "will be added by awsauth.Sign2(req)"
			"ResponseGroup": "History",

			"Url":         site.Name,
			"CountryCode": EffectiveParam(c, "CountryCode", "DE"), // has no effect :(
			"Start":       EffectiveParam(c, "DateBegin", "20150101"),
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

		trafHists := mdl.Histories{}
		err = xml.Unmarshal(respBytes, &trafHists)
		if err != nil {
			str := fmt.Sprintf("Error unmarschalling bytes for %v - size -%v-   - error %v\n\n", site.Name, len(respBytes), err)
			logx.Print(str)
			display = str + display

			err = ioutil.WriteFile("./traffic-data-"+site.Name+".xml", respBytes, 0644)
			util.CheckErr(err)

			continue
		}

		for _, oneHist := range trafHists.Histories {
			oneHist.Name = site.Name
			err = gorpx.DbMap().Insert(&oneHist)
			util.CheckErr(err, "duplicate entry")
		}

		display += util.IndentedDump(trafHists) + "\n"
		// c.WriteString(display)
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

		ParamStart:     EffectiveParam(c, "Start", "1"),
		ParamCount:     EffectiveParam(c, "Count", "5"),
		ParamDateBegin: EffectiveParam(c, "DateBegin", "20150101"),

		StructDump: template.HTML(display),
	}

	err = c.View("traffic-history.html", s)
	util.CheckErr(err)

}
