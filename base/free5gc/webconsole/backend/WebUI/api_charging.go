package WebUI

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/chf/cdr/asn"
	"github.com/free5gc/chf/cdr/cdrFile"
	"github.com/free5gc/chf/cdr/cdrType"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/mongoapi"
	"github.com/free5gc/webconsole/backend/logger"
	"github.com/free5gc/webconsole/backend/webui_context"
)

const (
	ChargingOffline = "Offline"
	ChargingOnline  = "Online"
)

// Get vol from CDR
// TS 32.297: Charging Data Record (CDR) file format and transfer
func parseCDR(supi string) (map[int64]RatingGroupDataUsage, error) {
	logger.BillingLog.Traceln("parseCDR")
	fileName := "/tmp/webconsole/" + supi + ".cdr"
	if _, err := os.Stat(fileName); err != nil {
		return nil, err
	}

	newCdrFile := cdrFile.CDRFile{}

	newCdrFile.Decoding(fileName)
	dataUsage := make(map[int64]RatingGroupDataUsage)

	for _, cdr := range newCdrFile.CdrList {
		recvByte := cdr.CdrByte
		val := reflect.New(reflect.TypeOf(&cdrType.ChargingRecord{}).Elem()).Interface()
		err := asn.UnmarshalWithParams(recvByte, val, "")
		if err != nil {
			logger.BillingLog.Errorf("parseCDR error when unmarshal with params: %+v", err)
			continue
		}

		chargingRecord := *(val.(*cdrType.ChargingRecord))

		for _, multipleUnitUsage := range chargingRecord.ListOfMultipleUnitUsage {
			rg := multipleUnitUsage.RatingGroup.Value
			du := dataUsage[rg]

			du.Snssai = fmt.Sprintf("%02d", chargingRecord.PDUSessionChargingInformation.NetworkSliceInstanceID.SST.Value) +
				string(chargingRecord.PDUSessionChargingInformation.NetworkSliceInstanceID.SD.Value)

			du.Dnn = string(chargingRecord.PDUSessionChargingInformation.DataNetworkNameIdentifier.Value)

			for _, usedUnitContainer := range multipleUnitUsage.UsedUnitContainers {
				du.TotalVol += usedUnitContainer.DataTotalVolume.Value
				du.UlVol += usedUnitContainer.DataVolumeUplink.Value
				du.DlVol += usedUnitContainer.DataVolumeDownlink.Value
			}

			dataUsage[rg] = du
		}
	}

	return dataUsage, nil
}

func GetChargingData(c *gin.Context) {
	logger.BillingLog.Info("Get Charging Data")
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	chargingMethod, exist := c.Params.Get("chargingMethod")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"cause": "chargingMethod not provided"})
		return
	}
	logger.BillingLog.Traceln(chargingMethod)

	if chargingMethod != ChargingOffline && chargingMethod != ChargingOnline {
		c.JSON(http.StatusBadRequest, gin.H{"cause": "not support chargingMethod" + chargingMethod})
		return
	}

	filter := bson.M{
		"chargingMethod": chargingMethod,
		"ratingGroup":    nil,
	}
	chargingDataInterface, err := mongoapi.RestfulAPIGetMany(chargingDataColl, filter)
	if err != nil {
		logger.BillingLog.Errorf("mongoapi error: %+v", err)
	}

	chargingDataBsonA := toBsonA(chargingDataInterface)

	c.JSON(http.StatusOK, chargingDataBsonA)
}

