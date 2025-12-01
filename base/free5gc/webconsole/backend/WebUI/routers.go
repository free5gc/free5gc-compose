package WebUI

import (
	"net/http"

	"github.com/gin-gonic/gin"

	logger_util "github.com/free5gc/util/logger"
	"github.com/free5gc/webconsole/backend/logger"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// Routes is the list of the generated Route.
type Routes []Route

// NewRouter returns a new router.
func NewRouter() *gin.Engine {
	router := logger_util.NewGinWithLogrus(logger.GinLog)
	AddService(router)
	return router
}

func AddService(engine *gin.Engine) *gin.RouterGroup {
	group := engine.Group("/api")

	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			group.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			group.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			group.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			group.DELETE(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			group.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodOptions:
			group.OPTIONS(route.Pattern, route.HandlerFunc)
		}
	}

	return group
}

var routes = Routes{
	{
		"GetExample",
		http.MethodGet,
		"/sample",
		GetSampleJSON,
	},

	{
		"Login",
		http.MethodPost,
		"/login",
		Login,
	},

	{
		"Logout",
		http.MethodPost,
		"/logout",
		Logout,
	},

	{
		"GetTenants",
		http.MethodGet,
		"/tenant",
		GetTenants,
	},

	{
		"GetTenantByID",
		http.MethodGet,
		"/tenant/:tenantId",
		GetTenantByID,
	},

	{
		"PostTenant",
		http.MethodPost,
		"/tenant",
		PostTenant,
	},

	{
		"PutTenantByID",
		http.MethodPut,
		"/tenant/:tenantId",
		PutTenantByID,
	},

	{
		"DeleteTenantByID",
		http.MethodDelete,
		"/tenant/:tenantId",
		DeleteTenantByID,
	},

	{
		"GetUsers",
		http.MethodGet,
		"/tenant/:tenantId/user",
		GetUsers,
	},

	{
		"GetUserByID",
		http.MethodGet,
		"/tenant/:tenantId/user/:userId",
		GetUserByID,
	},

	{
		"PostUserByID",
		http.MethodPost,
		"/tenant/:tenantId/user",
		PostUserByID,
	},

	{
		"PutUserByID",
		http.MethodPut,
		"/tenant/:tenantId/user/:userId",
		PutUserByID,
	},

	{
		"DeleteUserByID",
		http.MethodDelete,
		"/tenant/:tenantId/user/:userId",
		DeleteUserByID,
	},

	{
		"GetSubscribers",
		http.MethodGet,
		"/subscriber",
		GetSubscribers,
	},

	{
		"OptionsSubscribers",
		http.MethodOptions,
		"/subscriber",
		OptionsSubscribers,
	},

	{
		"GetSubscriberByID",
		http.MethodGet,
		"/subscriber/:ueId/:servingPlmnId",
		GetSubscriberByID,
	},

	{
		"PostSubscriberByID",
		http.MethodPost,
		"/subscriber/:ueId/:servingPlmnId",
		PostSubscriberByID,
	},

	{
		"PostMultiSubscriber",
		http.MethodPost,
		"/subscriber/:ueId/:servingPlmnId/:userNumber",
		PostSubscriberByID,
	},

	{
		"PutSubscriberByID",
		http.MethodPut,
		"/subscriber/:ueId/:servingPlmnId",
		PutSubscriberByID,
	},

	{
		"DeleteSubscriberByID",
		http.MethodDelete,
		"/subscriber/:ueId/:servingPlmnId",
		DeleteSubscriberByID,
	},

	{
		"DeleteMultipleSubscribers",
		http.MethodDelete,
		"/subscriber",
		DeleteMultipleSubscribers,
	},

	{
		"PatchSubscriberByID",
		http.MethodPatch,
		"/subscriber/:ueId/:servingPlmnId",
		PatchSubscriberByID,
	},

	{
		"Registered UE Context",
		http.MethodGet,
		"/registered-ue-context",
		GetRegisteredUEContext,
	},

	{
		"Individual Registered UE Context",
		http.MethodGet,
		"/registered-ue-context/:supi",
		GetRegisteredUEContext,
	},

	{
		"UE PDU Session Info",
		http.MethodGet,
		"/ue-pdu-session-info/:smContextRef",
		GetUEPDUSessionInfo,
	},

	{
		"Change Password",
		http.MethodPost,
		"/change-password",
		ChangePasswordInfo,
	},

	{
		"Charging Record",
		http.MethodGet,
		"/charging-record",
		GetChargingRecord,
	},

	{
		"Charging Data",
		http.MethodGet,
		"/charging-data/:chargingMethod",
		GetChargingData,
	},

	{
		"Verify StaticIP",
		http.MethodPost,
		"/verify-staticip",
		VerifyStaticIP,
	},

	{
		"Delete Profile",
		http.MethodDelete,
		"/profile/:profileName",
		DeleteProfile,
	},

	{
		"Delete Multiple Profiles",
		http.MethodDelete,
		"/profile",
		DeleteMultipleProfiles,
	},

	{
		"Get Profile List",
		http.MethodGet,
		"/profile",
		GetProfiles,
	},

	{
		"Get Profile",
		http.MethodGet,
		"/profile/:profileName",
		GetProfile,
	},

	{
		"Post Profile",
		http.MethodPost,
		"/profile",
		PostProfile,
	},

	{
		"Put Profile",
		http.MethodPut,
		"/profile/:profileName",
		PutProfile,
	},
}
