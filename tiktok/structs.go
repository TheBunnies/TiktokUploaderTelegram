package tiktok

type AwemeDetail struct {
	AwemeList []AwemeItem `json:"aweme_list"`
}

type AwemeItem struct {
	AwemeId    string `json:"aweme_id"`
	CreateTime int64  `json:"create_time"`
	Desc       string `json:"desc"`
	Author     struct {
		Nickname string `json:"nickname"`
	}
	Video struct {
		Duration  int64
		Play_Addr struct {
			Width    int
			Height   int
			URL_List []string
		}
	}
}
