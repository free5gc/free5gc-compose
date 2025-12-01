package WebUI

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/openapi/models"
	smf_factory "github.com/free5gc/smf/pkg/factory"
	"github.com/free5gc/util/mongoapi"
	"github.com/free5gc/webconsole/backend/logger"
	"github.com/free5gc/webconsole/backend/webui_context"
)

type VerifyScope struct {
	Supi   string `json:"supi"`
	Sd     string `json:"sd,omitempty"`
	Sst    int    `json:"sst"`
	Dnn    string `json:"dnn"`
	Ipaddr string `json:"ipaddr"`
}

func GetSmfUserPlaneInfo() (interface{}, error) {
	logger.ProcLog.Infoln("Get SMF UserPlane Info")

	webuiSelf := webui_context.GetSelf()
	webuiSelf.UpdateNfProfiles()

	var jsonData interface{}

	// TODO: support fetching data from multiple SMF
	if smfUris := webuiSelf.GetOamUris(models.NrfNfManagementNfType_SMF); smfUris != nil {
		requestUri := fmt.Sprintf("%s/nsmf-oam/v1/user-plane-info/", smfUris[0])

		ctx, pd, err := webuiSelf.GetTokenCtx(models.ServiceName_NSMF_OAM, models.NrfNfManagementNfType_SMF)
		if err != nil {
			logger.ConsumerLog.Infof("GetTokenCtx: service %v, err: %+v", models.ServiceName_NSMF_OAM, err)
			return pd, err
		}

		req, err_req := http.NewRequestWithContext(ctx, http.MethodGet, requestUri, nil)
		if err_req != nil {
			logger.ProcLog.Error(err_req)
			return jsonData, err_req
		}

		if err = webui_context.GetSelf().RequestBindToken(req, ctx); err != nil {
			logger.ProcLog.Error(err)
			return jsonData, err
		}

		resp, err_rsp := httpsClient.Do(req)
		if err_rsp != nil {
			logger.ProcLog.Error(err_rsp)
			return jsonData, err_rsp
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				logger.ProcLog.Error(closeErr)
			}
		}()

		json_err := json.NewDecoder(resp.Body).Decode(&jsonData)
		if json_err != nil {
			logger.ProcLog.Errorf("Decode Json err: %+v", err)
		}
		return jsonData, err
	} else {
		logger.ProcLog.Error("No SMF Found")
	}
	return jsonData, nil
}

func GetStaticIpPoolsFromUserPlaneInfomation(
	userplaneinfo *smf_factory.UserPlaneInformation,
	snssai models.Snssai,
	dnn string,
) ([]netip.Prefix, error) {
	result := []netip.Prefix{}

	for nodeName := range userplaneinfo.UPNodes {
		if userplaneinfo.UPNodes[nodeName].Type == "UPF" {
			// Find the UPF node
			for _, snssaiupfinfo := range userplaneinfo.UPNodes[nodeName].SNssaiInfos {
				// Find the Slice (snssai)
				if *snssaiupfinfo.SNssai == snssai {
					for _, dnnInfo := range snssaiupfinfo.DnnUpfInfoList {
						// Find the DNN name
						if dnnInfo.Dnn == dnn {
							if len(dnnInfo.StaticPools) > 0 {
								for _, pool := range dnnInfo.StaticPools {
									staticPoolstr := pool.Cidr
									net, parseErr := netip.ParsePrefix(staticPoolstr)
									if parseErr != nil {
										return result, parseErr
									}
									result = append(result, net)
								}
								return result, nil
							}
							// If there is no static pool, return smallest
							net, parseErr := netip.ParsePrefix("0.0.0.0/32")
							return []netip.Prefix{net}, parseErr
						}
					}
				}
			}
		}
	}
	net, parseErr := netip.ParsePrefix("0.0.0.0/32")
	return []netip.Prefix{net}, parseErr
}

