package context

import (
	"errors"

	"github.com/asaskevich/govalidator"
)

type TNGFNFInfo struct {
	GlobalTNGFID    GlobalTNGFID      `yaml:"GlobalTNGFID" valid:"required"`
	RanNodeName     string            `yaml:"Name,omitempty" valid:"optional"`
	SupportedTAList []SupportedTAItem `yaml:"SupportedTAList" valid:"required"`
}

type GlobalTNGFID struct {
	PLMNID PLMNID `yaml:"PLMNID" valid:"required"`
	TNGFID uint32 `yaml:"TNGFID" valid:"range(0|65535),required"` // with length 2 bytes
}

type SupportedTAItem struct {
	TAC               string              `yaml:"TAC" valid:"hexadecimal,stringlength(6|6),required"`
	BroadcastPLMNList []BroadcastPLMNItem `yaml:"BroadcastPLMNList" valid:"required"`
}

type BroadcastPLMNItem struct {
	PLMNID              PLMNID             `yaml:"PLMNID" valid:"required"`
	TAISliceSupportList []SliceSupportItem `yaml:"TAISliceSupportList" valid:"required"`
}

type PLMNID struct {
	Mcc string `yaml:"MCC" valid:"numeric,stringlength(3|3),required"`
	Mnc string `yaml:"MNC" valid:"numeric,stringlength(2|3),required"`
}

type SliceSupportItem struct {
	SNSSAI SNSSAIItem `yaml:"SNSSAI" valid:"required"`
}

type SNSSAIItem struct {
	SST string `yaml:"SST" valid:"hexadecimal,stringlength(1|1),required"`
	SD  string `yaml:"SD,omitempty" valid:"hexadecimal,stringlength(6|6),required"`
}

type AMFSCTPAddresses struct {
	IPAddresses []string `yaml:"IP" valid:"required"`
	Port        int      `yaml:"Port,omitempty" valid:"port,optional"` // Default port is 38412 if not defined.
}

func (a *AMFSCTPAddresses) Validate() (bool, error) {
	var errs govalidator.Errors

	for _, IPAddress := range a.IPAddresses {
		if !govalidator.IsHost(IPAddress) {
			err := errors.New("invalid AMFSCTPAddresses.IP: " + IPAddress + ", does not validate as IP")
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return false, errs
	}

	return true, nil
}
