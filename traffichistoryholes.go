package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
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
	_ = start
	count, _, _ := irisx.EffectiveParamInt(c, "Count", 5)
	granularity, _, _ := irisx.EffectiveParamInt(c, "Granularity", 10)
	dateBegin := irisx.EffectiveParam(c, "DateBegin", "20150101")

	sites := []mdl.Site{
		{Name: "7tv.de"},
		{Name: "advertiserurl.com"},
		{Name: "adexc.net"},
		{Name: "alexa.com"},
		{Name: "bidverdrd.com"},
		{Name: "cine.to"},
		{Name: "csgohouse.com"},
		{Name: "dirty-time.net"},
		{Name: "dw.com"},
		{Name: "ebay-kleinanzeigen.de"},
		{Name: "eurowings.com"},
		{Name: "freepornx.org"},
		{Name: "futwatch.com"},
		{Name: "heftig.de"},
		{Name: "hespress.com"},
		{Name: "henkel-lifetimes.de"},
		{Name: "hotmovs.com"},
		{Name: "just4single.com"},
		{Name: "lernhelfer.de"},
		{Name: "moneyhouse.de"},
		{Name: "nurxxx.mobi"},
		{Name: "playoverwatch.com"},
		{Name: "pussyspace.com"},
		{Name: "rock-am-ring.com"},
		{Name: "spotscenered.info"},
		{Name: "tvnow.de"},
		{Name: "wahnsinn.tv"},
		{Name: "wiocha.pl"},
	}

	logx.Printf("sites are %v", sites)

	for _, site := range sites {

		for i := 0; i < count; i += granularity {

			allSteps := dayStepsFromString(dateBegin, i)
			lastStep := allSteps[len(allSteps)-1]

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
				"Start":       lastStep,
				"Range":       fmt.Sprintf("%v", granularity),
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

			type TrafHistories struct {
				TrafficHistories []mdl.TrafficHistory `xml:"Response>TrafficHistoryResult>Alexa>TrafficHistory>HistoricalData>Data"`
			}
			trafHists := TrafHistories{}
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

		FormAction: TrafficHistory,

		ParamCount:     irisx.EffectiveParam(c, "Granularity", "10"),
		ParamDateBegin: irisx.EffectiveParam(c, "DateBegin", "20150101"),

		StructDump: template.HTML(display),
	}

	err = c.Render("traffic-history.html", s)
	util.CheckErr(err)

}

// usage:
// fmt.Printf("%+v", dayStepsFromString("20150125",10) )
func dayStepsFromString(strDate string, numberOfDays int) []string {

	date, err := time.Parse("20060102", strDate)

	if err != nil {
		logx.Fatal("we want format '20060102' ")
	}

	return daySteps(date.Year(), date.Month(), date.Day(), numberOfDays)
}

// usage:
// fmt.Printf("%+v", daySteps(2015,01,25,10) )
func daySteps(
	year1 int, month1 time.Month, day1 int,
	numberOfDays int,
) []string {

	ret := make([]string, 0, numberOfDays)

	for i := 0; i < numberOfDays; i++ {

		d1 := time.Date(
			year1, month1, day1+i,
			0, 0, 0, 0, time.UTC,
		)
		str1 := d1.Format("2006-01-02")
		str1 = d1.Format("20060102")

		ret = append(ret, str1)

	}
	// return int(d2.Sub(d1) / (24 * time.Hour))
	return ret
}
