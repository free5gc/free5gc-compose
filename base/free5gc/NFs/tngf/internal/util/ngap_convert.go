package util

import (
	"encoding/binary"
	"encoding/hex"
	"strings"

	"github.com/free5gc/aper"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/tngf/internal/logger"
	"github.com/free5gc/tngf/pkg/context"
)

func PlmnIdToNgap(plmnId context.PLMNID) (ngapPlmnId ngapType.PLMNIdentity) {
	var hexString string
	mcc := strings.Split(plmnId.Mcc, "")
	mnc := strings.Split(plmnId.Mnc, "")
	if len(plmnId.Mnc) == 2 {
		hexString = mcc[1] + mcc[0] + "f" + mcc[2] + mnc[1] + mnc[0]
	} else {
		hexString = mcc[1] + mcc[0] + mnc[0] + mcc[2] + mnc[2] + mnc[1]
	}
	var err error
	ngapPlmnId.Value, err = hex.DecodeString(hexString)
	if err != nil {
		logger.UtilLog.Errorf("DecodeString error: %+v", err)
	}
	return
}

func TngfIdToNgap(tngfId uint32) (ngapTngfId *aper.BitString) {
	ngapTngfId = new(aper.BitString)
	ngapTngfId.Bytes = make([]byte, 4)
	binary.BigEndian.PutUint32(ngapTngfId.Bytes, tngfId)
	ngapTngfId.BitLength = 32
	return
}
