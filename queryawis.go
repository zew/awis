package main

import (
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kataras/iris"

	"github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/logx"
	"github.com/zew/awis/util"
)

func queryawis(c *iris.Context) {

	var err error

	/*
	   $xml = new SimpleXMLElement($response,null,false,
	                               'http://awis.amazonaws.com/doc/2005-07-11');
	   if($xml->count() && $xml->Response->UrlInfoResult->Alexa->count()) {
	       $info = $xml->Response->UrlInfoResult->Alexa;
	       $nice_array = array(
	           'Phone Number'   => $info->ContactInfo->PhoneNumbers->PhoneNumber,
	           'Owner Name'     => $info->ContactInfo->OwnerName,
	           'Email'          => $info->ContactInfo->Email,
	           'Street'         => $info->ContactInfo->PhysicalAddress->Streets->Street,
	           'City'           => $info->ContactInfo->PhysicalAddress->City,
	           'State'          => $info->ContactInfo->PhysicalAddress->State,
	           'Postal Code'    => $info->ContactInfo->PhysicalAddress->PostalCode,
	           'Country'        => $info->ContactInfo->PhysicalAddress->Country,
	           'Links In Count' => $info->ContentData->LinksInCount,
	           'Rank'           => $info->TrafficData->Rank
	       );
	   }

	*/

	myUrl := url.URL{}
	myUrl.Host = ServiceHost1
	myUrl.Scheme = "http"
	logx.Printf("host is %v", myUrl.String())

	vals := map[string]string{
		"Action":           "TopSites",
		"AWSAccessKeyId":   util.EnvVar("AWS_ACCESS_KEY_ID"),
		"SignatureMethod":  "HmacSHA256",
		"SignatureVersion": "2",
		"Timestamp":        getTimestamp(),
		// "Signature" : "will be added by awsauth.Sign2(req)"
		"ResponseGroup": "Country",
		"Url":           util.EffectiveParam(c, "Url", "wwww.zew.de"),
		"CountryCode":   util.EffectiveParam(c, "CountryCode", "DE"),
		"Start":         util.EffectiveParam(c, "Start", "0"),
		"Count":         util.EffectiveParam(c, "Count", "5"),
	}

	vals2 := map[string]string{
		"Action1":           "UrlInfo",
		"Action2":           "SitesLinkingIn",
		"ResponseGroup":     "SitesLinkingIn",
		"ResponseGroupName": "Rank,ContactInfo,LinksInCount",
	}
	_ = vals2

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
	var reqSigned *http.Request
	reqSigned = req

	resp, err := client().Do(reqSigned)
	util.CheckErr(err)
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	target := html.EscapeString(string(respBytes))
	util.CheckErr(err)

	// xml.NewDecoder(resp.Body).Decode(target)

	//
	//
	//
	//

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
		HTMLTitle:        AppName() + " result",
		Title:            AppName() + " result",
		FlashMsg:         template.HTML("Alexa Web Information Service"),
		JSON:             template.HTML(target),
		URL:              reqSigned.URL.String(),
		ParamUrl:         util.EffectiveParam(c, "Url", "www.zew.de"),
		ParamStart:       util.EffectiveParam(c, "Start", "0"),
		ParamCount:       util.EffectiveParam(c, "Count", "5"),
		ParamCountryCode: util.EffectiveParam(c, "CountryCode", "DE"),
	}

	err = c.Render("index.html", s)
	util.CheckErr(err)

}
