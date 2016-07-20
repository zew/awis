package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/zew/logx"
	"github.com/zew/util"
)

type SQLHost struct {
	User             string            `json:"user"`
	Host             string            `json:"host"`
	Port             string            `json:"port"`
	DBName           string            `json:"db_name"`
	ConnectionParams map[string]string `json:"connection_params"`
}

type Config3 struct {
	Email        string             `json:"email"`
	VersionMajor int                `json:"version_major"`
	VersionMinor int                `json:"version_minor"`
	AppName      string             `json:"app_name"`
	SQLite       bool               `json:"sql_lite"`
	SQLHosts     map[string]SQLHost `json:"sql_hosts"`
}

var Config Config3

func init() {

	for _, v := range []string{"SQL_PW", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"} {
		util.EnvVar(v)
	}

	pwd, err := os.Getwd()
	util.CheckErr(err)
	fmt.Println("pwd2: ", pwd)

	_, srcFile, _, ok := runtime.Caller(1)
	if !ok {
		logx.Fatalf("runtime caller not found")
	}

	{
		// file, err := os.Open(filepath.Join(pwd, "/config/config1.json"))
		// file, err := os.Open("./config/config3.json")
		file, err := os.Open(path.Join(path.Dir(srcFile), "config3.json"))
		util.CheckErr(err)
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&Config)
		util.CheckErr(err)
		// logx.Printf("%#v", conf3)
		logx.Printf("\n%#s", util.IndentedDump(Config))
	}

}
