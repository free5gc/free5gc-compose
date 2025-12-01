package WebUI

// frontend's FlowChargingRecord
type RatingGroupDataUsage struct {
	Supi      string `bson:"Supi"`
	Filter    string `bson:"Filter"`
	Snssai    string `bson:"Snssai"`
	Dnn       string `bson:"Dnn"`
	TotalVol  int64  `bson:"TotalVol"`
	UlVol     int64  `bson:"UlVol"`
	DlVol     int64  `bson:"DlVol"`
	QuotaLeft int64  `bson:"QuotaLeft"`
	Usage     int64  `bson:"Usage"`
}
