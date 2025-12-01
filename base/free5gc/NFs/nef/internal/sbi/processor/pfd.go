package processor

import (
	"fmt"
	"net/http"

	nef_context "github.com/free5gc/nef/internal/context"
	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/nef/internal/util"
	"github.com/free5gc/nef/pkg/factory"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/metrics/sbi"
	"github.com/gin-gonic/gin"
)

const (
	DetailNoAF       = "Given AF is not existed"
	DetailNoPfdData  = "Absent of PfdManagement.PfdDatas"
	DetailNoPfd      = "Absent of PfdData.Pfds"
	DetailNoExtAppID = "Absent of PfdData.ExternalAppID"
	DetailNoPfdID    = "Absent of Pfd.PfdID"
	DetailNoPfdInfo  = "One of FlowDescriptions, Urls or DomainNames should be provided"
)

// GetPFDManagementTransactions Read all or queried PFDs for a given SCS/AS
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  :5.11.3.1-1
func (p *Processor) GetPFDManagementTransactions(c *gin.Context, scsAsID string) {
	logger.PFDManageLog.Infof("GetPFDManagementTransactions - scsAsID[%s]", scsAsID)

	nefCtx := p.Context()
	af := nefCtx.GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound(DetailNoAF)
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(http.StatusNotFound, pd)
		return
	}

	af.Mu.RLock()
	defer af.Mu.RUnlock()

	var pfdMngs []models.PfdManagement
	for _, afPfdTr := range af.PfdTrans {
		pfdMng, pd := p.buildPfdManagement(scsAsID, afPfdTr)
		if pd != nil {
			c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
			c.JSON(int(pd.Status), pd)
			return
		}
		pfdMngs = append(pfdMngs, *pfdMng)
	}

	c.JSON(http.StatusOK, &pfdMngs)
}

// PostPFDManagementTransactions Create PFDs for a given SCS/AS and one or more external Application Identifier(s).
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  : 5.11.3.2.3.3
func (p *Processor) PostPFDManagementTransactions(
	c *gin.Context,
	scsAsID string,
	pfdMng *models.PfdManagement,
) {
	logger.PFDManageLog.Infof("PostPFDManagementTransactions - scsAsID[%s]", scsAsID)

	// TODO: Authorize the AF

	nefCtx := p.Context()
	if pd := validatePfdManagement(scsAsID, "-1", pfdMng, nefCtx); pd != nil {
		if pd.Status == http.StatusInternalServerError {
			c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
			c.JSON(http.StatusInternalServerError, &pfdMng.PfdReports)
			return
		} else {
			c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
			c.JSON(int(pd.Status), pd)
			return
		}
	}

	af := nefCtx.GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound(DetailNoAF)
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(http.StatusNotFound, pd)
		return
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	afPfdTr := af.NewPfdTrans()
	if afPfdTr == nil {
		pd := openapi.ProblemDetailsSystemFailure("No resource can be allocated")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pfdNotifyContext := p.Notifier().PfdChangeNotifier.NewPfdNotifyContext()
	defer pfdNotifyContext.FlushNotifications()

	for appID, pfdData := range pfdMng.PfdDatas {
		afPfdTr.AddExtAppID(appID)
		pfdDataForApp := convertPfdDataToPfdDataForApp(&pfdData)
		if pfdReport := p.storePfdDataToUDR(appID, pfdDataForApp); pfdReport != nil {
			delete(pfdMng.PfdDatas, appID)
			addPfdReport(pfdMng, pfdReport)
		} else {
			pfdData.Self = p.genPfdDataURI(scsAsID, afPfdTr.TransID, appID)
			pfdMng.PfdDatas[appID] = pfdData
			pfdNotifyContext.AddNotification(appID, &models.PfdChangeNotification{
				ApplicationId: appID,
				Pfds:          pfdDataForApp.Pfds,
			})
		}
	}
	if len(pfdMng.PfdDatas) == 0 {
		// The PFDs for all applications were not created successfully.
		// PfdReport is included with detailed information.
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, util.METRICS_APP_PFDS_CREATION_ERR_MSG)
		c.JSON(http.StatusInternalServerError, &pfdMng.PfdReports)
		return
	}

	af.PfdTrans[afPfdTr.TransID] = afPfdTr
	afPfdTr.Log.Infoln("PFD Management Transaction is added")

	nefCtx.AddAf(af)

	pfdMng.Self = p.genPfdManagementURI(scsAsID, afPfdTr.TransID)

	c.JSON(http.StatusCreated, pfdMng)
}

