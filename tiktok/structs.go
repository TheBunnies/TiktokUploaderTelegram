package tiktok

type AwemeDetail struct {
	AwemeList []AwemeItem `json:"aweme_list"`
}

type Image struct {
	DisplayImage DisplayImage `json:"display_image"`
}

type DisplayImage struct {
	UrlList []string `json:"url_list"`
}

type ImagePostInfo struct {
	Images []Image `json:"images"`
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
	ImagePostInfo ImagePostInfo `json:"image_post_info"`
}