func getDnnStaticIpPools(snssai models.Snssai, dnn string) ([]netip.Prefix, error) {
	var userplaneinfo smf_factory.UserPlaneInformation

	raw_info, get_err := GetSmfUserPlaneInfo()
	if get_err != nil {
		logger.ProcLog.Errorf("GetSmfUserPlaneInfo(): %+v", get_err)
		return []netip.Prefix{}, get_err
	}

	tmp, err := json.Marshal(raw_info)
	if err != nil {
		logger.ProcLog.Errorf("Marshal err: %+v", err)
	}
	unmarshal_err := json.Unmarshal(tmp, &userplaneinfo)
	if unmarshal_err != nil {
		logger.ProcLog.Errorf("Unmarshal err: %+v", unmarshal_err)
		return []netip.Prefix{}, unmarshal_err
	}

	return GetStaticIpPoolsFromUserPlaneInfomation(&userplaneinfo, snssai, dnn)
}

func VerifyStaticIP(c *gin.Context) {
	logger.ProcLog.Info("Verify StaticIP")
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	var checkData VerifyScope
	if err := c.ShouldBindJSON(&checkData); err != nil {
		logger.ProcLog.Errorln(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"cause": err.Error(),
		})
		return
	}

	snssai := models.Snssai{
		Sst: int32(checkData.Sst),
	}
	if checkData.Sd != "" {
		snssai.Sd = checkData.Sd
	}

	staticPools, get_pool_err := getDnnStaticIpPools(snssai, checkData.Dnn)
	if get_pool_err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  get_pool_err,
			"ipaddr": checkData.Ipaddr,
			"valid":  false,
			"cause":  get_pool_err.Error(),
		})
		return
	}
	VerifyStaticIpProcedure(c, checkData, staticPools)
}

func VerifyStaticIpProcedure(
	c *gin.Context,
	checkData VerifyScope,
	staticPools []netip.Prefix,
) {
	staticIp, parse_err := netip.ParseAddr(checkData.Ipaddr)
	if parse_err != nil {
		logger.ProcLog.Errorln(parse_err.Error())
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
			"cause": parse_err.Error(),
		})
		return
	}
	logger.ProcLog.Debugln("check IP address:", staticIp)

	// Check in Static Pool
	result := false
	for _, staticPool := range staticPools {
		result = staticPool.Contains(staticIp)
		if result {
			break
		}
	}
	if !result {
		c.JSON(http.StatusOK, gin.H{
			"ipaddr": staticIp,
			"valid":  result,
			"cause":  "Not in static pools!",
		})
		logger.ProcLog.Debugln("StaticIP", staticIp, ": not in static pool!")
		return
	}

	if gin.Mode() != "test" && checkIpCollisionFromDb(c, checkData) != nil {
		return
	}

	// Return the result
	c.JSON(http.StatusOK, gin.H{
		"ipaddr": staticIp,
		"valid":  result,
		"cause":  "",
	})
}

// Check IP not used by other UE
func checkIpCollisionFromDb(
	c *gin.Context,
	checkData VerifyScope,
) error {
	snssai := models.Snssai{
		Sst: int32(checkData.Sst),
	}
	if checkData.Sd != "" {
		snssai.Sd = checkData.Sd
	}

	smDataColl := "subscriptionData.provisionedData.smData"
	filter := bson.M{
		"singleNssai": snssai,
		"ueId":        bson.D{{Key: "$ne", Value: checkData.Supi}}, // not this UE
	}
	smDataDataInterface, mongo_err := mongoapi.RestfulAPIGetMany(smDataColl, filter)
	if mongo_err != nil {
		logger.ProcLog.Warningln(smDataColl, "mongo error: ", mongo_err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"ipaddr": checkData.Ipaddr,
			"valid":  false,
			"cause":  mongo_err.Error(),
		})
		return mongo_err
	}
	var smDatas []models.SessionManagementSubscriptionData
	if err := json.Unmarshal(sliceToByte(smDataDataInterface), &smDatas); err != nil {
		logger.ProcLog.Errorf("Unmarshal smDatas err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return err
	}
	for _, smData := range smDatas {
		if dnnConfig, ok := smData.DnnConfigurations[checkData.Dnn]; ok {
			for _, ipData := range dnnConfig.StaticIpAddress {
				if checkData.Ipaddr == ipData.Ipv4Addr {
					msg := "StaticIP: " + checkData.Ipaddr + " has already exist!"
					logger.ProcLog.Warningln(msg)
					c.JSON(http.StatusOK, gin.H{
						"ipaddr": checkData.Ipaddr,
						"valid":  false,
						"cause":  msg,
					})
					return fmt.Errorf("%s", msg)
				}
			}
		}
	}
	return nil
}
