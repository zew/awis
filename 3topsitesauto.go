package main

import (
	"fmt"
	"time"

	"github.com/kataras/iris"
	"github.com/zew/logx"
	"github.com/zew/util"
)

func topSitesAuto(c *iris.Context) {

	starts := []string{
		"0", "100", "200", "300", "400", "500", "600", "700", "800", "900",
		"1000", "1100", "1200", "1300", "1400", "1500", "1600", "1700", "1800", "1900",
		// "2000",
	}

	now := time.Now()
	day := now.YearDay() + 1

	justOne := false

	if day%5 != 0 {
		logx.Printf("%v is not a multiple of 5", day)
		return
		starts = []string{"0"}
		justOne = true
	}

	for _, v := range starts {
		url1 := "http://localhost:8081/alexa_web_information_service/top-sites"
		keys := []string{"Start", "Count", "CountryCode", "submit"}
		vals := []string{v, "100", "DE", "+Submit+"}
		if v == "2000" || justOne {
			vals[1] = "1"
		}
		bytes, err := util.Request("GET", url1, keys, vals)
		_ = bytes
		if err != nil {
			c.WriteString(fmt.Sprintf("err %v - %v %v\n", err, url1, vals))
			return
		}
		// c.WriteString(fmt.Sprintf("%s", bytes))
		c.WriteString(fmt.Sprintf("success - %s\n", vals))

		time.Sleep(2500 * time.Millisecond)
	}
}
