package dto

type PanelGroupDataOutput struct {
	ServiceNum      int64 `json:"service_num"`
	RenterNum       int64 `json:"renter_num"`
	CurrentQPS      int64 `json:"current_QPS"`
	TodayRequestNum int64 `json:"today_request_num"`
}

type DashServiceStatItemOutput struct {
	Name     string `json:"name"`
	LoadType int    `json:"load_type"`
	Value    int64  `json:"value"`
}

type DashServiceStatOutput struct {
	Legend []string                    `json:"legend"`
	Data   []DashServiceStatItemOutput `json:"data"`
}