// DeletePFDManagementTransactions Remove all PFDs for a given SCS/AS
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource Structure: Missing from 5.11.3
// Request/Response  : 5.11.3.2.3.5
func (p *Processor) DeletePFDManagementTransactions(c *gin.Context, scsAsID string) {
	logger.PFDManageLog.Infof("DeletePFDManagementTransactions - scsAsID[%s]", scsAsID)

	nefCtx := p.Context()
	af := nefCtx.GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound(DetailNoAF)
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(http.StatusNotFound, pd)
		return
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	pfdNotifyContext := p.Notifier().PfdChangeNotifier.NewPfdNotifyContext()
	defer pfdNotifyContext.FlushNotifications()

	for _, afPfdTr := range af.PfdTrans {
		for extAppID := range afPfdTr.ExtAppIDs {
			pd := p.deletePfdDataFromUDR(extAppID)
			if pd != nil {
				c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
				c.JSON(int(pd.Status), pd)
				return
			}
			pfdNotifyContext.AddNotification(extAppID, &models.PfdChangeNotification{
				ApplicationId: extAppID,
				RemovalFlag:   true,
			})
		}
		delete(af.PfdTrans, afPfdTr.TransID)
		afPfdTr.Log.Infoln("PFD Management Transaction is deleted")
	}

	// TODO: Remove AfCtx if its subscriptions and transactions are both empty

	c.Status(http.StatusNoContent)
}

