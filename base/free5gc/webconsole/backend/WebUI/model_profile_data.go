package WebUI

import "github.com/free5gc/openapi/models"

type Profile struct {
	ProfileName                       string                                     `json:"profileName"`
	TenantId                          string                                     `json:"tenantId"`
	AccessAndMobilitySubscriptionData models.AccessAndMobilitySubscriptionData   `json:"AccessAndMobilitySubscriptionData"`
	SessionManagementSubscriptionData []models.SessionManagementSubscriptionData `json:"SessionManagementSubscriptionData"`
	SmfSelectionSubscriptionData      models.SmfSelectionSubscriptionData        `json:"SmfSelectionSubscriptionData"`
	AmPolicyData                      models.AmPolicyData                        `json:"AmPolicyData"`
	SmPolicyData                      models.SmPolicyData                        `json:"SmPolicyData"`
	FlowRules                         []FlowRule                                 `json:"FlowRules"`
	QosFlows                          []QosFlow                                  `json:"QosFlows"`
	ChargingDatas                     []ChargingData
}
