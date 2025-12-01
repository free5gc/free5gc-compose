package consumer

import (
	"context"

	"github.com/free5gc/openapi/amf/Communication"
	"github.com/free5gc/openapi/nrf/NFDiscovery"
	"github.com/free5gc/openapi/nrf/NFManagement"
	"github.com/free5gc/openapi/pcf/AMPolicyControl"
	"github.com/free5gc/openapi/udr/DataRepository"
	pcf_context "github.com/free5gc/pcf/internal/context"
	"github.com/free5gc/pcf/pkg/factory"
)

type pcf interface {
	Config() *factory.Config
	Context() *pcf_context.PCFContext
	CancelContext() context.Context
}

type Consumer struct {
	pcf

	// consumer services
	*nnrfService
	*namfService
	*nudrService
	*npcfService
	*nbsfService
}

func NewConsumer(pcf pcf) (*Consumer, error) {
	c := &Consumer{
		pcf: pcf,
	}

	c.nnrfService = &nnrfService{
		consumer:        c,
		nfMngmntClients: make(map[string]*NFManagement.APIClient),
		nfDiscClients:   make(map[string]*NFDiscovery.APIClient),
	}

	c.namfService = &namfService{
		consumer:     c,
		nfComClients: make(map[string]*Communication.APIClient),
	}

	c.nudrService = &nudrService{
		consumer:         c,
		nfDataSubClients: make(map[string]*DataRepository.APIClient),
	}

	c.npcfService = &npcfService{
		consumer:                c,
		nfAMPolicyControlClient: make(map[string]*AMPolicyControl.APIClient),
	}

	c.nbsfService = &nbsfService{
		consumer: c,
	}

	return c, nil
}

// BSF Service Methods
func (c *Consumer) RegisterPCFBinding(smPolicyData *pcf_context.UeSmPolicyData) (string, error) {
	return c.nbsfService.RegisterPCFBinding(smPolicyData)
}

func (c *Consumer) UpdatePCFBinding(bindingId string, smPolicyData *pcf_context.UeSmPolicyData) error {
	return c.nbsfService.UpdatePCFBinding(bindingId, smPolicyData)
}

func (c *Consumer) DeletePCFBinding(bindingId string) error {
	return c.nbsfService.DeletePCFBinding(bindingId)
}
