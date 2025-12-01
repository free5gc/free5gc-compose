package WebUI

type QosFlow struct {
	Snssai string `json:"snssai" yaml:"snssai" bson:"snssai" mapstructure:"snssai"`
	Dnn    string `json:"dnn" yaml:"dnn" bson:"dnn" mapstructure:"dnn"`
	QosRef uint8  `json:"qosRef" yaml:"qosRef" bson:"qosRef" mapstructure:"qosRef"`
	Var5QI int    `json:"5qi" yaml:"5qi" bson:"5qi" mapstructure:"5qi"`
	MBRUL  string `json:"mbrUL,omitempty" yaml:"mbrUL" bson:"mbrUL" mapstructure:"mbrUL"`
	MBRDL  string `json:"mbrDL,omitempty" yaml:"mbrDL" bson:"mbrDL" mapstructure:"mbrDL"`
	GBRUL  string `json:"gbrUL,omitempty" yaml:"gbrUL" bson:"gbrUL" mapstructure:"gbrUL"`
	GBRDL  string `json:"gbrDL,omitempty" yaml:"gbrDL" bson:"gbrDL" mapstructure:"gbrDL"`
}
