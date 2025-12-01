package webui_context

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	Nnrf_NFDiscovery "github.com/free5gc/openapi/nrf/NFDiscovery"
	Nnrf_NFManagement "github.com/free5gc/openapi/nrf/NFManagement"
	"github.com/free5gc/openapi/oauth"
	"github.com/free5gc/webconsole/backend/factory"
	"github.com/free5gc/webconsole/backend/logger"
)

var webuiContext WEBUIContext

type WEBUIContext struct {
	NfInstanceID   string
	NFProfiles     []models.NrfNfDiscoveryNfProfile
	NFOamInstances []NfOamInstance

	// is registered to NRF as AF
	IsRegistered bool

	NrfUri         string
	OAuth2Required bool

	NFManagementClient *Nnrf_NFManagement.APIClient
	NFDiscoveryClient  *Nnrf_NFDiscovery.APIClient
}

type NfOamInstance struct {
	NfId   string
	NfType models.NrfNfManagementNfType
	Uri    string
}

func Init() {
	webuiContext.NfInstanceID = uuid.New().String()
	webuiContext.NrfUri = factory.WebuiConfig.Configuration.NrfUri

	webuiContext.IsRegistered = false

	ManagementConfig := Nnrf_NFManagement.NewConfiguration()
	ManagementConfig.SetBasePath(GetSelf().NrfUri)
	webuiContext.NFManagementClient = Nnrf_NFManagement.NewAPIClient(ManagementConfig)

	NFDiscovryConfig := Nnrf_NFDiscovery.NewConfiguration()
	NFDiscovryConfig.SetBasePath(GetSelf().NrfUri)
	webuiContext.NFDiscoveryClient = Nnrf_NFDiscovery.NewAPIClient(NFDiscovryConfig)
}

func (context *WEBUIContext) UpdateNfProfiles() {
	var nfProfiles []models.NrfNfDiscoveryNfProfile

	nfProfiles, err := SendSearchNFInstances(models.NrfNfManagementNfType_AMF)
	if err != nil {
		logger.CtxLog.Error(err)
		return
	}
	context.NFProfiles = append(context.NFProfiles, nfProfiles...)

	nfProfiles, err = SendSearchNFInstances(models.NrfNfManagementNfType_SMF)
	if err != nil {
		logger.CtxLog.Error(err)
		return
	}
	context.NFProfiles = append(context.NFProfiles, nfProfiles...)

	for idx := 0; idx < len(context.NFProfiles); idx++ {
		nfProfile := &context.NFProfiles[idx]

		if nfProfile.NfServices == nil || context.NfProfileAlreadyExists(nfProfile) {
			continue
		}

		var uri string
		switch nfProfile.NfType {
		case models.NrfNfManagementNfType_AMF:
			uri = getNfOamUri(nfProfile, models.ServiceName("namf-oam"))
		case models.NrfNfManagementNfType_SMF:
			uri = getNfOamUri(nfProfile, models.ServiceName("nsmf-oam"))
		}
		if uri != "" {
			context.NFOamInstances = append(context.NFOamInstances, NfOamInstance{
				NfId:   nfProfile.NfInstanceId,
				NfType: nfProfile.NfType,
				Uri:    uri,
			})
		}
	}
}

func (context *WEBUIContext) NfProfileAlreadyExists(nfProfile *models.NrfNfDiscoveryNfProfile) bool {
	for _, instance := range context.NFOamInstances {
		if instance.NfId == nfProfile.NfInstanceId {
			return true
		}
	}
	return false
}

func getNfOamUri(nfProfile *models.NrfNfDiscoveryNfProfile, serviceName models.ServiceName) (nfOamUri string) {
	for _, service := range nfProfile.NfServices {
		if service.ServiceName == serviceName && service.NfServiceStatus == models.NfServiceStatus_REGISTERED {
			if nfProfile.Fqdn != "" {
				nfOamUri = nfProfile.Fqdn
			} else if service.Fqdn != "" {
				nfOamUri = service.Fqdn
			} else if service.ApiPrefix != "" {
				nfOamUri = service.ApiPrefix
			} else if service.IpEndPoints != nil {
				point := service.IpEndPoints[0]
				if point.Ipv4Address != "" {
					nfOamUri = getSbiUri(service.Scheme, point.Ipv4Address, point.Port)
				} else if len(nfProfile.Ipv4Addresses) != 0 {
					nfOamUri = getSbiUri(service.Scheme, nfProfile.Ipv4Addresses[0], point.Port)
				}
			}
		}
		if nfOamUri != "" {
			break
		}
	}
	return
}

func (context *WEBUIContext) GetOamUris(targetNfType models.NrfNfManagementNfType) (uris []string) {
	for _, oamInstance := range context.NFOamInstances {
		if oamInstance.NfType == targetNfType {
			uris = append(uris, oamInstance.Uri)
			break
		}
	}
	return
}

func GetSelf() *WEBUIContext {
	return &webuiContext
}

func getSbiUri(scheme models.UriScheme, ipv4Address string, port int32) (uri string) {
	if port != 0 {
		uri = fmt.Sprintf("%s://%s:%d", scheme, ipv4Address, port)
	} else {
		switch scheme {
		case models.UriScheme_HTTP:
			uri = fmt.Sprintf("%s://%s:80", scheme, ipv4Address)
		case models.UriScheme_HTTPS:
			uri = fmt.Sprintf("%s://%s:443", scheme, ipv4Address)
		}
	}
	return
}

func (c *WEBUIContext) GetTokenCtx(serviceName models.ServiceName, targetNF models.NrfNfManagementNfType) (
	context.Context, *models.ProblemDetails, error,
) {
	if !c.OAuth2Required {
		return context.TODO(), nil, nil
	}

	logger.ConsumerLog.Infoln("GetTokenCtx:", targetNF, serviceName)
	return oauth.GetTokenCtx(models.NrfNfManagementNfType_AF, targetNF,
		c.NfInstanceID, c.NrfUri, string(serviceName))
}

// NewRequestWithContext() will not apply header in ctx
// so httpsClient.Do(req) will not have token in header if OAuth2 enable
func (c *WEBUIContext) RequestBindToken(req *http.Request, ctx context.Context) error {
	if tok, ok := ctx.Value(openapi.ContextOAuth2).(oauth2.TokenSource); ok {
		// We were able to grab an oauth2 token from the context
		var latestToken *oauth2.Token
		var err error
		if latestToken, err = tok.Token(); err != nil {
			logger.ConsumerLog.Error(err)
			return err
		}
		latestToken.SetAuthHeader(req)
	}
	return nil
}
