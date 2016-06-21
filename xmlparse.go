package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	// "launchpad.net/xmlpath"
	"gopkg.in/xmlpath.v2"

	"github.com/kataras/iris"

	"github.com/zew/awis/util"
)

type TGlobal struct {
	Rank int `xml:aws:Rank`
}
type TPageViews struct {
	PerMillion int     `xml:PerMillion`
	PerUser    float64 `xml:PerUser`
}
type TReach struct {
	PerMillion int `xml:PerMillion`
}

type TCountry struct {
	Rank      int        `xml:Rank`
	Reach     TReach     `xml:Reach`
	PageViews TPageViews `xml:PageViews`
}

type TSite struct {
	DataUrl string   `xml:DataUrl`
	Country TCountry `xml:Country`
	Global  TGlobal  `xml:Global`
}

type TSites struct {
	Sites []TSite `xml:Sites`
}

func xmlparse(c *iris.Context) {
	dat, err := ioutil.ReadFile("example.xml")
	util.CheckErr(err)

	// c.Text(200, string(dat))

	rdr := bytes.NewReader(dat)
	sites := TSites{}

	err = xml.NewDecoder(rdr).Decode(&sites)

	if err != nil {
		c.Text(200, err.Error())
	} else {
		c.Text(200, "suck sess - in decoding<br>\n")
	}
	res := util.IndentedDump(sites)
	c.Text(200, res)
	c.Text(200, "<br>\n<br>\n")

	path := xmlpath.MustCompile("/library/book/isbn")
	path = xmlpath.MustCompile("//Alexa")
	path = xmlpath.MustCompile("/TopSitesResponse/Response/TopSitesResult/Alexa/TopSites")
	path = xmlpath.MustCompile("/TopSitesResponse/Response/TopSitesResult/Alexa/TopSites/Country")
	path = xmlpath.MustCompile("/TopSitesResponse/Response/TopSitesResult/Alexa/TopSites/Country/Sites/Site")

	file, err := os.Open("example2.xml")
	util.CheckErr(err)
	root, err := xmlpath.Parse(file)
	if err != nil {
		c.Text(200, err.Error())
	}
	_ = root
	_ = path

	if path.Exists(root) {
		c.Text(200, "path exists<br>\n")

		if false {
			if value, ok := path.String(root); ok {
				c.Text(200, value+"<br>\n")
			}
			if subBytes, ok := path.Bytes(root); ok {
				c.RequestCtx.Write(subBytes)
			}
		}
		nodes := path.Iter(root)
		c.Text(200, "Nodes are there<br>\n")
		for nodes.Next() {
			node := nodes.Node()
			str1 := fmt.Sprintf("Node is %v  <br>\n", node)
			c.Text(200, str1+"<br>\n")

			rdr := bytes.NewReader(node.Bytes())
			site := TSite{}
			err = xml.NewDecoder(rdr).Decode(&site)
			res := util.IndentedDump(site)
			c.Text(200, res)

		}

	} else {
		c.Text(200, "path NOT there<br>\n")
	}
}
