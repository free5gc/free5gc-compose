package consumer

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/nrf/NFDiscovery"
	"github.com/free5gc/openapi/pcf/PolicyAuthorization"
	sbi_metrics "github.com/free5gc/util/metrics/sbi"
)

type npcfService struct {
	consumer *Consumer

	mu      sync.RWMutex
	clients map[string]*PolicyAuthorization.APIClient
}

func (s *npcfService) getPolicyAuthClient(uri string) *PolicyAuthorization.APIClient {
	s.mu.RLock()
	if client, ok := s.clients[uri]; ok {
		defer s.mu.RUnlock()
		return client
	}

	configuration := PolicyAuthorization.NewConfiguration()
	configuration.SetBasePath(uri)
	configuration.SetMetrics(sbi_metrics.SbiMetricHook)
	configuration.SetHTTPClient(http.DefaultClient)
	cli := PolicyAuthorization.NewAPIClient(configuration)

	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[uri] = cli
	return cli
}

func (s *npcfService) getPcfPolicyAuthUri() (string, error) {
	uri := s.consumer.Context().PcfPaUri()
	if uri == "" {
		localVarOptionals := NFDiscovery.SearchNFInstancesRequest{
			ServiceNames: []models.ServiceName{
				models.ServiceName_NPCF_POLICYAUTHORIZATION,
			},
		}
		logger.ConsumerLog.Infoln(s.consumer.Config().NrfUri())
		_, sUri, err := s.consumer.SearchNFInstances(
			s.consumer.Config().NrfUri(),
			models.ServiceName_NPCF_POLICYAUTHORIZATION,
			models.NrfNfManagementNfType_PCF,
			models.NrfNfManagementNfType_NEF,
			&localVarOptionals,
		)
		if err == nil {
			s.consumer.Context().SetPcfPaUri(sUri)
		}
		logger.ConsumerLog.Debugf("Search NF Instances failed")
		return sUri, err
	}
	return uri, nil
}

// GetAppSession Reads an existing Individual Application Session Context resource.
// 3GPP TS 29.514 release 17 version 17.6.0
// Resource structure: 5.3.1
// Request/Response: 5.3.3.3.1
func (s *npcfService) GetAppSession(appSessionId string) (
	*models.AppSessionContext, *models.ProblemDetails, error,
) {
	uri, err := s.getPcfPolicyAuthUri()
	if err != nil {
		return nil, nil, err
	}

	client := s.getPolicyAuthClient(uri)

	if client == nil {
		return nil, nil, openapi.ReportError("could not initialize the PolicyAuthorization client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(
		models.ServiceName_NPCF_POLICYAUTHORIZATION, models.NrfNfManagementNfType_PCF)
	if err != nil {
		return nil, nil, err
	}

	appSessionsRequest := PolicyAuthorization.GetAppSessionRequest{
		AppSessionId: &appSessionId,
	}

	getAppSessionRsp, errGetAppSessionRsp := client.IndividualApplicationSessionContextDocumentApi.
		GetAppSession(ctx, &appSessionsRequest)

	if errGetAppSessionRsp != nil {
		switch apiErr := errGetAppSessionRsp.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case PolicyAuthorization.GetAppSessionError:
				// TODO: handle the 307/308 http status code
				return nil, &errorModel.ProblemDetails, nil
			case error:
				return nil, openapi.ProblemDetailsSystemFailure(errorModel.Error()), nil
			default:
				return nil, nil, openapi.ReportError("openapi error")
			}
		case error:
			return nil, openapi.ProblemDetailsSystemFailure(apiErr.Error()), nil
		default:
			return nil, nil, openapi.ReportError("server no response")
		}
	}

	return &getAppSessionRsp.AppSessionContext, nil, nil
}

