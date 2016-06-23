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

type Meta struct {
	Id          int    `db:"meta_id, primarykey, autoincrement"`
	Name        string `xml:"ContactInfo>DataUrl" db:"domain_name, size:200, not null"`                       // unique with SetUnique
	PhoneNumber string `xml:"ContactInfo>PhoneNumbers>PhoneNumber" db:"meta_phonenumber, size:200, not null"` // multiple
	OwnerName   string `xml:"ContactInfo>OwnerName" db:"meta_ownername, size:200, not null"`
	Email       string `xml:"ContactInfo>Email" db:"meta_email, size:200, not null"`
	Street      string `xml:"ContactInfo>PhysicalAddress>Streets>Street" db:"meta_street, size:200, not null"` // multiple
	City        string `xml:"ContactInfo>PhysicalAddress>City" db:"meta_city, size:200, not null"`
	Country     string `xml:"ContactInfo>PhysicalAddress>Country" db:"meta_country, size:200, not null"`

	Title        string `xml:"ContentData>SiteData>Title" db:"meta_title, size:200, not null"`
	Description  string `xml:"ContentData>SiteData>Description" db:"meta_description, size:1200, not null"`
	OnlineSince  string `xml:"ContentData>SiteData>OnlineSince" db:"meta_onlinesince, size:1200, not null"`
	AdultContent string `xml:"ContentData>AdultContent" db:"meta_adultcontent, size:10, not null"`

	Locale   string `xml:"ContentData>Language>Locale" db:"meta_locale, size:10, not null"`
	Encoding string `xml:"ContentData>Language>Encoding" db:"meta_encoding, size:20, not null"`

	CategoryTitle string `xml:"Related>Categories>CategoryData>Title" db:"meta_categorytitle, size:200, not null"`
	CategoryPath  string `xml:"Related>Categories>CategoryData>AbsolutePath" db:"meta_categorypath, size:200, not null"`

	// CharData string `xml:"TrafficData>RankByCountry,chardata" db:"meta_rankchars, size:600, not null"`
	// RankByCountry string `xml:"TrafficData>RankByCountry>Country>Code,attr" db:"meta_rankbycountry, size:50, not null"`
	Ranks []TRank `xml:"TrafficData>RankByCountry>Country" db:"meta_xxx, size:600, not null"`
}

type TRank struct {
	Value string `xml:",chardata"`

	Code      string `xml:"Code,attr"`
	Rank      string `xml:"Rank"`
	PageViews string `xml:"Contribution>PageViews"`
	Users     string `xml:"Contribution>Users"`
}
