package main

import (
	"github.com/kataras/iris/v12"

	appcfg "github.com/zew/awis/config"
	"github.com/zew/awis/mdl"
	"github.com/zew/gorpx"
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
	PathTopSitesAuto               = "/top-sites-auto"
	TrafficHistory                 = "/top-sites-batched"
	TrafficHistoryFillMissingHoles = "/traffic-history-missing-holes"
	PathDomainInfo                 = "/domain-info"
)

func main() {
	i01 := iris.New()
	irisInctanceConfig(i01)
	irisSessionsConfig(i01)

	i01.HandleDir(Pref("/js"), "./static/js/")
	// i01.Static("/js", "./static/js/", 1)
	i01.HandleDir(Pref("/img"), "./static/img/")
	i01.HandleDir(Pref("/css"), "./static/css/")

	i01.Get("/", index)
	i01.Get(Pref(""), index)
	i01.Get(Pref("/"), index)

	i01.Get(Pref(PathDomainInfo), awisDomainInfo)
	i01.Get(Pref(PathTopSites), topSites)
	i01.Get(Pref(PathTopSitesAuto), topSitesAuto)
	i01.Get(Pref(TrafficHistory), trafficHistory)
	i01.Get(Pref(TrafficHistoryFillMissingHoles), trafficHistoryFillMissingHoles)
	i01.Get(Pref("/xmlparse"), xmlparse)

	//
	//
	logx.Printf("setting up mysql server...")
	gorpx.SetAndInitDatasourceId(appcfg.Config.SQLHosts, 0)
	defer gorpx.DbClose()

	DDL()

	gorpx.DbMap().AddTable(mdl.Domain{})
	gorpx.DbMap().AddTable(mdl.Meta{})
	gorpx.DbMap().AddTable(mdl.Rank{})
	gorpx.DbMap().AddTable(mdl.Category{})
	gorpx.DbMap().AddTable(mdl.Delta{})
	gorpx.DbMap().AddTable(mdl.History{})

	logx.Printf("starting http server...")
	i01.Run(iris.Addr(":8081"), irisBaseConfig)
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
		t.SetUniqueTogether("domain_name", "for_date")
		err = mp.CreateTables()
		if err != nil {
			logx.Printf("error creating table: %v", err)
		} else {
			err = mp.CreateIndex()
			util.CheckErr(err)
		}
	}

}
