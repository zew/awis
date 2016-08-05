package config

import (
	"encoding/json"

	"github.com/zew/gorpx"
	"github.com/zew/logx"
	"github.com/zew/util"
)

type ConfigT struct {
	Email        string         `json:"email"`
	VersionMajor int            `json:"version_major"`
	VersionMinor int            `json:"version_minor"`
	AppName      string         `json:"app_name"`
	SQLite       bool           `json:"sql_lite"`
	SQLHosts     gorpx.SQLHosts `json:"sql_hosts"`
}

var Config ConfigT

func init() {

	for _, v := range []string{"SQL_PW"} {
		util.EnvVar(v)
	}

	fileReader := util.LoadConfig()
	decoder := json.NewDecoder(fileReader)
	err := decoder.Decode(&Config)
	util.CheckErr(err)
	logx.Printf("\n%#s", util.IndentedDump(Config))

}
