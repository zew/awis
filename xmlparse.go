package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	// "launchpad.net/xmlpath"
	"gopkg.in/xmlpath.v2"

	"github.com/kataras/iris"

	"github.com/zew/awis/util"
)

type Site struct {
	DataUrl                    string
	GlobalRank                 int     `xml:"Global>Rank"`
	CountryRank                int     `xml:"Country>Rank"`
	CountryReachPerMillion     int     `xml:"Country>Reach>PerMillion"`
	CountryPageViewsPerMillion int     `xml:"Country>PageViews>PerMillion"`
	CountryPageViewsPerUser    float64 `xml:"Country>PageViews>PerUser"`
}
type Result struct {
	// Sites []Site `xml:"TopSitesResponse>Response>TopSitesResult>Alexa>TopSites>Country>Sites>Site"`
	Sites []Site `xml:"Response>TopSitesResult>Alexa>TopSites>Country>Sites>Site"` // omit the outmost tag name TopSitesResponse
}

func xmlparse(c *iris.Context) {

	dat, err := ioutil.ReadFile("example1.xml")
	util.CheckErr(err)
	// c.Text(200, string(dat))

	sites := Result{}
	err = xml.Unmarshal(dat, &sites)
	if err != nil {
		c.Text(200, err.Error())
	}
	res := util.IndentedDump(sites)
	c.Text(200, res)

}

//
//
func xmlparseExamplePath(c *iris.Context) {

	path := xmlpath.MustCompile("//Alexa")
	path = xmlpath.MustCompile("/TopSitesResponse/Response/TopSitesResult/Alexa/TopSites/Country/Sites/Site")

	file, err := os.Open("example2.xml")
	util.CheckErr(err)
	root, err := xmlpath.Parse(file)
	if err != nil {
		c.Text(200, err.Error())
	}

	if path.Exists(root) {
		c.Text(200, "path exists\n")

		if false {
			if value, ok := path.String(root); ok {
				c.Text(200, value+"\n")
			}
			if subBytes, ok := path.Bytes(root); ok {
				c.RequestCtx.Write(subBytes)
			}
		}
		nodes := path.Iter(root)
		c.Text(200, "Nodes are there\n")
		for nodes.Next() {
			node := nodes.Node()
			str1 := fmt.Sprintf("\n\nNode is %+v  \n\n", node)
			c.Text(200, str1)
		}

	} else {
		c.Text(200, "path NOT there\n")
	}

}

//
//
func xmlparseExampleSimple(c *iris.Context) {
	type Email struct {
		Addr string
	}
	type Result1 struct {
		FullName string
		Email    []Email
		Groups   []string `xml:"Group>Value"`
	}
	v := Result1{}
	data := `
   			<Person>
   				<FullName>Grace R. Emlin</FullName>
   				<Email>
   					<Addr>gre@example.com</Addr>
   				</Email>
   				<Email>
   					<Addr>gre@work.com</Addr>
   				</Email>
   				<Group>
   					<Value>Friends</Value>
   					<Value>Squash</Value>
   				</Group>
   				<City>Hanga Roa</City>
   				<State>Easter Island</State>
   			</Person>
   		`
	err := xml.Unmarshal([]byte(data), &v)
	if err != nil {
		c.Text(200, err.Error())
	}
	vs := util.IndentedDump(v)
	c.Text(200, vs)

}
