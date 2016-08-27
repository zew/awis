package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kataras/iris"

	"github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/mdl"
	"github.com/zew/gorpx"
	"github.com/zew/irisx"
	"github.com/zew/logx"
	"github.com/zew/util"
)

func trafficHistoryFillMissingHoles(c *iris.Context) {

	var err error
	reqSigned, _ := http.NewRequest("GET", Pref(), nil)
	display := ""
	respBytes := []byte{}

	start, _, _ := irisx.EffectiveParamInt(c, "Start", 0)
	count, _, _ := irisx.EffectiveParamInt(c, "Count", 5)
	granularity, _, _ := irisx.EffectiveParamInt(c, "Granularity", 1)
	dateBegin := irisx.EffectiveParam(c, "DateBegin", "20150101")

	sites := []mdl.Domain{
		{Name: "7tv.de"},
		{Name: "advertiserurl.com"},
		{Name: "adexc.net"},
		{Name: "alexa.com"},
		{Name: "bidverdrd.com"},
		{Name: "cine.to"},
		{Name: "cakepornvids.com"},
		{Name: "cloudzonetrk.com"},
		{Name: "csgohouse.com"},
		{Name: "dirty-time.net"},
		{Name: "dw.com"},
		{Name: "ebay-kleinanzeigen.de"},
		{Name: "eurowings.com"},
		{Name: "freepornx.org"},
		{Name: "futwatch.com"},
		{Name: "fussball-em-2016.com"},
		{Name: "gogoanime.io"},
		{Name: "heftig.de"},
		{Name: "hespress.com"},
		{Name: "henkel-lifetimes.de"},
		{Name: "hotmovs.com"},
		{Name: "just4single.com"},
		{Name: "labymod.net"},
		{Name: "lernhelfer.de"},
		{Name: "magentacloud.de"},
		{Name: "moneyhouse.de"},
		{Name: "netbet.de"},
		{Name: "nurxxx.mobi"},
		{Name: "onedio.com"},
		{Name: "ontests.me"},
		{Name: "pckeeper.software"},
		{Name: "playoverwatch.com"},
		{Name: "pussyspace.com"},
		{Name: "rock-am-ring.com"},
		{Name: "relink.to"},
		{Name: "spotscenered.info"},
		{Name: "shadbase.com"},
		{Name: "tvnow.de"},
		{Name: "vicomi.com"},
		{Name: "vidaxl.de"},
		{Name: "wahnsinn.tv"},
		{Name: "wiocha.pl"},
		{Name: "xxxstreams.org"},
	}

	// sites = sites[0:3]

	logx.Printf("sites are %v", sites)

	for idxSite, site := range sites {

		if irisx.EffectiveParam(c, "submit", "none") == "none" {
			break
		}
		display += site.Name + "\n"

		allExistingRecords, err := gorpx.DbMap1().SelectInt(
			"SELECT count(*) FROM "+gorpx.Db1TableName(mdl.History{})+" WHERE domain_name = :site ",
			map[string]interface{}{
				"site": site.Name,
			},
		)
		util.CheckErr(err)
		sites[idxSite].GlobalRank = int(allExistingRecords)
		if allExistingRecords < 20 {
			display += fmt.Sprintf("    only %v records - we skip\n", allExistingRecords)
			continue
		}

		for i := start; i < count; i += granularity {

			allDatesAws, allDatesSql := dayStepsFromString(dateBegin, i, granularity)

			datesSql := "'" + strings.Join(allDatesSql, "', '") + "'"

			logx.Printf("datesSql are %v", datesSql)

			sql := "SELECT count(*) FROM " + gorpx.Db1TableName(mdl.History{}) +
				" WHERE domain_name = :site AND date IN (" + datesSql + ")				"

			existingRecords, err := gorpx.DbMap1().SelectInt(
				sql,
				map[string]interface{}{
					"site": site.Name,
				},
			)
			util.CheckErr(err)

			if existingRecords >= int64(granularity) {
				continue
			}
			continue

			display += fmt.Sprintf("found only %v sql records for %v (%v)\n", existingRecords, site.Name, datesSql)

			// lastStep := allDatesAWS[len(allDatesAWS)-1]
			firstStep := allDatesAws[0]

			myUrl := url.URL{}
			var ServiceHost2 = "awis.amazonaws.com"
			myUrl.Host = ServiceHost2
			myUrl.Scheme = "http"
			// logx.Printf("host is %v", myUrl.String())

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
				"Start":       firstStep,
				"Range":       fmt.Sprintf("%v", granularity),
			}

			queryStr := ""
			for k, v := range vals {
				queryStr += fmt.Sprintf("%v=%v&", k, v)
			}
			// logx.Printf("queryStr is %v", queryStr)

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
				// err = ioutil.WriteFile("./traffic-data-"+site.Name+".xml", respBytes, 0644)
				// util.CheckErr(err)
				continue
			}

			display += fmt.Sprintf("found %v xml traffic history nodes for %v\n", len(trafHists.Histories), firstStep)
			for idx, oneHist := range trafHists.Histories {
				oneHist.Name = site.Name
				trafHists.Histories[idx].Name = site.Name
				err = gorpx.DbMap1().Insert(&oneHist)
				util.CheckErr(err, "duplicate entry")
			}

			display += util.IndentedDump(trafHists) + "\n"
			// c.Text(200, display)
		}
	}

	for idx, site := range sites {
		display += fmt.Sprintf("%2v - %3v  %-44v\n", idx, site.GlobalRank, site.Name)
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
		HTMLTitle: AppName() + " - Fill Wholes of Traffic History",
		Title:     AppName() + " - Fill Wholes of Traffic History",
		FlashMsg:  template.HTML("Alexa Web Information Service"),

		URL: reqSigned.URL.String(),

		FormAction: TrafficHistoryFillMissingHoles,

		ParamStart:     irisx.EffectiveParam(c, "Start", "0"),
		ParamCount:     irisx.EffectiveParam(c, "Count", "10"),
		ParamDateBegin: irisx.EffectiveParam(c, "DateBegin", "20150101"),

		StructDump: template.HTML(display),
	}

	err = c.Render("traffic-history.html", s)
	util.CheckErr(err)

}

// usage:
// fmt.Printf("%+v", dayStepsFromString("20150125",10) )
func dayStepsFromString(strDate string, start, numberOfDays int) ([]string, []string) {

	date, err := time.Parse("20060102", strDate)
	if err != nil {
		logx.Fatal("we want format '20060102' ")
	}

	ret1 := make([]string, 0, numberOfDays)
	ret2 := make([]string, 0, numberOfDays)

	for i := start; i < start+numberOfDays; i++ {

		d1 := time.Date(
			date.Year(), date.Month(), date.Day()+i,
			0, 0, 0, 0, time.UTC,
		)
		str1 := d1.Format("20060102")
		str2 := d1.Format("2006-01-02")

		ret1 = append(ret1, str1)
		ret2 = append(ret2, str2)

	}
	// return int(d2.Sub(d1) / (24 * time.Hour))
	return ret1, ret2

}
