package tiktok

type AwemeDetail struct {
	Author struct {
		Unique_ID string
	}
	Aweme_ID    string
	Create_Time int64
	Desc        string
	Video       struct {
		Duration  int64
		Play_Addr struct {
			Width    int
			Height   int
			URL_List []string
		}
	}
}