// PostAppSessions Creates a models.AppSessionContext in the NPCF_policyAuthorization service.
// 3GPP TS 29.514 release 17 version 17.6.0
// Resource structure: 5.3.1
// Request/Response: 5.3.2.3.1
func (s *npcfService) PostAppSessions(asc *models.AppSessionContext) (string, *models.ProblemDetails, error) {
	uri, err := s.getPcfPolicyAuthUri()
	if err != nil {
		return "", nil, err
	}

	client := s.getPolicyAuthClient(uri)

	if client == nil {
		return "", nil, openapi.ReportError("could not initialize the PolicyAuthorization client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(
		models.ServiceName_NPCF_POLICYAUTHORIZATION, models.NrfNfManagementNfType_PCF)
	if err != nil {
		return "", nil, err
	}

	appSessionsRequest := PolicyAuthorization.PostAppSessionsRequest{
		AppSessionContext: asc,
	}

	postAppSessionsRsp, errPostAppSessionRsp := client.ApplicationSessionsCollectionApi.
		PostAppSessions(ctx, &appSessionsRequest)

	if errPostAppSessionRsp != nil {
		switch apiErr := errPostAppSessionRsp.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case PolicyAuthorization.PostAppSessionsError:
				return "", &errorModel.ProblemDetails, nil
			case error:
				return "", openapi.ProblemDetailsSystemFailure(errorModel.Error()), nil
			default:
				return "", nil, openapi.ReportError("openapi error")
			}
		case error:
			return "", openapi.ProblemDetailsSystemFailure(apiErr.Error()), nil
		default:
			return "", nil, openapi.ReportError("server no response")
		}
	}

	var sessId string

	if postAppSessionsRsp != nil {
		sessId = getAppSessIDFromRspLocationHeader(postAppSessionsRsp.Location)
	}

	return sessId, nil, nil
}

func (s *npcfService) PutAppSession(
	appSessionId string,
	ascUpdateData *models.AppSessionContextUpdateData,
	asc *models.AppSessionContext,
) (int, interface{}, string) {
	var (
		err     error
		rspCode int
		rspBody interface{}
		rsp     *PolicyAuthorization.GetAppSessionResponse
		modRsp  *PolicyAuthorization.ModAppSessionResponse
	)

	uri, err := s.getPcfPolicyAuthUri()
	if err != nil {
		return rspCode, rspBody, appSessionId
	}
	client := s.getPolicyAuthClient(uri)

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NPCF_POLICYAUTHORIZATION,
		models.NrfNfManagementNfType_PCF)
	if err != nil {
		return rspCode, rspBody, appSessionId
	}

	appSessReq := &PolicyAuthorization.GetAppSessionRequest{
		AppSessionId: &appSessionId,
	}
	rsp, err = client.IndividualApplicationSessionContextDocumentApi.
		GetAppSession(ctx, appSessReq)

	if rsp != nil {
		appSessModReq := &PolicyAuthorization.ModAppSessionRequest{
			AppSessionId: &appSessionId,
			AppSessionContextUpdateDataPatch: &models.AppSessionContextUpdateDataPatch{
				AscReqData: ascUpdateData,
			},
		}
		modRsp, err = client.IndividualApplicationSessionContextDocumentApi.ModAppSession(ctx, appSessModReq)

		if modRsp != nil {
			if reflect.DeepEqual(modRsp.AppSessionContext, models.AppSessionContext{}) {
				rspCode = http.StatusNoContent
				rspBody = nil
			} else {
				rspCode = http.StatusOK
				rspBody = modRsp.AppSessionContext
				logger.ConsumerLog.Debugf("PostAppSessions RspData: %+v", rsp.AppSessionContext)
			}
		} else {
			rspCode, rspBody = handleAPIServiceNoResponse(err)
		}
	} else {
		// API Service Internal Error or Server No Response
		rspCode, rspBody = handleAPIServiceNoResponse(err)
	}

	return rspCode, rspBody, appSessionId
}

