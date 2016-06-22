package main

import (
	"html/template"

	"github.com/kataras/iris"
	"github.com/zew/awis/util"
)

func index(c *iris.Context) {

	var err error
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
		HTMLTitle: AppName() + " main",
		Title:     AppName() + " main",
		FlashMsg:  template.HTML("Alexa Web Information Service"),
		// JSON:      template.HTML(target),
		// URL:       reqSigned.URL.String(),
		ParamUrl:         util.EffectiveParam(c, "Url", "www.zew.de"),
		ParamStart:       util.EffectiveParam(c, "Start", "0"),
		ParamCount:       util.EffectiveParam(c, "Count", "5"),
		ParamCountryCode: util.EffectiveParam(c, "CountryCode", "DE"),
	}

	err = c.Render("index.html", s)
	util.CheckErr(err)

}
