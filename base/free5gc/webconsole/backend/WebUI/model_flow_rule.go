package WebUI

type FlowRule struct {
	Filter     string `json:"filter,omitempty" yaml:"filter" bson:"filter" mapstructure:"filter"`
	Precedence int    `json:"precedence,omitempty" yaml:"precedence" bson:"precedence" mapstructure:"precedence"`
	Snssai     string `json:"snssai,omitempty" yaml:"snssai" bson:"snssai" mapstructure:"snssai"`
	Dnn        string `json:"dnn,omitempty" yaml:"dnn" bson:"dnn" mapstructure:"dnn"`
	QosRef     int    `json:"qosRef,omitempty" yaml:"qosRef" bson:"qosRef" mapstructure:"qosRef"`
}
