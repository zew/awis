package main

import (
	"github.com/kataras/iris"

	appcfg "github.com/zew/awis/config"
	"github.com/zew/awis/mdl"
	"github.com/zew/gorpx"
	"github.com/zew/irisx"
	"github.com/zew/logx"
	"github.com/zew/util"
)

// The name of the application
func AppName(p ...string) string {
	s := appcfg.Config.AppName
	if len(p) > 0 {
		return s + p[0]
	}
	return s
}

const (
	PathTopSites                   = "/top-sites"
	TrafficHistory                 = "/top-sites-batched"
	TrafficHistoryFillMissingHoles = "/traffic-history-missing-holes"
	PathDomainInfo                 = "/domain-info"
)

func main() {

	i01 := iris.New(irisBaseConfig())
	irisInctanceConfig(i01)

	var keysToPersist = map[string]string{
		"country": "DE",
	}
	irisx.ConfigSession(keysToPersist)

	i01.Static(Pref("/js"), "./static/js/", 2)
	// i01.Static("/js", "./static/js/", 1)
	i01.Static(Pref("/img"), "./static/img/", 2)
	i01.Static(Pref("/css"), "./static/css/", 2)

	i01.Get("/", index)
	i01.Get(Pref(""), index)
	i01.Get(Pref("/"), index)

	i01.Get(Pref(PathDomainInfo), awisDomainInfo)
	i01.Get(Pref(PathTopSites), topSites)
	i01.Get(Pref(TrafficHistory), trafficHistory)
	i01.Get(Pref(TrafficHistoryFillMissingHoles), trafficHistoryFillMissingHoles)
	i01.Get(Pref("/xmlparse"), xmlparse)

	//
	//
	logx.Printf("setting up mysql server...")
	gorpx.DB(appcfg.Config.SQLHosts)
	defer gorpx.DB().Close()

	DDL()

	gorpx.DBMapAddTable(mdl.Domain{})
	gorpx.DBMapAddTable(mdl.Meta{})
	gorpx.DBMapAddTable(mdl.Rank{})
	gorpx.DBMapAddTable(mdl.Category{})
	gorpx.DBMapAddTable(mdl.Delta{})
	gorpx.DBMapAddTable(mdl.History{})

	logx.Printf("starting http server...")
	i01.Listen(":8081")

}
func DDL() {

	var err error

	{
		mp := gorpx.IndependentDbMapper()
		t := mp.AddTable(mdl.Domain{})
		// t.ColMap("domain_name").SetUnique(true)
		t.SetUniqueTogether("domain_name", "last_updated")
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			err = mp.CreateIndex()
			util.CheckErr(err)
		}
	}

	{
		mp := gorpx.IndependentDbMapper()
		t := mp.AddTable(mdl.Meta{})
		t.ColMap("domain_name").SetUnique(true)
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			err = mp.CreateIndex()
			util.CheckErr(err)
		}
	}

	{
		mp := gorpx.IndependentDbMapper()
		t := mp.AddTable(mdl.Rank{})
		// t.ColMap("domain_name").SetUnique(true)
		// t.AddIndex("idx_name_desc", "Btree", []string{"domain_name", "rank_code"})
		t.SetUniqueTogether("domain_name", "last_updated", "rank_code")
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			err = mp.CreateIndex()
			util.CheckErr(err)
		}
	}

	{
		mp := gorpx.IndependentDbMapper()
		t := mp.AddTable(mdl.Category{})
		t.SetUniqueTogether("domain_name", "category_path")
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			err = mp.CreateIndex()
			util.CheckErr(err)
		}
	}

	{
		mp := gorpx.IndependentDbMapper()
		t := mp.AddTable(mdl.Delta{})
		t.SetUniqueTogether("domain_name", "last_updated", "months", "days")
		t.AddIndex("idx_domain_name", "Btree", []string{"domain_name", "months", "days"})
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			err = mp.CreateIndex()
			util.CheckErr(err)
		}
	}

	{
		mp := gorpx.IndependentDbMapper()
		t := mp.AddTable(mdl.History{})
		t.SetUniqueTogether("domain_name", "date")
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			err = mp.CreateIndex()
			util.CheckErr(err)
		}
	}

}
