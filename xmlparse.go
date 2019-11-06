package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"

	// "launchpad.net/xmlpath"
	"gopkg.in/xmlpath.v2"

	"github.com/kataras/iris/v12"

	"github.com/zew/util"
)

func xmlparse(c iris.Context) {
	dat, err := ioutil.ReadFile("top_sites_example.xml")
	util.CheckErr(err)

	sites, err := ParseIntoDomains(dat)
	if err != nil {
		c.WriteString(err.Error())
		return
	}

	c.WriteString("xml parsed into structs\n")
	display := util.IndentedDump(sites)
	c.WriteString(display)
}

// XmlPathDemo() retrieves some deeply nested nodes.
// Each node could
//
func XmlPathDemo(c iris.Context) {

	path := xmlpath.MustCompile("//Site")
	path = xmlpath.MustCompile("/TopSitesResponse/Response/TopSitesResult/Alexa/TopSites/Country/Sites/Site")

	file, err := os.Open("top_sites_example.xml")
	util.CheckErr(err)
	root, err := xmlpath.Parse(file)
	if err != nil {
		c.WriteString(err.Error())
	}

	if path.Exists(root) {
		c.WriteString("path exists\n")

		if false {
			if value, ok := path.String(root); ok {
				c.WriteString(value + "\n")
			}
			if subBytes, ok := path.Bytes(root); ok {
				c.Write(subBytes)
			}
		}
		nodes := path.Iter(root)
		c.WriteString("Nodes are there\n")
		for nodes.Next() {
			node := nodes.Node()
			str1 := fmt.Sprintf("\n\nNode is %+v  \n\n", node)
			c.WriteString(str1)
		}

	} else {
		c.WriteString("path NOT there\n")
	}

}

// ParseDemo shows how to omit the outmost tag.
// It shows how to read into a slice
// And it shows how to "deep link" with tag1>tag2 syntax.
func ParseDemo(c iris.Context) {
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
		c.WriteString(err.Error())
	}
	vs := util.IndentedDump(v)
	c.WriteString(vs)

}
