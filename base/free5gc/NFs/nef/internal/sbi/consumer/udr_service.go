package consumer

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/nrf/NFDiscovery"
	"github.com/free5gc/openapi/udr/DataRepository"
	sbi_metrics "github.com/free5gc/util/metrics/sbi"
)

type nudrService struct {
	consumer *Consumer

	mu      sync.RWMutex
	clients map[string]*DataRepository.APIClient
}

func (s *nudrService) getDataRepositoryClient(uri string) *DataRepository.APIClient {
	if uri == "" {
		return nil
	}

	s.mu.RLock()

	client, ok := s.clients[uri]

	if ok {
		defer s.mu.RUnlock()
		return client
	}

	configuration := DataRepository.NewConfiguration()
	configuration.SetBasePath(uri)
	configuration.SetMetrics(sbi_metrics.SbiMetricHook)
	configuration.SetHTTPClient(http.DefaultClient)
	client = DataRepository.NewAPIClient(configuration)

	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[uri] = client
	return client
}

func (s *nudrService) getUdrDrUri() (string, error) {
	uri := s.consumer.Context().UdrDrUri()
	if uri == "" {
		localVarOptionals := NFDiscovery.SearchNFInstancesRequest{
			ServiceNames: []models.ServiceName{
				models.ServiceName_NUDR_DR,
			},
		}
		_, sUri, err := s.consumer.SearchNFInstances(s.consumer.Config().NrfUri(),
			models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR, models.NrfNfManagementNfType_NEF, &localVarOptionals)
		if err == nil {
			s.consumer.Context().SetUdrDrUri(sUri)
		}
		return sUri, err
	}
	return uri, nil
}

