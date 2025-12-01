package WebUI

type ChargingRecord struct {
	Snssai string `json:"snssai,omitempty" yaml:"snssai" bson:"snssai" mapstructure:"snssai"`
	Dnn    string `json:"dnn" yaml:"dnn" bson:"dnn" mapstructure:"dnn"`
	Filter string `json:"filter" yaml:"filter" bson:"filter" mapstructure:"filter"`
	QosRef int    `json:"qosRef,omitempty" yaml:"qosRef" bson:"qosRef" mapstructure:"qosRef"`
	// nolint
	ChargingMethod string `json:"chargingMethod,omitempty" yaml:"chargingMethod" bson:"chargingMethod" mapstructure:"chargingMethod"`
	Quota          string `json:"quota,omitempty" yaml:"quota" bson:"quota" mapstructure:"quota"`
	UnitCost       string `json:"unitCost,omitempty" yaml:"unitCost" bson:"unitCost" mapstructure:"unitCost"`
	RatingGroup    int64  `json:"ratingGroup,omitempty" yaml:"ratingGroup" bson:"ratingGroup" mapstructure:"ratingGroup"`
}
