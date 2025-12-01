package webui_context

import (
	"fmt"

	"github.com/free5gc/openapi/models"
	Nnrf_NFDiscovery "github.com/free5gc/openapi/nrf/NFDiscovery"
	"github.com/free5gc/webconsole/backend/logger"
)

type NfInstance struct {
	ValidityPeriod int                              `json:"validityPeriod"`
	NfInstances    []models.NrfNfDiscoveryNfProfile `json:"nfInstances"`
}

func SendSearchNFInstances(targetNfType models.NrfNfManagementNfType) ([]models.NrfNfDiscoveryNfProfile, error) {
	var nfProfiles []models.NrfNfDiscoveryNfProfile

	ctx, _, err := GetSelf().GetTokenCtx(models.ServiceName_NNRF_DISC, models.NrfNfManagementNfType_NRF)
	if err != nil {
		logger.ConsumerLog.Errorln(err.Error())
		return nfProfiles, err
	}

	client := GetSelf().NFDiscoveryClient
	requestNfType := models.NrfNfManagementNfType_AF

	req := &Nnrf_NFDiscovery.SearchNFInstancesRequest{
		TargetNfType:    &targetNfType,
		RequesterNfType: &requestNfType,
	}

	res, err := client.NFInstancesStoreApi.SearchNFInstances(ctx, req)
	if err != nil {
		logger.ConsumerLog.Errorf("SearchNFInstances failed: %+v", err)
		return nfProfiles, err
	}
	if res == nil {
		return nfProfiles, fmt.Errorf("SearchNFInstances resule nil:%+v", err)
	}
	nfProfiles = res.SearchResult.NfInstances

	return nfProfiles, nil
}
