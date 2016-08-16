package mdl

type Domain struct {
	Id                         int     `db:"domain_id, primarykey, autoincrement"`
	Name                       string  `xml:"DataUrl" db:"domain_name, size:200, not null"` // unique with SetUnique
	LastUpdated                int     `db:"last_updated, not null"`
	GlobalRank                 int     `xml:"Global>Rank" db:"global_rank, not null"`
	CountryRank                int     `xml:"Country>Rank" db:"country_rank, not null"`
	CountryReachPerMillion     float64 `xml:"Country>Reach>PerMillion" db:"country_reach_permillion, not null"`
	CountryPageViewsPerMillion float64 `xml:"Country>PageViews>PerMillion" db:"country_pageviews_permillion, not null"`
	CountryPageViewsPerUser    float64 `xml:"Country>PageViews>PerUser" db:"country_pageviews_peruser, not null"`
}

type Meta struct {
	Id          int    `db:"meta_id, primarykey, autoincrement"`
	Name        string `xml:"ContactInfo>DataUrl" db:"domain_name, size:200, not null"` // unique with SetUnique
	LastUpdated int    `db:"last_updated, not null"`
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

	// Into separate struct - 1:n slice:
	// Ranks      []Rank         `xml:"TrafficData>RankByCountry>Country"`
	// Categories []mdl.Category `xml:"Response>UrlInfoResult>Alexa>Related>Categories"`

	/*
	     <LinksInCount>159636</LinksInCount>
	     <OwnedDomains>
	       <OwnedDomain>
	         <Domain>gogoole.de</Domain>
	         <Title>gogoole.de</Title>
	       </OwnedDomain>
	       <OwnedDomain>
	         <Domain>goldengate-marken.de</Domain>
	         <Title>goldengate-marken.de</Title>
	       </OwnedDomain>


	   <Related>
	     <DataUrl type="canonical">google.de</DataUrl>
	     <RelatedLinks>
	       <RelatedLink>
	         <DataUrl type="canonical">web.de/</DataUrl>
	       </RelatedLink>
	       <RelatedLink>
	         <DataUrl type="canonical">ebay.de/</DataUrl>


	     <UsageStatistics>
	       <UsageStatistic>
	         <TimeRange>
	           <Months>3</Months>
	         </TimeRange>
	         <Rank>
	           <Value>25</Value>
	           <Delta>0</Delta>
	         </Rank>


	     <ContributingSubdomains>
	       <ContributingSubdomain>
	         <DataUrl>images.google.de</DataUrl>
	         <DataUrl>translate.google.de</DataUrl>

	*/

}

type Category struct {
	Id          int    `db:"category_id, primarykey, autoincrement"`
	Name        string `db:"domain_name, size:200, not null"` // unique with SetUnique
	LastUpdated int    `db:"last_updated, not null"`

	Title string `xml:"Title" db:"category_title, size:200, not null"`
	Path  string `xml:"AbsolutePath" db:"category_path, size:255, not null"`
}

type Rank struct {
	Id          int    `db:"rank_id, primarykey, autoincrement"`
	Name        string `db:"domain_name, size:200, not null"` // unique with SetUnique
	LastUpdated int    `db:"last_updated, not null"`

	Code        string `xml:"Code,attr" db:"rank_code, size:10, not null"`
	CountryRank string `xml:"Rank" db:"rank_rank, size:10, not null"`
	PageViews   string `xml:"Contribution>PageViews" db:"rank_pageviews, size:20, not null"`
	Users       string `xml:"Contribution>Users" db:"rank_users, size:20, not null"`
}

// UsageStatistics
type Delta struct {
	Id          int    `db:"usage_id, primarykey, autoincrement"`
	Name        string `db:"domain_name, size:200, not null"` // unique with SetUnique
	LastUpdated int    `db:"last_updated, not null"`

	TimeRangeMonths int `xml:"TimeRange>Months" db:"months, not null"`
	TimeRangeDays   int `xml:"TimeRange>Days"   db:"days, not null"`

	// impossible to parse this as float64
	PageViewsPerMillionValue string `xml:"PageViews>PerMillion>Value" db:"pageviews_permillion_value, size:30, not null"`
	PageViewsPerMillionDelta string `xml:"PageViews>PerMillion>Delta" db:"pageviews_permillion_delta, size:30, not null"`

	PageViewsRankValue string `xml:"PageViews>Rank>Value" db:"pageviews_rank_value, size:30, not null"`
	PageViewsRankDelta string `xml:"PageViews>Rank>Delta" db:"pageviews_rank_delta, size:30, not null"`

	PageViewsPerUserValue string `xml:"PageViews>PerUser>Value" db:"pageviews_peruser_value, size:30, not null"`
	PageViewsPerUserDelta string `xml:"PageViews>PerUser>Delta" db:"pageviews_peruser_delta, size:30, not null"`
}

//
type History struct {
	Id               int     `db:"site_id, primarykey, autoincrement"`
	Name             string  `db:"domain_name, size:200, not null"`    // unique together wiht date
	Date             string  `xml:"Date" db:"date, size:14, not null"` // unique together wiht date
	PageViewsPerMio  float64 `xml:"PageViews>PerMillion" db:"pageviews_per_mio, not null"`
	PageViewsPerUser float64 `xml:"PageViews>PerUser" db:"pageviews_per_user, not null"`
	Rank             float64 `xml:"Rank" db:"global_rank, not null"` // dont ask me why the rank is sometimes not an integer - but it IS
	ReachPerMio      float64 `xml:"Reach>PerMillion" db:"country_reach_per_mio, not null"`
}

type Histories struct {
	Histories []History `xml:"Response>TrafficHistoryResult>Alexa>TrafficHistory>HistoricalData>Data"`
}
