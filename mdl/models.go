package mdl

type Site struct {
	Id                         int     `db:"domain_id, primarykey, autoincrement"`
	Name                       string  `xml:"DataUrl" db:"domain_name, size:200, not null"` // unique with SetUnique
	Label                      string  `db:"domain_label, size:200, not null"`
	Description                string  `db:"domain_description, size:200, not null"`
	GlobalRank                 int     `xml:"Global>Rank" db:"global_rank, not null"`
	CountryRank                int     `xml:"Country>Rank" db:"country_rank, not null"`
	CountryReachPerMillion     float64 `xml:"Country>Reach>PerMillion" db:"country_reach_permillion, not null"`
	CountryPageViewsPerMillion float64 `xml:"Country>PageViews>PerMillion" db:"country_pageviews_permillion, not null"`
	CountryPageViewsPerUser    float64 `xml:"Country>PageViews>PerUser" db:"country_pageviews_peruser, not null"`
}

//
//
type Detail struct {
	Id           int    `db:"detail_id, primarykey, autoincrement"`
	DomainId     int    `db:"domain_id, not null"`
	Name         string `db:"detail_name, size:200, not null"`
	Label        string `db:"detail_label, size:200, not null"`
	Description  string `db:"detail_description, size:200, not null"`
	Type         string `db:"detail_type, not null, server default:float"`
	CXLabel      int    `db:"detail_cx_label, not null, server default:150"`
	CXControl    int    `db:"detail_cx_control, not null, server default:150"`
	RenderMethod string `db:"detail_render_method, size:200, not null, server default:input"`
}
