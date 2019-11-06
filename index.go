package main

import (
	"github.com/kataras/iris/v12"
	"github.com/zew/util"
)

func index(c iris.Context) {

	links := map[string]string{
		"Site Infos":                PathDomainInfo,
		"Top Sites":                 PathTopSites,
		"Traffic History":           TrafficHistory,
		"Löcher in Traffic History": TrafficHistoryFillMissingHoles,
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

	err = c.View("index.html", s)
	util.CheckErr(err)
}
