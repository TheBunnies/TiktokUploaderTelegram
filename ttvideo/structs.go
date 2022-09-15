package ttvideo

type TTVideoDetail struct {
	Id   string
	Url  []Url
	Meta Meta
}

type Meta struct {
	Title    string
	Duration string
}

type Url struct {
	Url string
	Ext string
}
