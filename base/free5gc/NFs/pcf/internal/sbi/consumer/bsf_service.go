package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/nrf/NFDiscovery"
	pcf_context "github.com/free5gc/pcf/internal/context"
	"github.com/free5gc/pcf/internal/logger"
	"github.com/free5gc/pcf/pkg/factory"
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

// isBSFEnabled checks if BSF integration is enabled in configuration
func (s *nbsfService) isBSFEnabled() bool {
	config := factory.PcfConfig

	// Default to false if BSF config is not specified
	if config.Configuration.Bsf == nil {
		return false
	}

	return config.Configuration.Bsf.Enable
}

// BSFSelection discovers and selects BSF for PCF binding registration
func (s *nbsfService) BSFSelection() (string, error) {
	// Discover BSF via NRF
	targetNfType := models.NrfNfManagementNfType_BSF
	requesterNfType := models.NrfNfManagementNfType_PCF
	request := NFDiscovery.SearchNFInstancesRequest{
		TargetNfType:    &targetNfType,
		RequesterNfType: &requesterNfType,
	}

	// Use existing NRF discovery service
	result, err := s.consumer.nnrfService.SendSearchNFInstances(
		s.consumer.Context().NrfUri,
		targetNfType,
		requesterNfType,
		request,
	)
	if err != nil {
		return "", fmt.Errorf("BSF discovery failed: %w", err)
	}

	if result == nil || len(result.NfInstances) == 0 {
		return "", fmt.Errorf("no BSF instances found")
	}

	// Select first BSF instance and find management service URI
	bsfProfile := result.NfInstances[0]
	for _, service := range bsfProfile.NfServices {
		if service.ServiceName == models.ServiceName_NBSF_MANAGEMENT &&
			service.NfServiceStatus == models.NfServiceStatus_REGISTERED {
			return service.ApiPrefix, nil
		}
	}

	return "", fmt.Errorf("BSF management service not found")
}

// RegisterPCFBinding registers PCF binding for a PDU session context
// Returns empty string and no error if BSF is not available (fallback mode)
func (s *nbsfService) RegisterPCFBinding(smPolicyData *pcf_context.UeSmPolicyData) (string, error) {
	// Check if BSF is enabled in configuration
	if !s.isBSFEnabled() {
		logger.ConsumerLog.Debugf("BSF integration is disabled in configuration")
		return "", nil
	}

	bsfUri, err := s.BSFSelection()
	if err != nil {
		logger.ConsumerLog.Warnf("BSF selection failed, continuing without BSF integration: %v", err)
		return "", nil // Return success with empty binding ID to continue without BSF
	}

	// Extract data from policy context
	ctx := smPolicyData.PolicyContext
	if ctx == nil {
		return "", fmt.Errorf("policy context is nil")
	}

	// Get PCF service information
	pcfContext := s.consumer.Context()
	var pcfFqdn string
	var pcfIpEndPoints []models.IpEndPoint

	if policyAuthService, exists := pcfContext.NfService[models.ServiceName_NPCF_POLICYAUTHORIZATION]; exists {
		pcfFqdn = policyAuthService.ApiPrefix
		pcfIpEndPoints = policyAuthService.IpEndPoints
	} else {
		// Fallback to basic PCF info
		pcfFqdn = fmt.Sprintf("%s://%s:%d", pcfContext.UriScheme, pcfContext.RegisterIPv4, pcfContext.SBIPort)
		pcfIpEndPoints = []models.IpEndPoint{
			{
				Ipv4Address: pcfContext.RegisterIPv4,
				Port:        int32(pcfContext.SBIPort),
			},
		}
	}

	// Create binding request
	binding := models.PcfBinding{
		Supi:           ctx.Supi,
		Gpsi:           ctx.Gpsi,
		Ipv4Addr:       ctx.Ipv4Address,
		Ipv6Prefix:     ctx.Ipv6AddressPrefix,
		IpDomain:       ctx.IpDomain,
		Dnn:            ctx.Dnn,
		Snssai:         ctx.SliceInfo,
		PcfFqdn:        pcfFqdn,
		PcfIpEndPoints: pcfIpEndPoints,
		PcfId:          pcfContext.NfId,
	}

	// Marshal binding request
	jsonData, err := json.Marshal(binding)
	if err != nil {
		return "", fmt.Errorf("failed to marshal binding: %w", err)
	}

	// Create HTTP request
	createURL := fmt.Sprintf("%s/nbsf-management/v1/pcfBindings", bsfUri)
	req, err := http.NewRequestWithContext(context.Background(), "POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	client := s.getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.ConsumerLog.Errorf("Failed to register PCF binding in BSF: %v", err)
		return "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.ConsumerLog.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusCreated {
		// Extract binding ID from Location header
		location := resp.Header.Get("Location")
		if location != "" {
			// Extract binding ID from location path like: /nbsf-management/v1/pcfBindings/{bindingId}
			parts := strings.Split(location, "/")
			if len(parts) > 0 {
				bindingId := parts[len(parts)-1]
				logger.ConsumerLog.Infof("Successfully registered PCF binding in BSF: %s", bindingId)
				return bindingId, nil
			}
		}
		logger.ConsumerLog.Infof("Successfully registered PCF binding in BSF for SUPI: %s, DNN: %s", ctx.Supi, ctx.Dnn)
		return "", nil // Binding created but ID not extracted
	}

	return "", fmt.Errorf("unexpected response status: %d", resp.StatusCode)
}

