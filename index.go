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

var ActionName = "UrlInfo"
var ActionName2 = "SitesLinkingIn"

var AWSAccessKeyId = util.EnvVar("AWS_ACCESS_KEY_ID")
var SecretAccessKey = util.EnvVar("AWS_SECRET_ACCESS_KEY")

var ResponseGroupName = "Rank,ContactInfo,LinksInCount"
var ServiceHost = "awis.amazonaws.com"
var NumReturn = "10"
var StartNum = "1"
var SigVersion = "2"
var SignatureMethod = "HmacSHA256" // hash algo
var site = "zew.de"

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


	   // Get site info from AWIS.
	   public function getUrlInfo() {
	       $queryParams = $this->buildQueryParams();
	       $sig = $this->generateSignature($queryParams);
	       $url = 'http://' . self::$ServiceHost . '/?' . $queryParams . '&Signature=' . $sig;
	       $ret = self::makeRequest($url);
	       echo "\nResults for " . $this->site .":\n\n";
	       self::parseResponse($ret);
	   }


	   // Makes request to AWIS
	   // @param String $url   URL to make request to
	   // @return String       Result of request
	   protected static function makeRequest($url) {
	       echo "\nMaking request to:\n$url\n";
	       $ch = curl_init($url);
	       curl_setopt($ch, CURLOPT_TIMEOUT, 4);
	       curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
	       $result = curl_exec($ch);
	       curl_close($ch);
	       return $result;
	   }

	   // Parses XML response from AWIS and displays selected data
	   // @param String $response    xml response from AWIS
	   public static function parseResponse($response) {
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

	   //Generates an HMAC signature per RFC 2104.
	   //@param String $url       URL to use in createing signature
	   protected function generateSignature($url) {
	       $sign = "GET\n" . strtolower(self::$ServiceHost) . "\n/\n". $url;
	       echo "String to sign: \n" . $sign . "\n";
	       $sig = base64_encode(hash_hmac('sha256', $sign, $this->SecretAccessKey, true));
	       echo "\nSignature: " . $sig ."\n";
	       return rawurlencode($sig);
	   }



	*/

	// c.RequestCtx.WriteString("aaa<br>\n")

	// Builds query parameters for the request to AWIS.
	// Parameter names will be in alphabetical order and
	// parameter values will be urlencoded per RFC 3986.
	// @return String query parameters for the request

	myUrl := url.URL{}
	myUrl.Host = ServiceHost
	myUrl.Scheme = "http"
	logx.Printf("host is %v", myUrl.String())

	vals := map[string]string{
		"Action":         "SitesLinkingIn",
		"AWSAccessKeyId": util.EnvVar("AWS_ACCESS_KEY_ID"),
		// "Signature" : ""
		"SignatureMethod":  "HmacSHA256",
		"SignatureVersion": "2",
		"Timestamp":        getTimestamp(),
		"Url":              "www.spiegel.de",
		"ResponseGroup":    "SitesLinkingIn",
		"Count":            "10",
		"Start":            "1",
	}
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

	var reqSigned *http.Request
	if false {
		reqSigned = awsauth.Sign2(req, awsauth.Credentials{
			AccessKeyID:     util.EnvVar("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: util.EnvVar("AWS_SECRET_ACCESS_KEY"),
			// SecurityToken:   "Security Token", // STS (optional)
		})

	} else {
		reqSigned = awsauth.Sign2(req)
	}

	if reqSigned == nil {
		logx.Printf("what is this return value DOING?")
		reqSigned = req
	}

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
	resp, err := netClient.Do(reqSigned)
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
