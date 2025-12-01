package processor

import (
	"fmt"
	"net/http"

	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/nef/pkg/factory"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/metrics/sbi"
	"github.com/gin-gonic/gin"
)

// GetApplicationsPFD Returns a representation of PFDs for multiple applications
// 3GPP TS 29.551 release 17 version 17.6.0
// Resource structure: 5.3.1
// Request/Response  : 5.3.2.3.1
func (p *Processor) GetApplicationsPFD(c *gin.Context, appIDs []string) {
	logger.PFDFLog.Infof("GetApplicationsPFD - appIDs: %v", appIDs)

	// TODO: Support SupportedFeatures
	pdfDataForAppExt, pd, errAppDataGet := p.Consumer().AppDataPfdsGet(appIDs)

	switch {
	case pd != nil:
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	case errAppDataGet != nil:
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Detail: "Query to UDR failed",
		}
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, problemDetails.Cause)
		c.JSON(int(problemDetails.Status), problemDetails)
		return
	}

	var pfdDataForApp []models.PfdDataForApp

	for _, dataForExt := range pdfDataForAppExt {
		pfdDataForApp = append(pfdDataForApp, *convertPdfDataForAppExtToPfdDataForApp(&dataForExt))
	}

	c.JSON(http.StatusOK, pfdDataForApp)
}

// GetIndividualApplicationPFD Returns a representation of PFDs for an application
// 3GPP TS 29.551 release 17 version 17.6.0
// Resource structure: 5.3.1
// Request/Response  : 5.3.3.3.1
func (p *Processor) GetIndividualApplicationPFD(c *gin.Context, appID string) {
	logger.PFDFLog.Infof("GetIndividualApplicationPFD - appID[%s]", appID)

	// TODO: Support SupportedFeatures
	pdfDataRsp, pdfDataProblemDetails, errPdfData := p.Consumer().AppDataPfdsAppIdGet(appID)

	switch {
	case pdfDataProblemDetails != nil:
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pdfDataProblemDetails.Cause)
		c.JSON(int(pdfDataProblemDetails.Status), pdfDataProblemDetails)
		return
	case errPdfData != nil:
		problemDetails := models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Detail: "Query to UDR failed",
		}
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, problemDetails.Cause)
		c.JSON(int(problemDetails.Status), problemDetails)
		return
	}

	pdfDataForApp := convertPdfDataForAppExtToPfdDataForApp(&pdfDataRsp.PfdDataForAppExt)

	c.JSON(http.StatusOK, pdfDataForApp)
}

// PostPFDSubscriptions Subscribe the notification of PFD changes.
// 3GPP TS 29.551 release 17 version 17.6.0
// Resource structure: 5.3.1
// Request/Response  : 5.3.4.3.1
func (p *Processor) PostPFDSubscriptions(c *gin.Context, pfdSubsc *models.PfdSubscription) {
	logger.PFDFLog.Infof("PostPFDSubscriptions - appIDs: %v", pfdSubsc.ApplicationIds)

	// TODO: Support SupportedFeatures
	if len(pfdSubsc.NotifyUri) == 0 {
		pd := openapi.ProblemDetailsDataNotFound("Absent of Notify URI")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	subID := p.Notifier().PfdChangeNotifier.AddPfdSub(pfdSubsc)
	hdrs := make(map[string][]string)
	addLocationheader(hdrs, p.genPfdSubscriptionURI(subID))

	for k, values := range hdrs {
		for _, value := range values {
			c.Header(k, value)
		}
	}
	c.JSON(http.StatusCreated, pfdSubsc)
}

// DeleteIndividualPFDSubscription Delete a subscription to PFD change notifications.
// 3GPP TS 29.551 release 17 version 17.6.0
// Resource structure: 5.3.1
// Request/Response  : 5.3.5.3.1
func (p *Processor) DeleteIndividualPFDSubscription(c *gin.Context, subID string) {
	logger.PFDFLog.Infof("DeleteIndividualPFDSubscription - subID[%s]", subID)

	if err := p.Notifier().PfdChangeNotifier.DeletePfdSub(subID); err != nil {
		pd := openapi.ProblemDetailsDataNotFound(err.Error())
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	c.Status(http.StatusNoContent)
}

func (p *Processor) genPfdSubscriptionURI(subID string) string {
	// E.g. "https://localhost:29505/nnef-pfdmanagement/v1/subscriptions/{subscriptionId}
	return fmt.Sprintf("%s/subscriptions/%s", p.Config().ServiceUri(factory.ServiceNefPfd), subID)
}