// GetIndividualPFDManagementTransaction Read all PFDs for a given SCS/AS and a transaction for one or more external
// Application Identifier(s)
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  : 5.11.3.3.3.1
func (p *Processor) GetIndividualPFDManagementTransaction(
	c *gin.Context, scsAsID, transID string,
) {
	logger.PFDManageLog.Infof("GetIndividualPFDManagementTransaction - scsAsID[%s], transID[%s]",
		scsAsID, transID)

	af := p.Context().GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	af.Mu.RLock()
	defer af.Mu.RUnlock()

	afPfdTr, ok := af.PfdTrans[transID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("PFD transaction not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pfdMng, pd := p.buildPfdManagement(scsAsID, afPfdTr)
	if pd != nil {
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	c.JSON(http.StatusOK, pfdMng)
}

// PutIndividualPFDManagementTransaction Update PFD(s) for a given SCS/AS and a transaction for one or more external
// Application Identifier(s)
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  : 5.11.3.3.3.2
func (p *Processor) PutIndividualPFDManagementTransaction(
	c *gin.Context,
	scsAsID, transID string,
	pfdMng *models.PfdManagement,
) {
	logger.PFDManageLog.Infof("PutIndividualPFDManagementTransaction - scsAsID[%s], transID[%s]",
		scsAsID, transID)

	// TODO: Authorize the AF

	nefCtx := p.Context()
	if pd := validatePfdManagement(scsAsID, transID, pfdMng, nefCtx); pd != nil {
		if pd.Status == http.StatusInternalServerError {
			c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
			c.JSON(http.StatusInternalServerError, &pfdMng.PfdReports)
			return
		} else {
			c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
			c.JSON(int(pd.Status), pd)
			return
		}
	}

	af := nefCtx.GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	afPfdTr, ok := af.PfdTrans[transID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("PFD transaction not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pfdNotifyContext := p.Notifier().PfdChangeNotifier.NewPfdNotifyContext()
	defer pfdNotifyContext.FlushNotifications()

	// Delete PfdDataForApps in UDR with appID absent in new PfdManagement
	var deprecatedAppIDs []string
	for extAppID := range afPfdTr.ExtAppIDs {
		if _, exist := pfdMng.PfdDatas[extAppID]; !exist {
			deprecatedAppIDs = append(deprecatedAppIDs, extAppID)
		}
	}
	for _, appID := range deprecatedAppIDs {
		if pd := p.deletePfdDataFromUDR(appID); pd != nil {
			c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
			c.JSON(int(pd.Status), pd)
			return
		}
		pfdNotifyContext.AddNotification(appID, &models.PfdChangeNotification{
			ApplicationId: appID,
			RemovalFlag:   true,
		})
	}

	afPfdTr.DeleteAllExtAppIDs()
	for appID, pfdData := range pfdMng.PfdDatas {
		afPfdTr.AddExtAppID(appID)
		pfdDataForAppExt := convertPfdDataToPfdDataForAppExt(&pfdData)
		if pfdReport := p.storePfdDataToUDR(appID, pfdDataForAppExt); pfdReport != nil {
			delete(pfdMng.PfdDatas, appID)
			addPfdReport(pfdMng, pfdReport)
		} else {
			pfdData.Self = p.genPfdDataURI(scsAsID, afPfdTr.TransID, appID)
			pfdMng.PfdDatas[appID] = pfdData
			pfdNotifyContext.AddNotification(appID, &models.PfdChangeNotification{
				ApplicationId: appID,
				Pfds:          pfdDataForAppExt.Pfds,
			})
		}
	}
	if len(pfdMng.PfdDatas) == 0 {
		// The PFDs for all applications were not created successfully.
		// PfdReport is included with detailed information.
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, util.METRICS_APP_PFDS_CREATION_ERR_MSG)

		c.JSON(http.StatusInternalServerError, &pfdMng.PfdReports)
		return
	}

	pfdMng.Self = p.genPfdManagementURI(scsAsID, afPfdTr.TransID)

	c.JSON(http.StatusOK, pfdMng)
}

// DeleteIndividualPFDManagementTransaction Delete PFDs for a given SCS/AS and a transaction for one or more external
// Application Identifier(s)
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  : 5.11.3.3.3.5
func (p *Processor) DeleteIndividualPFDManagementTransaction(
	c *gin.Context, scsAsID, transID string,
) {
	logger.PFDManageLog.Infof("DeleteIndividualPFDManagementTransaction - scsAsID[%s], transID[%s]", scsAsID, transID)

	nefCtx := p.Context()
	af := nefCtx.GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	afPfdTr, ok := af.PfdTrans[transID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("PFD transaction not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pfdNotifyContext := p.Notifier().PfdChangeNotifier.NewPfdNotifyContext()
	defer pfdNotifyContext.FlushNotifications()

	for extAppID := range afPfdTr.ExtAppIDs {
		if pd := p.deletePfdDataFromUDR(extAppID); pd != nil {
			c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
			c.JSON(int(pd.Status), pd)
			return
		}
		pfdNotifyContext.AddNotification(extAppID, &models.PfdChangeNotification{
			ApplicationId: extAppID,
			RemovalFlag:   true,
		})
	}
	delete(af.PfdTrans, afPfdTr.TransID)
	afPfdTr.Log.Infoln("PFD Management Transaction is deleted")

	// TODO: Remove AfCtx if its subscriptions and transactions are both empty

	c.Status(http.StatusNoContent)
}

// GetIndividualApplicationPFDManagement Read PFDs at individual application level
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  : 5.11.3.4.3.1
func (p *Processor) GetIndividualApplicationPFDManagement(
	c *gin.Context, scsAsID, transID, appID string,
) {
	logger.PFDManageLog.Infof("GetIndividualApplicationPFDManagement - scsAsID[%s], transID[%s], appID[%s]",
		scsAsID, transID, appID)

	af := p.Context().GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	af.Mu.RLock()
	defer af.Mu.RUnlock()

	afPfdTr, ok := af.PfdTrans[transID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("PFD transaction not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	_, ok = afPfdTr.ExtAppIDs[appID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("Application ID not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pdfData, pd, errPfdData := p.Consumer().AppDataPfdsAppIdGet(appID)

	switch {
	case pd != nil:
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	case errPfdData != nil:
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Detail: "Query to UDR failed",
		}
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, problemDetails.Cause)
		c.JSON(int(problemDetails.Status), problemDetails)
		return
	}

	pfdData := convertPdfDataForAppExtToPfdData(&pdfData.PfdDataForAppExt)
	pfdData.Self = p.genPfdDataURI(scsAsID, transID, appID)

	c.JSON(http.StatusOK, pfdData)
}

// DeleteIndividualApplicationPFDManagement Delete PFDs at individual application level.
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  : 5.11.3.4.3.5
func (p *Processor) DeleteIndividualApplicationPFDManagement(
	c *gin.Context, scsAsID, transID, appID string,
) {
	logger.PFDManageLog.Infof("DeleteIndividualApplicationPFDManagement - scsAsID[%s], transID[%s], appID[%s]",
		scsAsID, transID, appID)

	af := p.Context().GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	afPfdTr, ok := af.PfdTrans[transID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("PFD transaction not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	_, ok = afPfdTr.ExtAppIDs[appID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("Application ID not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pfdNotifyContext := p.Notifier().PfdChangeNotifier.NewPfdNotifyContext()
	defer pfdNotifyContext.FlushNotifications()

	if pd := p.deletePfdDataFromUDR(appID); pd != nil {
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}
	afPfdTr.DeleteExtAppID(appID)
	pfdNotifyContext.AddNotification(appID, &models.PfdChangeNotification{
		ApplicationId: appID,
		RemovalFlag:   true,
	})

	// TODO: Remove afPfdTr if its appID is empty

	// TODO: Remove AfCtx if its subscriptions and transactions are both empty

	c.Status(http.StatusNoContent)
}

// PutIndividualApplicationPFDManagement Update PFDs at individual application level.
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  : 5.11.3.4.3.2
func (p *Processor) PutIndividualApplicationPFDManagement(
	c *gin.Context,
	scsAsID, transID, appID string,
	pfdData *models.PfdData,
) {
	logger.PFDManageLog.Infof("PutIndividualApplicationPFDManagement - scsAsID[%s], transID[%s], appID[%s]",
		scsAsID, transID, appID)

	// TODO: Authorize the AF

	nefCtx := p.Context()
	af := nefCtx.GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	afPfdTr, ok := af.PfdTrans[transID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("PFD transaction not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	_, ok = afPfdTr.ExtAppIDs[appID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("Application ID not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	if pd := validatePfdData(pfdData, false); pd != nil {
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pfdNotifyContext := p.Notifier().PfdChangeNotifier.NewPfdNotifyContext()
	defer pfdNotifyContext.FlushNotifications()

	pfdDataForApp := convertPfdDataToPfdDataForApp(pfdData)
	if pfdReport := p.storePfdDataToUDR(appID, pfdDataForApp); pfdReport != nil {
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pfdReport.FailureCode)
		c.JSON(http.StatusInternalServerError, pfdReport)
		return
	}
	pfdData.Self = p.genPfdDataURI(scsAsID, transID, appID)
	pfdNotifyContext.AddNotification(appID, &models.PfdChangeNotification{
		ApplicationId: appID,
		Pfds:          pfdDataForApp.Pfds,
	})

	c.JSON(http.StatusOK, pfdData)
}

// PatchIndividualApplicationPFDManagement Update PFDs at individual application level.
// 3GPP TS 29.122 release 17 version 17.6.0
// Resource structure: 5.11.3
// Request/Response  : 5.11.3.4.3.3
func (p *Processor) PatchIndividualApplicationPFDManagement(
	c *gin.Context,
	scsAsID, transID, appID string,
	pfdData *models.PfdData,
) {
	logger.PFDManageLog.Infof("PatchIndividualApplicationPFDManagement - scsAsID[%s], transID[%s], appID[%s]",
		scsAsID, transID, appID)

	// TODO: Authorize the AF

	nefCtx := p.Context()
	af := nefCtx.GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	afPfdTr, ok := af.PfdTrans[transID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("PFD transaction not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	_, ok = afPfdTr.ExtAppIDs[appID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("Application ID not found")
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	if pd := validatePfdData(pfdData, true); pd != nil {
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pfdNotifyContext := p.Notifier().PfdChangeNotifier.NewPfdNotifyContext()
	defer pfdNotifyContext.FlushNotifications()

	pdfData, problemDetails, errPfdData := p.Consumer().AppDataPfdsAppIdGet(appID)

	switch {
	case problemDetails != nil:
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, problemDetails.Cause)
		c.JSON(int(problemDetails.Status), problemDetails)
		return
	case errPfdData != nil:
		problemDetailsErr := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Detail: "Query to UDR failed",
		}
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, problemDetails.Cause)
		c.JSON(int(problemDetailsErr.Status), problemDetailsErr)
		return
	}

	oldPfdData := convertPdfDataForAppExtToPfdData(&pdfData.PfdDataForAppExt)
	if pd := patchModifyPfdData(oldPfdData, pfdData); pd != nil {
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		c.JSON(int(pd.Status), pd)
		return
	}

	pfdDataForAppExt := convertPfdDataToPfdDataForAppExt(oldPfdData)
	if pfdReport := p.storePfdDataToUDR(appID, pfdDataForAppExt); pfdReport != nil {
		c.Set(sbi.IN_PB_DETAILS_CTX_STR, pfdReport.FailureCode)
		c.JSON(http.StatusInternalServerError, pfdReport)
		return
	}
	oldPfdData.Self = p.genPfdDataURI(scsAsID, transID, appID)
	pfdNotifyContext.AddNotification(appID, &models.PfdChangeNotification{
		ApplicationId: appID,
		Pfds:          pfdDataForAppExt.Pfds,
	})

	c.JSON(http.StatusOK, oldPfdData)
}

func (p *Processor) buildPfdManagement(afID string, afPfdTr *nef_context.AfPfdTransaction) (
	*models.PfdManagement, *models.ProblemDetails,
) {
	transID := afPfdTr.TransID
	appIDs := afPfdTr.GetExtAppIDs()
	pfdMng := &models.PfdManagement{
		Self:     p.genPfdManagementURI(afID, transID),
		PfdDatas: make(map[string]models.PfdData, len(appIDs)),
	}

	data, pd, err := p.Consumer().AppDataPfdsGet(appIDs)

	switch {
	case pd != nil:
		return nil, pd
	case err != nil:
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Detail: "Query to UDR failed",
		}
		return nil, problemDetails
	}

	for _, pfdDataForApp := range data {
		pfdData := convertPdfDataForAppExtToPfdData(&pfdDataForApp)
		pfdData.Self = p.genPfdDataURI(afID, transID, pfdData.ExternalAppId)

		pfdMng.PfdDatas[pfdData.ExternalAppId] = *pfdData
	}
	return pfdMng, nil
}

func (p *Processor) storePfdDataToUDR(appID string, pfdDataForApp *models.PfdDataForAppExt) *models.PfdReport {
	_, pd, errAppData := p.Consumer().AppDataPfdsAppIdPut(appID, pfdDataForApp)

	switch {
	case errAppData != nil || pd != nil:
		return &models.PfdReport{
			ExternalAppIds: []string{appID},
			FailureCode:    models.FailureCode_MALFUNCTION,
		}
	}

	return nil
}

func (p *Processor) deletePfdDataFromUDR(appID string) *models.ProblemDetails {
	pd, err := p.Consumer().AppDataPfdsAppIdDelete(appID)

	switch {
	case pd != nil:
		return pd
	case err != nil:
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Detail: "Query to UDR failed",
		}
		return problemDetails
	}

	return nil
}

// The behavior of PATCH update is based on TS 29.250 v1.15.1 clause 4.4.1
func patchModifyPfdData(oldPfdData, newPfdData *models.PfdData) *models.ProblemDetails {
	for pfdID, newPfd := range newPfdData.Pfds {
		_, exist := oldPfdData.Pfds[pfdID]
		if len(newPfd.FlowDescriptions) == 0 && len(newPfd.Urls) == 0 && len(newPfd.DomainNames) == 0 {
			if exist {
				// New Pfd with existing PfdID and empty content implies deletion from old PfdData.
				delete(oldPfdData.Pfds, pfdID)
			} else {
				// Otherwire, if the PfdID doesn't exist yet, the Pfd still needs valid content.
				return openapi.ProblemDetailsDataNotFound(DetailNoPfdInfo)
			}
		} else {
			// Either add or update the Pfd to the old PfdData.
			oldPfdData.Pfds[pfdID] = newPfd
		}
	}
	return nil
}

func convertPfdDataToPfdDataForApp(pfdData *models.PfdData) *models.PfdDataForAppExt {
	pfdDataForApp := &models.PfdDataForAppExt{
		ApplicationId: pfdData.ExternalAppId,
	}
	for _, pfd := range pfdData.Pfds {
		var pfdContent models.PfdContent
		pfdContent.PfdId = pfd.PfdId
		pfdContent.FlowDescriptions = pfd.FlowDescriptions
		pfdContent.Urls = pfd.Urls
		pfdContent.DomainNames = pfd.DomainNames
		pfdDataForApp.Pfds = append(pfdDataForApp.Pfds, pfdContent)
	}
	return pfdDataForApp
}

func convertPdfDataForAppExtToPfdData(pfdDataForAppExt *models.PfdDataForAppExt) *models.PfdData {
	pfdData := &models.PfdData{
		ExternalAppId: pfdDataForAppExt.ApplicationId,
		Pfds:          make(map[string]models.Pfd, len(pfdDataForAppExt.Pfds)),
	}
	for _, pfdContent := range pfdDataForAppExt.Pfds {
		var pfd models.Pfd
		pfd.PfdId = pfdContent.PfdId
		pfd.FlowDescriptions = pfdContent.FlowDescriptions
		pfd.Urls = pfdContent.Urls
		pfd.DomainNames = pfdContent.DomainNames
		pfdData.Pfds[pfdContent.PfdId] = pfd
	}
	return pfdData
}

func convertPdfDataForAppExtToPfdDataForApp(pfdDataForAppExt *models.PfdDataForAppExt) *models.PfdDataForApp {
	pfdDataForApp := &models.PfdDataForApp{
		ApplicationId: pfdDataForAppExt.ApplicationId,
	}
	for _, pfdContent := range pfdDataForAppExt.Pfds {
		var pfd models.PfdContent
		pfd.PfdId = pfdContent.PfdId
		pfd.FlowDescriptions = pfdContent.FlowDescriptions
		pfd.Urls = pfdContent.Urls
		pfd.DomainNames = pfdContent.DomainNames
		pfdDataForApp.Pfds = append(pfdDataForApp.Pfds, pfdContent)
	}
	return pfdDataForApp
}

func convertPfdDataToPfdDataForAppExt(pfdData *models.PfdData) *models.PfdDataForAppExt {
	pfdDataForAppExt := &models.PfdDataForAppExt{
		ApplicationId: pfdData.ExternalAppId,
	}
	for _, pfd := range pfdData.Pfds {
		var pfdContent models.PfdContent
		pfdContent.PfdId = pfd.PfdId
		pfdContent.FlowDescriptions = pfd.FlowDescriptions
		pfdContent.Urls = pfd.Urls
		pfdContent.DomainNames = pfd.DomainNames
		pfdDataForAppExt.Pfds = append(pfdDataForAppExt.Pfds, pfdContent)
	}
	return pfdDataForAppExt
}

func (p *Processor) genPfdManagementURI(afID, transID string) string {
	// E.g. https://localhost:29505/3gpp-pfd-management/v1/{afID}/transactions/{transID}
	return fmt.Sprintf("%s/%s/transactions/%s",
		p.Config().ServiceUri(factory.ServicePfdMng), afID, transID)
}

func (p *Processor) genPfdDataURI(afID, transID, appID string) string {
	// E.g. https://localhost:29505/3gpp-pfd-management/v1/{afID}/transactions/{transID}/applications/{appID}
	return fmt.Sprintf("%s/%s/transactions/%s/applications/%s",
		p.Config().ServiceUri(factory.ServicePfdMng), afID, transID, appID)
}

func validatePfdManagement(
	afID, transID string,
	pfdMng *models.PfdManagement,
	nefCtx *nef_context.NefContext,
) *models.ProblemDetails {
	pfdMng.PfdReports = make(map[string]models.PfdReport)

	if len(pfdMng.PfdDatas) == 0 {
		return openapi.ProblemDetailsDataNotFound(DetailNoPfdData)
	}

	for appID, pfdData := range pfdMng.PfdDatas {
		// Check whether the received external Application Identifier(s) are already provisioned
		appAfID, appTransID, ok := nefCtx.IsAppIDExisted(appID)
		if ok && (appAfID != afID || appTransID != transID) {
			delete(pfdMng.PfdDatas, appID)
			addPfdReport(pfdMng, &models.PfdReport{
				ExternalAppIds: []string{appID},
				FailureCode:    models.FailureCode_APP_ID_DUPLICATED,
			})
		}
		if pd := validatePfdData(&pfdData, false); pd != nil {
			return pd
		}
	}

	if len(pfdMng.PfdDatas) == 0 {
		// The PFDs for all applications were not created successfully.
		// PfdReport is included with detailed information.
		return openapi.ProblemDetailsSystemFailure("None of the PFDs were created")
	}
	return nil
}

func validatePfdData(pfdData *models.PfdData, isPatch bool) *models.ProblemDetails {
	if pfdData.ExternalAppId == "" {
		return openapi.ProblemDetailsDataNotFound(DetailNoExtAppID)
	}

	if len(pfdData.Pfds) == 0 {
		return openapi.ProblemDetailsDataNotFound(DetailNoPfd)
	}

	for _, pfd := range pfdData.Pfds {
		if pfd.PfdId == "" {
			return openapi.ProblemDetailsDataNotFound(DetailNoPfdID)
		}
		// For PATCH method, empty these three attributes is used to imply the deletion of this PFD
		if !isPatch && len(pfd.FlowDescriptions) == 0 && len(pfd.Urls) == 0 && len(pfd.DomainNames) == 0 {
			return openapi.ProblemDetailsDataNotFound(DetailNoPfdInfo)
		}
	}

	return nil
}

func addPfdReport(pfdMng *models.PfdManagement, newReport *models.PfdReport) {
	if oldReport, ok := pfdMng.PfdReports[string(newReport.FailureCode)]; ok {
		oldReport.ExternalAppIds = append(oldReport.ExternalAppIds, newReport.ExternalAppIds...)
	} else {
		pfdMng.PfdReports[string(newReport.FailureCode)] = *newReport
	}
}
