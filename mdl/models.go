package mdl

// The prefix creep is necessary, to make the structs anonymously embeddable
type Domain struct {
	Id          int    `db:"domain_id, primarykey, autoincrement"`
	Name        string `db:"domain_name, size:200, not null"`
	Label       string `db:"domain_label, size:200, not null"`
	Description string `db:"domain_description, size:200, not null"`
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
