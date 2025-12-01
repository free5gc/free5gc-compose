package consumer

import (
	"github.com/free5gc/openapi/amf/Communication"
	"github.com/free5gc/openapi/chf/ConvergedCharging"
	"github.com/free5gc/openapi/nrf/NFDiscovery"
	"github.com/free5gc/openapi/nrf/NFManagement"
	"github.com/free5gc/openapi/pcf/SMPolicyControl"
	"github.com/free5gc/openapi/smf/PDUSession"
	"github.com/free5gc/openapi/udm/SubscriberDataManagement"
	"github.com/free5gc/openapi/udm/UEContextManagement"
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/pkg/app"
)

type Consumer struct {
	app.App

	// consumer services
	*nsmfService
	*namfService
	*nchfService
	*npcfService
	*nudmService
	*nnrfService
	*nbsfService // BSF service for PCF binding discovery
}

func NewConsumer(smf app.App) (*Consumer, error) {
	c := &Consumer{
		App: smf,
	}

	c.nsmfService = &nsmfService{
		consumer:          c,
		PDUSessionClients: make(map[string]*PDUSession.APIClient),
	}

	c.namfService = &namfService{
		consumer:             c,
		CommunicationClients: make(map[string]*Communication.APIClient),
	}

	c.nchfService = &nchfService{
		consumer:                 c,
		ConvergedChargingClients: make(map[string]*ConvergedCharging.APIClient),
	}

	c.nudmService = &nudmService{
		consumer:                        c,
		SubscriberDataManagementClients: make(map[string]*SubscriberDataManagement.APIClient),
		UEContextManagementClients:      make(map[string]*UEContextManagement.APIClient),
	}

	c.nnrfService = &nnrfService{
		consumer:            c,
		NFManagementClients: make(map[string]*NFManagement.APIClient),
		NFDiscoveryClients:  make(map[string]*NFDiscovery.APIClient),
	}

	c.npcfService = &npcfService{
		consumer:               c,
		SMPolicyControlClients: make(map[string]*SMPolicyControl.APIClient),
	}

	c.nbsfService = &nbsfService{
		consumer: c,
	}

	return c, nil
}

// BSFAwarePCFSelection performs PCF selection with BSF binding awareness
func (c *Consumer) BSFAwarePCFSelection(smContext *smf_context.SMContext) error {
	return c.nbsfService.PCFSelectionWithBSF(smContext)
}

// NotifyBSFBindingRelease notifies BSF about PCF binding release
func (c *Consumer) NotifyBSFBindingRelease(smContext *smf_context.SMContext) {
	c.nbsfService.NotifyPCFBindingRelease(smContext)
}
