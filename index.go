package main

import (
	"github.com/kataras/iris"
	"github.com/zew/util"
)

func index(c *iris.Context) {

	links := map[string]string{
		"Top Sites":  PathTopSites,
		"Site Infos": PathDomainInfo,
	}

	var err error
	s := struct {
		HTMLTitle string
		Title     string
		Links     map[string]string
	}{
		HTMLTitle: AppName() + " main",
		Title:     AppName() + " main",
		Links:     links,
	}

	err = c.Render("index.html", s)
	util.CheckErr(err)

}
