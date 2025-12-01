package WebUI

type ChargingData struct {
	Snssai string `json:"snssai,omitempty" yaml:"snssai" bson:"snssai" mapstructure:"snssai"`
	Dnn    string `json:"dnn" yaml:"dnn" bson:"dnn" mapstructure:"dnn"`
	QosRef int    `json:"qosRef,omitempty" yaml:"qosRef" bson:"qosRef" mapstructure:"qosRef"`
	Filter string `json:"filter" yaml:"filter" bson:"filter" mapstructure:"filter"`
	// nolint
	ChargingMethod string `json:"chargingMethod,omitempty" yaml:"chargingMethod" bson:"chargingMethod" mapstructure:"chargingMethod"`
	Quota          string `json:"quota,omitempty" yaml:"quota" bson:"quota" mapstructure:"quota"`
	UnitCost       string `json:"unitCost,omitempty" yaml:"unitCost" bson:"unitCost" mapstructure:"unitCost"`
	UeId           string `json:"ueId,omitempty" yaml:"ueId" bson:"ueId" mapstructure:"ueId"`
}
