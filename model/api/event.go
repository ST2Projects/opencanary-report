package api

type Event struct {
	DstHost           string  `json:"dst_host"`
	DstPort           int     `json:"dst_port"`
	LocalTime         string  `json:"local_time"`
	LocalTimeAdjusted string  `json:"local_time_adjusted"`
	Logdata           Logdata `json:"logdata"`
	Logtype           int     `json:"logtype"`
	NodeID            string  `json:"node_id"`
	SrcHost           string  `json:"src_host"`
	SrcPort           int     `json:"src_port"`
	UtcTime           string  `json:"utc_time"`
}
type Logdata struct {
	Password string `json:"PASSWORD"`
	Username string `json:"USERNAME"`
}
