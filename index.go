package main

import (
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kataras/iris"
	"github.com/smartystreets/go-aws-auth"
	"github.com/zew/awis/logx"
	"github.com/zew/awis/util"
)

var AWSAccessKeyId = util.EnvVar("AWS_ACCESS_KEY_ID")
var SecretAccessKey = util.EnvVar("AWS_SECRET_ACCESS_KEY")

var ResponseGroupName = "Rank,ContactInfo,LinksInCount"
var ServiceHost = "awis.amazonaws.com"

var sess *session.Session

func init() {
	sess = session.New(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
}

// Builds current ISO8601 timestamp.
// eg 2007-08-31T16:47+00:00
func getTimestamp() string {
	t := time.Now()
	ts := t.Format("2006-01-02T15:04:05")
	ts += ".0000Z"
	// return fmt.Println(t.Format("20060102150405"))
	// return gmdate("Y-m-d\TH:i:s.\\0\\0\\0\\Z", time())
	logx.Printf("ts is %v", ts)
	return ts
}

func index(c *iris.Context) {

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
	       foreach($nice_array as $k => $v) {
	           echo $k . ': ' . $v ."\n";
	       }
	   }


	*/

	myUrl := url.URL{}
	myUrl.Host = ServiceHost
	myUrl.Scheme = "http"
	logx.Printf("host is %v", myUrl.String())

	vals := map[string]string{
		"Action":           "SitesLinkingIn",
		"AWSAccessKeyId":   util.EnvVar("AWS_ACCESS_KEY_ID"),
		"SignatureMethod":  "HmacSHA256",
		"SignatureVersion": "2",
		"Timestamp":        getTimestamp(),
		// "Signature" : "will be added by awsauth.Sign2(req)"
		"Url":           "www.spiegel.de",
		"ResponseGroup": "SitesLinkingIn",
		"Count":         "10",
		"Start":         "0",
	}

	vals2 := map[string]string{
		"Action":            "UrlInfo",
		"ResponseGroupName": "Rank,ContactInfo,LinksInCount",
	}
	_ = vals2

	queryStr := ""
	for k, v := range vals {
		queryStr += fmt.Sprintf("%v=%v&", k, v)
	}
	logx.Printf("queryStr is %v", queryStr)

	strUrl := myUrl.String() + "/?" + queryStr

	// privKey := rsa.PrivateKey{}
	// signer := sign.NewURLSigner(util.EnvVar("ACCESS_KEY_ID"), &privKey)
	// strUrlSigned, err := signer.Sign(strUrl, time.Now())

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
		URL       string
		JSON      template.HTML
	}{
		HTMLTitle: AppName() + " main",
		Title:     AppName() + " main",
		FlashMsg:  template.HTML("Alexa Web Information Service"),
		JSON:      template.HTML(target),
		URL:       reqSigned.URL.String(),
	}

	err = c.Render("index.html", s)
	util.CheckErr(err)

}

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