// PatchAppSession Updates a models.AppSessionContext and returns its representation.
// 3GPP TS 29.514 release 17 version 17.6.0
// Resource structure: 5.3.1-1
// Request/Response: 5.3.3.3.2
func (s *npcfService) PatchAppSession(appSessionId string,
	ascUpdateData *models.AppSessionContextUpdateData,
) (*models.AppSessionContext, *models.ProblemDetails, error) {
	uri, err := s.getPcfPolicyAuthUri()
	if err != nil {
		return nil, nil, err
	}

	client := s.getPolicyAuthClient(uri)

	if client == nil {
		return nil, nil, openapi.ReportError("could not initialize the PolicyAuthorization client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(
		models.ServiceName_NPCF_POLICYAUTHORIZATION, models.NrfNfManagementNfType_PCF)
	if err != nil {
		return nil, nil, err
	}

	appSessionCtxUpdateDataPatch := models.AppSessionContextUpdateDataPatch{AscReqData: ascUpdateData}

	modAppSessionReq := PolicyAuthorization.ModAppSessionRequest{
		AppSessionId:                     &appSessionId,
		AppSessionContextUpdateDataPatch: &appSessionCtxUpdateDataPatch,
	}

	modAppSessionRsp, errModAppSessionRsp := client.IndividualApplicationSessionContextDocumentApi.ModAppSession(
		ctx, &modAppSessionReq)

	if errModAppSessionRsp != nil {
		switch apiErr := errModAppSessionRsp.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case PolicyAuthorization.ModAppSessionError:
				// TODO: handle the 307/308 http status code
				return nil, &errorModel.ProblemDetails, nil
			case error:
				return nil, openapi.ProblemDetailsSystemFailure(errorModel.Error()), nil
			default:
				return nil, nil, openapi.ReportError("openapi error")
			}
		case error:
			return nil, openapi.ProblemDetailsSystemFailure(apiErr.Error()), nil
		default:
			return nil, nil, openapi.ReportError("server no response")
		}
	}

	logger.ConsumerLog.Debugf("PatchAppSessions RspData: %+v", modAppSessionRsp.AppSessionContext)

	return &modAppSessionRsp.AppSessionContext, nil, nil
}

// DeleteAppSession Sends out a deleteAppSession API request to the PCF and returns either a status code,
// a problemDetails or an error format.
// 3GPP TS 29.514 Release 17 version 17.6.0
// Resource structure 5.3.1
// Request/Response: 5.3.3.4.2
func (s *npcfService) DeleteAppSession(appSessionId string) (int, *models.ProblemDetails, error) {
	uri, err := s.getPcfPolicyAuthUri()
	if err != nil {
		return 0, nil, err
	}

	client := s.getPolicyAuthClient(uri)

	if client == nil {
		return 0, nil, openapi.ReportError("could not initialize the PolicyAuthorization client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(
		models.ServiceName_NPCF_POLICYAUTHORIZATION, models.NrfNfManagementNfType_PCF)
	if err != nil {
		return 0, nil, err
	}

	deleteAppSessionReq := PolicyAuthorization.DeleteAppSessionRequest{
		AppSessionId: &appSessionId,
		// Parameter is optional, to change when the PCF will handle it.
		PcfPolicyAuthorizationEventsSubscReqData: nil,
	}

	// Here the response do not have any interest as we do not return
	_, errDeleteAppSessRsp := client.IndividualApplicationSessionContextDocumentApi.DeleteAppSession(
		ctx, &deleteAppSessionReq)

	if errDeleteAppSessRsp != nil {
		var problemDetails *models.ProblemDetails
		switch apiErr := errDeleteAppSessRsp.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case PolicyAuthorization.DeleteAppSessionError:
				// TODO: handle the 307/308 http status code
				problemDetails = &errorModel.ProblemDetails
				return int(problemDetails.Status), problemDetails, nil
			case error:
				problemDetails = openapi.ProblemDetailsSystemFailure(errorModel.Error())
				return int(problemDetails.Status), problemDetails, nil
			default:
				return 0, nil, openapi.ReportError("openapi error")
			}
		case error:
			problemDetails = openapi.ProblemDetailsSystemFailure(apiErr.Error())
			return int(problemDetails.Status), problemDetails, nil
		default:
			return 0, nil, openapi.ReportError("server no response")
		}
	}

	// As per 5.4.1.3.3.5-3, we return StatusNoContent
	return http.StatusNoContent, nil, nil
}

func getAppSessIDFromRspLocationHeader(loc string) string {
	appSessID := ""
	if strings.Contains(loc, "http") {
		index := strings.LastIndex(loc, "/")
		appSessID = loc[index+1:]
	}
	logger.ConsumerLog.Infof("appSessID=%q", appSessID)
	return appSessID
}
