package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/nrf/NFDiscovery"
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/internal/logger"
)

type nbsfService struct {
	consumer *Consumer

	httpClientMu sync.RWMutex
	httpClient   *http.Client
}

func (s *nbsfService) getHTTPClient() *http.Client {
	s.httpClientMu.RLock()
	defer s.httpClientMu.RUnlock()
	if s.httpClient == nil {
		s.httpClient = &http.Client{}
	}
	return s.httpClient
}

// BSFSelection discovers and selects BSF for PCF binding operations
func (s *nbsfService) BSFSelection() (string, error) {
	// Discover BSF via NRF
	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NNRF_DISC, models.NrfNfManagementNfType_NRF)
	if err != nil {
		return "", fmt.Errorf("failed to get token context: %w", err)
	}

	targetNfType := models.NrfNfManagementNfType_BSF
	requesterNfType := models.NrfNfManagementNfType_SMF
	request := &NFDiscovery.SearchNFInstancesRequest{
		TargetNfType:    &targetNfType,
		RequesterNfType: &requesterNfType,
	}

	client := s.consumer.nnrfService.getNFDiscoveryClient(s.consumer.Context().NrfUri)
	res, err := client.NFInstancesStoreApi.SearchNFInstances(ctx, request)
	if err != nil {
		return "", fmt.Errorf("BSF discovery failed: %w", err)
	}

	if res == nil || len(res.SearchResult.NfInstances) == 0 {
		return "", fmt.Errorf("no BSF instances found")
	}

	// Select first BSF instance and find management service URI
	bsfProfile := res.SearchResult.NfInstances[0]
	for _, service := range bsfProfile.NfServices {
		if service.ServiceName == models.ServiceName_NBSF_MANAGEMENT &&
			service.NfServiceStatus == models.NfServiceStatus_REGISTERED {
			return service.ApiPrefix, nil
		}
	}

	return "", fmt.Errorf("BSF management service not found")
}

// QueryPCFBinding queries BSF for existing PCF binding
func (s *nbsfService) QueryPCFBinding(smContext *smf_context.SMContext) (*models.PcfBinding, error) {
	bsfUri, err := s.BSFSelection()
	if err != nil {
		logger.ConsumerLog.Warnf("BSF selection failed: %v", err)
		return nil, err
	}

	// Build query URL with parameters
	queryURL := fmt.Sprintf("%s/nbsf-management/v1/pcfBindings", bsfUri)
	params := url.Values{}

	// Query by SUPI and DNN combination
	if smContext.Supi != "" {
		params.Add("supi", smContext.Supi)
	}
	if smContext.Dnn != "" {
		params.Add("dnn", smContext.Dnn)
	}

	// Add S-NSSAI if available
	if smContext.SNssai != nil {
		snssaiStr := fmt.Sprintf(`{"sst":%d`, smContext.SNssai.Sst)
		if smContext.SNssai.Sd != "" {
			snssaiStr += fmt.Sprintf(`,"sd":"%s"`, smContext.SNssai.Sd)
		}
		snssaiStr += "}"
		params.Add("snssai", snssaiStr)
	}

	if len(params) > 0 {
		queryURL += "?" + params.Encode()
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(context.Background(), "GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	client := s.getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.ConsumerLog.Warnf("BSF query failed: %v", err)
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.ConsumerLog.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusNoContent {
		logger.ConsumerLog.Tracef("No existing PCF binding found for SUPI: %s, DNN: %s", smContext.Supi, smContext.Dnn)
		return nil, nil
	}

	if resp.StatusCode == http.StatusOK {
		var pcfBinding models.PcfBinding
		if decodeErr := json.NewDecoder(resp.Body).Decode(&pcfBinding); decodeErr != nil {
			return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
		}
		logger.ConsumerLog.Infof("Found existing PCF binding for SUPI: %s, DNN: %s", smContext.Supi, smContext.Dnn)

		// Store binding ID if provided in response header (enhancement)
		if bindingId := resp.Header.Get("X-BSF-Binding-ID"); bindingId != "" {
			smContext.BSFBindingID = bindingId
			logger.ConsumerLog.Debugf("Stored BSF binding ID: %s for SMContext", bindingId)
		}

		return &pcfBinding, nil
	}

	return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
}

// PCFSelectionWithBSF performs PCF selection using BSF binding discovery first
func (s *nbsfService) PCFSelectionWithBSF(smContext *smf_context.SMContext) error {
	// First, try to discover existing PCF binding via BSF
	pcfBinding, err := s.QueryPCFBinding(smContext)
	if err != nil {
		logger.ConsumerLog.Warnf("BSF query error (falling back to NRF): %v", err)
	}

	if pcfBinding != nil {
		// Found existing PCF binding, discover PCF NF profile from NRF using PCF ID
		logger.ConsumerLog.Infof("Using existing PCF from BSF binding: %s", pcfBinding.PcfId)

		// Discover the specific PCF instance from NRF
		ctx, _, tokenErr := s.consumer.Context().GetTokenCtx(models.ServiceName_NNRF_DISC, models.NrfNfManagementNfType_NRF)
		if tokenErr != nil {
			logger.ConsumerLog.Warnf("Failed to get token context for PCF discovery: %v", tokenErr)
			return s.consumer.nnrfService.PCFSelection(smContext)
		}

		targetNfType := models.NrfNfManagementNfType_PCF
		requesterNfType := models.NrfNfManagementNfType_SMF
		request := &NFDiscovery.SearchNFInstancesRequest{
			TargetNfType:       &targetNfType,
			RequesterNfType:    &requesterNfType,
			TargetNfInstanceId: &pcfBinding.PcfId,
		}

		client := s.consumer.nnrfService.getNFDiscoveryClient(s.consumer.Context().NrfUri)
		res, searchErr := client.NFInstancesStoreApi.SearchNFInstances(ctx, request)
		if searchErr != nil || res == nil || len(res.SearchResult.NfInstances) == 0 {
			logger.ConsumerLog.Warnf("Failed to discover PCF %s from NRF, falling back to general PCF selection: %v",
				pcfBinding.PcfId, searchErr)
			return s.consumer.nnrfService.PCFSelection(smContext)
		}

		// Use the discovered PCF
		smContext.SelectedPCFProfile = res.SearchResult.NfInstances[0]
		logger.ConsumerLog.Infof("Successfully discovered PCF %s for existing BSF binding", pcfBinding.PcfId)
		return nil
	}

	// No existing binding found, use traditional NRF-based PCF selection
	logger.ConsumerLog.Infof("No PCF binding found in BSF, using NRF discovery for SUPI: %s, DNN: %s",
		smContext.Supi, smContext.Dnn)
	return s.consumer.nnrfService.PCFSelection(smContext)
}

// NotifyPCFBindingRelease notifies BSF about PCF binding release when SMF session terminates
// This is called when SMContext.BSFBindingID is available and the session is being cleaned up
func (s *nbsfService) NotifyPCFBindingRelease(smContext *smf_context.SMContext) {
	if smContext.BSFBindingID == "" {
		// No binding ID stored, nothing to notify
		return
	}

	logger.ConsumerLog.Infof("Notifying BSF about PCF binding release for binding ID: %s", smContext.BSFBindingID)

	// Note: In a complete implementation, this could trigger PCF to delete the binding
	// For now, we just log the event. The actual deletion should be handled by PCF
	// when it receives SMF context termination notification.

	// Clear the binding ID from context
	smContext.BSFBindingID = ""
}