// AppDataInfluenceDataGet Query the UDR to retrieve models.TrafficInfluData for each matching combination
// of the values of the elements of the array given in parameters.
// 3GPP TS 29.519 release 17 version 17.6.0
// Resource structure: 6.2.2
// Request/Response: 6.2.5.3.1
func (s *nudrService) AppDataInfluenceDataGet(influenceIDs []string) (
	[]models.TrafficInfluData, *models.ProblemDetails, error,
) {
	uri, err := s.getUdrDrUri()
	if err != nil {
		return nil, nil, err
	}

	client := s.getDataRepositoryClient(uri)

	if client == nil {
		return nil, nil, fmt.Errorf("could not initialize the DataRepository client")
	}

	param := DataRepository.ReadInfluenceDataRequest{
		InfluenceIds: influenceIDs,
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR)
	if err != nil {
		return nil, nil, err
	}

	influenceDataRsp, influenceDataErr := client.InfluenceDataStoreApi.ReadInfluenceData(ctx, &param)

	if influenceDataErr != nil {
		switch apiErr := influenceDataErr.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case DataRepository.ReadInfluenceDataError:
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

	return influenceDataRsp.TrafficInfluData, nil, nil
}

// AppDataInfluenceDataPut Stores the models.TrafficInfluData for the related influenceID.
// 3GPP TS 29.519 release 17 version 17.6.0
// Resource structure: 6.2.2
// Request/Response: 6.2.6.3.1
func (s *nudrService) AppDataInfluenceDataPut(influenceID string,
	tiData *models.TrafficInfluData,
) (*models.TrafficInfluData, *models.ProblemDetails, error) {
	uri, err := s.getUdrDrUri()
	if err != nil {
		return nil, nil, err
	}

	client := s.getDataRepositoryClient(uri)

	if client == nil {
		return nil, nil, openapi.ReportError("could not initialize the DataRepository client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR)
	if err != nil {
		return nil, nil, err
	}

	influenceDataReq := DataRepository.CreateOrReplaceIndividualInfluenceDataRequest{
		InfluenceId:      &influenceID,
		TrafficInfluData: tiData,
	}

	influenceDataResp, errInfluenceData := client.IndividualInfluenceDataDocumentApi.
		CreateOrReplaceIndividualInfluenceData(ctx, &influenceDataReq)

	if errInfluenceData != nil {
		switch apiErr := errInfluenceData.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case DataRepository.CreateOrReplaceIndividualInfluenceDataError:

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

	return &influenceDataResp.TrafficInfluData, nil, nil
}

// AppDataPfdsGet Retrieve PFDs for related application identifier(s).
// 3GPP TS 29.519 release 17 version 17.6.0
// Resource structure: 6.2.2
// Request/Response: 6.2.3.3.1
func (s *nudrService) AppDataPfdsGet(appIDs []string) ([]models.PfdDataForAppExt, *models.ProblemDetails, error) {
	uri, err := s.getUdrDrUri()
	if err != nil {
		return nil, nil, err
	}

	client := s.getDataRepositoryClient(uri)

	if client == nil {
		return nil, nil, openapi.ReportError("could not initialize the DataRepository client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR)
	if err != nil {
		return nil, nil, err
	}

	pfdDataReq := DataRepository.ReadPFDDataRequest{
		AppId: appIDs,
	}

	pfdDataResp, errPfdData := client.PFDDataStoreApi.ReadPFDData(ctx, &pfdDataReq)

	if errPfdData != nil {
		switch apiErr := errPfdData.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case DataRepository.ReadPFDDataError:
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

	return pfdDataResp.PfdDataForAppExt, nil, nil
}

// AppDataPfdsAppIdPut Creates, updates an individual PFD given an appId and the content to store into the UDR.
// 3GPP TS 29.519 release 17 version 17.6.0
// Resource structure: 6.2.2
// Request/Response: 6.2.4.3.3
func (s *nudrService) AppDataPfdsAppIdPut(appID string, pfdDataForApp *models.PfdDataForAppExt) (
	*models.PfdDataForAppExt, *models.ProblemDetails, error,
) {
	uri, err := s.getUdrDrUri()
	if err != nil {
		return nil, nil, err
	}

	client := s.getDataRepositoryClient(uri)

	if client == nil {
		return nil, nil, openapi.ReportError("could not initialize the DataRepository client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR)
	if err != nil {
		return nil, nil, err
	}

	individualPfdDataReq := DataRepository.CreateOrReplaceIndividualPFDDataRequest{
		AppId:            &appID,
		PfdDataForAppExt: pfdDataForApp,
	}

	individualPfdDataRsp, errIndividualPfdData := client.IndividualPFDDataDocumentApi.
		CreateOrReplaceIndividualPFDData(ctx, &individualPfdDataReq)

	if errIndividualPfdData != nil {
		switch apiErr := errIndividualPfdData.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case DataRepository.CreateOrReplaceIndividualPFDDataError:
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

	return &individualPfdDataRsp.PfdDataForAppExt, nil, nil
}

// AppDataPfdsAppIdDelete Deletes the individual PFD Data resource related to the application identifier.
// 3GPP TS 29.519 release 17 version 17.6.0
// Resource structure: 6.2.2
// Request/Response: 6.2.4.3.2
func (s *nudrService) AppDataPfdsAppIdDelete(appID string) (*models.ProblemDetails, error) {
	uri, err := s.getUdrDrUri()
	if err != nil {
		return nil, err
	}

	client := s.getDataRepositoryClient(uri)

	if client == nil {
		return nil, openapi.ReportError("could not initialize the DataRepository client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR)
	if err != nil {
		return nil, err
	}

	DeletePdfDataReq := DataRepository.DeleteIndividualPFDDataRequest{
		AppId: &appID,
	}

	_, errDeletePfdData := client.IndividualPFDDataDocumentApi.DeleteIndividualPFDData(ctx, &DeletePdfDataReq)

	if errDeletePfdData != nil {
		switch apiErr := errDeletePfdData.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case DataRepository.DeleteIndividualPFDDataError:
				return &errorModel.ProblemDetails, nil
			case error:
				return openapi.ProblemDetailsSystemFailure(errorModel.Error()), nil
			default:
				return nil, openapi.ReportError("openapi error")
			}
		case error:
			return openapi.ProblemDetailsSystemFailure(apiErr.Error()), nil
		default:
			return nil, openapi.ReportError("server no response")
		}
	}
	return nil, nil
}

// AppDataPfdsAppIdGet Returns a representation of PFDs for the related applicationID.
// 3GPP TS 29.519 release 17 version 17.6.0
// Resource structure: 6.2.2
// Request/Response: 6.2.4.3.1
func (s *nudrService) AppDataPfdsAppIdGet(appID string) (
	*DataRepository.ReadIndividualPFDDataResponse, *models.ProblemDetails, error,
) {
	uri, err := s.getUdrDrUri()
	if err != nil {
		return nil, nil, err
	}
	client := s.getDataRepositoryClient(uri)

	if client == nil {
		return nil, nil, openapi.ReportError("could not initialize the DataRepository client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR)
	if err != nil {
		return nil, nil, err
	}

	pfdDataReq := DataRepository.ReadIndividualPFDDataRequest{
		AppId: &appID,
	}

	pfdData, errPfdData := client.IndividualPFDDataDocumentApi.ReadIndividualPFDData(ctx, &pfdDataReq)

	if errPfdData != nil {
		switch apiErr := errPfdData.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case DataRepository.ReadIndividualPFDDataError:
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
	return pfdData, nil, nil
}

// AppDataInfluenceDataPatch Patch the TrafficInfluData for the related influenceID and tiSubPatch and returns it.
// 3GPP TS 29.519 release 17 version 17.6.0
// Resource structure: 6.2.2
// Request/Response: 6.2.6.3.2
func (s *nudrService) AppDataInfluenceDataPatch(
	influenceID string, tiSubPatch *models.TrafficInfluDataPatch,
) (*models.TrafficInfluData, *models.ProblemDetails, error) {
	uri, err := s.getUdrDrUri()
	if err != nil {
		return nil, nil, err
	}
	client := s.getDataRepositoryClient(uri)

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR)
	if err != nil {
		return nil, nil, err
	}

	tiDataReq := DataRepository.UpdateIndividualInfluenceDataRequest{
		InfluenceId:           &influenceID,
		TrafficInfluDataPatch: tiSubPatch,
	}

	trafficDataRsp, errTiData := client.IndividualInfluenceDataDocumentApi.UpdateIndividualInfluenceData(ctx, &tiDataReq)

	if errTiData != nil {
		switch apiErr := errTiData.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case DataRepository.UpdateIndividualInfluenceDataError:
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

	var trafficInfluData *models.TrafficInfluData

	if trafficDataRsp != nil {
		trafficInfluData = &trafficDataRsp.TrafficInfluData
	}

	return trafficInfluData, nil, nil
}

// AppDataInfluenceDataDelete Deletes the TrafficInfluenceData for the related influenceID.
// 3GPP TS 29.519 release 17 version 17.6.0
// Resource structure: 6.2.2
// Request/Response: 6.2.6.3.3
func (s *nudrService) AppDataInfluenceDataDelete(influenceID string) (*models.ProblemDetails, error) {
	uri, err := s.getUdrDrUri()
	if err != nil {
		return nil, err
	}
	client := s.getDataRepositoryClient(uri)

	if client == nil {
		return nil, openapi.ReportError("could not initialize the DataRepository client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDR_DR, models.NrfNfManagementNfType_UDR)
	if err != nil {
		return nil, err
	}

	deleteInfluenceReq := DataRepository.DeleteIndividualInfluenceDataRequest{
		InfluenceId: &influenceID,
	}

	_, errDeleteInfluenceData := client.IndividualInfluenceDataDocumentApi.
		DeleteIndividualInfluenceData(ctx, &deleteInfluenceReq)

	if errDeleteInfluenceData != nil {
		switch apiErr := errDeleteInfluenceData.(type) {
		// API error
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case DataRepository.DeleteIndividualInfluenceDataError:
				return &errorModel.ProblemDetails, nil
			case error:
				return openapi.ProblemDetailsSystemFailure(errorModel.Error()), nil
			default:
				return nil, openapi.ReportError("openapi error")
			}
		case error:
			return openapi.ProblemDetailsSystemFailure(apiErr.Error()), nil
		default:
			return nil, openapi.ReportError("server no response")
		}
	}

	return nil, nil
}