// UpdatePCFBinding updates existing PCF binding in BSF
// Returns nil (success) if BSF is not available or bindingId is empty (fallback mode)
func (s *nbsfService) UpdatePCFBinding(bindingId string, smPolicyData *pcf_context.UeSmPolicyData) error {
	// Check if BSF is enabled in configuration
	if !s.isBSFEnabled() {
		logger.ConsumerLog.Debugf("BSF integration is disabled in configuration")
		return nil
	}

	// If no binding ID, skip BSF update (fallback mode)
	if bindingId == "" {
		logger.ConsumerLog.Debugf("No BSF binding ID provided, skipping BSF update")
		return nil
	}

	bsfUri, err := s.BSFSelection()
	if err != nil {
		logger.ConsumerLog.Warnf("BSF selection failed, skipping BSF update: %v", err)
		return nil // Return success to continue without BSF
	}

	// Extract data from policy context for updates
	ctx := smPolicyData.PolicyContext
	if ctx == nil {
		return fmt.Errorf("policy context is nil")
	}

	// Create patch updates with current context data
	updates := map[string]interface{}{}

	// Only include non-empty fields for updates
	if ctx.Ipv4Address != "" {
		updates["ipv4Addr"] = ctx.Ipv4Address
	}
	if ctx.Ipv6AddressPrefix != "" {
		updates["ipv6Prefix"] = ctx.Ipv6AddressPrefix
	}
	if ctx.IpDomain != "" {
		updates["ipDomain"] = ctx.IpDomain
	}

	// Add PCF info updates
	pcfContext := s.consumer.Context()
	if policyAuthService, exists := pcfContext.NfService[models.ServiceName_NPCF_POLICYAUTHORIZATION]; exists {
		updates["pcfFqdn"] = policyAuthService.ApiPrefix
		updates["pcfIpEndPoints"] = policyAuthService.IpEndPoints
	}

	// Marshal updates
	jsonData, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("failed to marshal updates: %w", err)
	}

	// Create HTTP request
	updateURL := fmt.Sprintf("%s/nbsf-management/v1/pcfBindings/%s", bsfUri, bindingId)
	req, err := http.NewRequestWithContext(context.Background(), "PATCH", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json-patch+json")
	req.Header.Set("Accept", "application/json")

	// Send request
	client := s.getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.ConsumerLog.Errorf("Failed to update PCF binding in BSF: %v", err)
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.ConsumerLog.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		logger.ConsumerLog.Infof("Successfully updated PCF binding %s in BSF", bindingId)
		return nil
	}

	return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
}

// DeletePCFBinding removes PCF binding from BSF
// Returns nil (success) if BSF is not available or bindingId is empty (fallback mode)
func (s *nbsfService) DeletePCFBinding(bindingId string) error {
	// Check if BSF is enabled in configuration
	if !s.isBSFEnabled() {
		logger.ConsumerLog.Debugf("BSF integration is disabled in configuration")
		return nil
	}

	// If no binding ID, skip BSF deletion (fallback mode)
	if bindingId == "" {
		logger.ConsumerLog.Debugf("No BSF binding ID provided, skipping BSF deletion")
		return nil
	}

	bsfUri, err := s.BSFSelection()
	if err != nil {
		logger.ConsumerLog.Warnf("BSF selection failed, skipping BSF deletion: %v", err)
		return nil // Return success to continue without BSF
	}

	// Create HTTP request
	deleteURL := fmt.Sprintf("%s/nbsf-management/v1/pcfBindings/%s", bsfUri, bindingId)
	req, err := http.NewRequestWithContext(context.Background(), "DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	client := s.getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.ConsumerLog.Errorf("Failed to delete PCF binding from BSF: %v", err)
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.ConsumerLog.Warnf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		logger.ConsumerLog.Infof("Successfully deleted PCF binding %s from BSF", bindingId)
		return nil
	}

	return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
}
