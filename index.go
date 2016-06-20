package main

import (
	"html/template"

	"github.com/kataras/iris"
	"github.com/zew/awis/util"
)

func index(c *iris.Context) {

	// c.RequestCtx.WriteString("aaa<br>\n")

	s := struct {
		HTMLTitle string
		Title     string
		FlashMsg  template.HTML
	}{
		HTMLTitle: AppName() + " main",
		Title:     AppName() + " main",
		FlashMsg:  template.HTML("Alexa Web Information Service"),
	}

	err := c.Render("index.html", s)
	util.CheckErr(err)

}
