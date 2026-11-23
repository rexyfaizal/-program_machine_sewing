package models

type GccSmvSyncRow struct {
	StyleName   string  `json:"styleName"`
	ProcessName string  `json:"processName"`
	CtTotal     float64 `json:"ctTotal"`
	HeaderID    int64   `json:"headerId,omitempty"`
	Buyer       string  `json:"buyer,omitempty"`
	Item        string  `json:"item,omitempty"`
}

type GccSmvSyncResponse struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	PagesFetched   int    `json:"pagesFetched"`
	HeadersFetched int    `json:"headersFetched"`
	TotalHeaders   int    `json:"totalHeaders"`
	RowsPrepared   int    `json:"rowsPrepared"`
	StylesInserted int    `json:"stylesInserted"`
	StylesUpdated  int    `json:"stylesUpdated"`
	StylesDeleted  int    `json:"stylesDeleted"`
	CtInserted     int    `json:"ctInserted"`
	SkippedEmpty   int    `json:"skippedEmpty"`
}

type GccSmvFlatRow struct {
	HeaderID   int64   `json:"headerId"`
	Tanggal    string  `json:"tanggal"`
	Buyer      string  `json:"buyer"`
	Style      string  `json:"style"`
	Item       string  `json:"item"`
	NamaProses string  `json:"namaProses"`
	ProsesID   string  `json:"prosesId"`
	CycleTime  float64 `json:"cycleTime"`
	SmvMinute  float64 `json:"smvMinute"`
	OutputPerH float64 `json:"outputPerH"`
	Mesin      string  `json:"mesin"`
}

type GccSmvListPagination struct {
	Total       int  `json:"total"`
	PerPage     int  `json:"perPage"`
	CurrentPage int  `json:"currentPage"`
	LastPage    int  `json:"lastPage"`
	From        int  `json:"from"`
	To          int  `json:"to"`
	NextPage    *int `json:"nextPage"`
	PrevPage    *int `json:"prevPage"`
}

type GccSmvListPageResponse struct {
	Status     string               `json:"status"`
	Message    string               `json:"message"`
	Source     string               `json:"source"` // gcc-live | local-cache
	Rows       []GccSmvFlatRow      `json:"rows"`
	Pagination GccSmvListPagination `json:"pagination"`
}