func GetChargingRecord(c *gin.Context) {
	logger.BillingLog.Info("Get Charging Record")
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	webuiSelf := webui_context.GetSelf()
	webuiSelf.UpdateNfProfiles()

	// Get supi of UEs
	var uesJsonData interface{}
	if amfUris := webuiSelf.GetOamUris(models.NrfNfManagementNfType_AMF); amfUris != nil {
		requestUri := fmt.Sprintf("%s/namf-oam/v1/registered-ue-context", amfUris[0])

		ctx, pd, tokerErr := webui_context.GetSelf().GetTokenCtx(
			models.ServiceName_NAMF_OAM, models.NrfNfManagementNfType_AMF)
		if tokerErr != nil {
			logger.ProcLog.Errorf("GetTokenCtx error: %+v", tokerErr)
			c.JSON(http.StatusInternalServerError, pd)
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUri, nil)
		if err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		if err = webui_context.GetSelf().RequestBindToken(req, ctx); err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		resp, res_err := httpsClient.Do(req)
		if res_err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				logger.ProcLog.Error(closeErr)
			}
		}()

		err = json.NewDecoder(resp.Body).Decode(&uesJsonData)
		if err != nil {
			logger.BillingLog.Error(err)
		}
	}

	// build charging records
	uesBsonA := toBsonA(uesJsonData)
	chargingRecordsBsonA := make([]interface{}, 0, len(uesBsonA))

	type OfflineSliceTypeMap struct {
		supi           string
		snssai         string
		dnn            string
		unitcost       int64
		flowTotalVolum int64
		flowTotalUsage int64
	}
	// Use for sum all the flow-based charging, and add to the slice at the end.
	offlineChargingSliceTypeMap := make(map[string]OfflineSliceTypeMap)

	for _, ueData := range uesBsonA {
		ueBsonM := toBsonM(ueData)

		supi := ueBsonM["Supi"].(string)

		ratingGroupDataUsages, err := parseCDR(supi)
		if err != nil {
			logger.BillingLog.Warnln(err)
			continue
		}

		for rg, du := range ratingGroupDataUsages {
			filter := bson.M{
				"ueId":        supi,
				"ratingGroup": rg,
			}
			chargingDataInterface, err_get := mongoapi.RestfulAPIGetOne(chargingDataColl, filter)
			if err_get != nil {
				logger.ProcLog.Errorf("PostSubscriberByID err: %+v", err_get)
			}
			if len(chargingDataInterface) == 0 {
				logger.BillingLog.Warningf("ratingGroup: %d not found in mongoapi, may change the rg id", rg)
				continue
			}

			var chargingData ChargingData
			err = json.Unmarshal(mapToByte(chargingDataInterface), &chargingData)
			if err != nil {
				logger.BillingLog.Error(err)
			}
			logger.BillingLog.Debugf("add ratingGroup: %d, supi: %s, method: %s", rg, supi, chargingData.ChargingMethod)

			switch chargingData.ChargingMethod {
			case ChargingOffline:
				unitcost, err_parse := strconv.ParseInt(chargingData.UnitCost, 10, 64)
				if err_parse != nil {
					logger.BillingLog.Error("Offline unitCost strconv: ", err_parse.Error())
					unitcost = 1
				}

				key := chargingData.UeId + chargingData.Snssai
				pdu_level, exist := offlineChargingSliceTypeMap[key]
				if !exist {
					pdu_level = OfflineSliceTypeMap{}
				}
				if chargingData.Filter != "" {
					// Flow-based charging
					du.Usage = du.TotalVol * unitcost
					pdu_level.flowTotalUsage += du.Usage
					pdu_level.flowTotalVolum += du.TotalVol
				} else {
					// Slice-level charging
					pdu_level.snssai = chargingData.Snssai
					pdu_level.dnn = chargingData.Dnn
					pdu_level.supi = chargingData.UeId
					pdu_level.unitcost = unitcost
				}
				offlineChargingSliceTypeMap[key] = pdu_level
			case ChargingOnline:
				tmpInt, err1 := strconv.Atoi(chargingData.Quota)
				if err1 != nil {
					logger.BillingLog.Error("Quota strconv: ", err1, rg, du, chargingData)
				}
				du.QuotaLeft = int64(tmpInt)
			}
			du.Snssai = chargingData.Snssai
			du.Dnn = chargingData.Dnn
			du.Supi = supi
			du.Filter = chargingData.Filter

			ratingGroupDataUsages[rg] = du
			chargingRecordsBsonA = append(chargingRecordsBsonA, toBsonM(du))
		}
	}
	for idx, record := range chargingRecordsBsonA {
		tmp, err := json.Marshal(record)
		if err != nil {
			logger.BillingLog.Errorln("Marshal chargingRecordsBsonA error:", err.Error())
			continue
		}

		var rd RatingGroupDataUsage

		err = json.Unmarshal(tmp, &rd)
		if err != nil {
			logger.BillingLog.Errorln("Unmarshall RatingGroupDataUsage error:", err.Error())
			continue
		}

		if rd.Filter != "" {
			// Skip the Flow-based charging
			continue
		}

		key := rd.Supi + rd.Snssai
		if val, exist := offlineChargingSliceTypeMap[key]; exist {
			rd.Usage += val.flowTotalUsage
			rd.Usage += (rd.TotalVol - val.flowTotalVolum) * val.unitcost
			chargingRecordsBsonA[idx] = toBsonM(rd)
		}
	}

	c.JSON(http.StatusOK, chargingRecordsBsonA)
}
