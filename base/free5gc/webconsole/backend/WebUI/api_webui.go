package WebUI

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/mongoapi"
	"github.com/free5gc/webconsole/backend/logger"
	"github.com/free5gc/webconsole/backend/webui_context"
)

const (
	authSubsDataColl    = "subscriptionData.authenticationData.authenticationSubscription"
	authWebSubsDataColl = "subscriptionData.authenticationData.webAuthenticationSubscription"
	amDataColl          = "subscriptionData.provisionedData.amData"
	smDataColl          = "subscriptionData.provisionedData.smData"
	smfSelDataColl      = "subscriptionData.provisionedData.smfSelectionSubscriptionData"
	amPolicyDataColl    = "policyData.ues.amData"
	smPolicyDataColl    = "policyData.ues.smData"
	flowRuleDataColl    = "policyData.ues.flowRule"
	qosFlowDataColl     = "policyData.ues.qosFlow"
	chargingDataColl    = "policyData.ues.chargingData"
	userDataColl        = "userData"
	tenantDataColl      = "tenantData"
	identityDataColl    = "subscriptionData.identityData"
	profileDataColl     = "profileData" // store profile data
)

// Security and configuration constants
const (
	DefaultBcryptCost  = 12
	JWTKeySize         = 256
	JWTExpirationHours = 24
)

var jwtKey = "" // for generating JWT

var httpsClient *http.Client

func init() {
	httpsClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// Create Admin's Tenant & Account
func SetAdmin() {
	err := mongoapi.RestfulAPIDeleteOne("tenantData", bson.M{"tenantName": "admin"})
	if err != nil {
		logger.InitLog.Errorf("RestfulAPIDeleteOne err: %+v", err)
	}
	err = mongoapi.RestfulAPIDeleteOne("userData", bson.M{"email": "admin"})
	if err != nil {
		logger.InitLog.Errorf("RestfulAPIDeleteOne err: %+v", err)
	}

	// Create Admin tenant
	logger.InitLog.Infoln("Create tenant: admin")

	adminTenantData := bson.M{
		"tenantId":   uuid.Must(uuid.NewRandom()).String(),
		"tenantName": "admin",
	}

	_, err = mongoapi.RestfulAPIPutOne("tenantData", bson.M{"tenantName": "admin"}, adminTenantData)
	if err != nil {
		logger.InitLog.Errorf("RestfulAPIPutOne err: %+v", err)
	}

	AmdinTenant, err := mongoapi.RestfulAPIGetOne("tenantData", bson.M{"tenantName": "admin"})
	if err != nil {
		logger.InitLog.Errorf("RestfulAPIGetOne err: %+v", err)
	}

	// Create Admin user
	logger.InitLog.Infoln("Create user: admin")

	hash, err := bcrypt.GenerateFromPassword([]byte("free5gc"), DefaultBcryptCost)
	if err != nil {
		logger.InitLog.Errorf("GenerateFromPassword err: %+v", err)
	}

	adminUserData := bson.M{
		"userId":            uuid.Must(uuid.NewRandom()).String(),
		"tenantId":          AmdinTenant["tenantId"],
		"email":             "admin",
		"encryptedPassword": string(hash),
	}

	_, err = mongoapi.RestfulAPIPutOne("userData", bson.M{"email": "admin"}, adminUserData)
	if err != nil {
		logger.InitLog.Errorf("RestfulAPIPutOne err: %+v", err)
	}
}

func InitJwtKey() error {
	randomBytes := make([]byte, JWTKeySize)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return errors.Wrap(err, "Init JWT key error")
	} else {
		jwtKey = string(randomBytes)
	}
	return nil
}

func mapToByte(data map[string]interface{}) (ret []byte) {
	ret, err := json.Marshal(data)
	if err != nil {
		logger.ProcLog.Errorf("mapToByte err: %+v", err)
	}
	return
}

func sliceToByte(data []map[string]interface{}) (ret []byte) {
	ret, err := json.Marshal(data)
	if err != nil {
		logger.ProcLog.Errorf("sliceToByte err: %+v", err)
	}
	return
}

func toBsonM(data interface{}) (ret bson.M) {
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.ProcLog.Errorf("toBsonM err: %+v", err)
	}
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.ProcLog.Errorf("toBsonM err: %+v", err)
	}
	return
}

func toBsonA(data interface{}) (ret bson.A) {
	tmp, err := json.Marshal(data)
	if err != nil {
		logger.ProcLog.Errorf("toBsonM err: %+v", err)
	}
	err = json.Unmarshal(tmp, &ret)
	if err != nil {
		logger.ProcLog.Errorf("toBsonM err: %+v", err)
	}
	return
}

func EscapeDnn(dnn string) string {
	return strings.ReplaceAll(dnn, ".", "_")
}

func UnescapeDnn(dnnKey string) string {
	return strings.ReplaceAll(dnnKey, "_", ".")
}

func setCorsHeader(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set(
		"Access-Control-Allow-Headers",
		"Content-Type, Content-Length, Accept-Encoding, "+
			"X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Token",
	)
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
}

func getMsisdn(gpsis interface{}) string { // select first msisdn from gpsis
	msisdn := ""
	gpsisReflected := reflect.ValueOf(gpsis) // use reflect to range over interface{}
	for i := 0; i < gpsisReflected.Len(); i++ {
		gpsi := gpsisReflected.Index(i).Interface().(string) // transform type reflect.value to string
		if strings.HasPrefix(gpsi, "msisdn-") {              // check if gpsi contain prefix "msisdn-"
			msisdn = gpsi
		}
	}

	return msisdn
}

func gpsiToSupi(ueId string) string { // input ueId could be: IMSI, GPSI
	// If ueId is IMSI, immediately return IMSI without modify.
	// If ueId prefix is "msisdn-", it means that ueId now is GPSI (MSISDN), and we have to map to SUPI.
	if strings.HasPrefix(ueId, "msisdn-") {
		filter := bson.M{"gpsi": ueId}
		dbResult, err := mongoapi.RestfulAPIGetOne(identityDataColl, filter)
		if err != nil {
			logger.ProcLog.Errorf("GetSupibyMsisdn err: %+v", err)
		}
		if dbResult != nil {
			ueId = dbResult["ueId"].(string)
		} else {
			// db cannot find a supi mapped to msisdn, return null string for error detection
			logger.ProcLog.Error("msisdn not found")
			return ""
		}
	}
	return ueId // This must match SUPI format
}

func sendResponseToClient(c *gin.Context, response *http.Response) {
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		logger.ProcLog.Errorf("sendResponseToClient err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	c.JSON(response.StatusCode, jsonData)
}

func sendResponseToClientFilterTenant(c *gin.Context, response *http.Response, tenantId string) {
	// Subscription data.
	filterTenantIdOnly := bson.M{"tenantId": tenantId}
	amDataList, err := mongoapi.RestfulAPIGetMany(amDataColl, filterTenantIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("sendResponseToClientFilterTenant err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	tenantCheck := func(supi string) bool {
		for _, amData := range amDataList {
			if supi == amData["ueId"] {
				return true
			}
		}
		return false
	}

	// Response data.
	var jsonData interface{}
	err = json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		logger.ProcLog.Errorf("sendResponseToClientFilterTenant err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	s := reflect.ValueOf(jsonData)
	if s.Kind() != reflect.Slice {
		c.JSON(response.StatusCode, jsonData)
		return
	}

	var sliceData []interface{}
	for i := 0; i < s.Len(); i++ {
		mapData := s.Index(i).Interface()
		m := reflect.ValueOf(mapData)
		for _, key := range m.MapKeys() {
			if key.String() == "Supi" {
				strct := m.MapIndex(key)
				if tenantCheck(strct.Interface().(string)) {
					sliceData = append(sliceData, mapData)
				}
			}
		}
	}

	c.JSON(response.StatusCode, sliceData)
}

type OAuth struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func JWT(email, userId, tenantId string) string {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userId
	claims["iat"] = time.Now()
	claims["exp"] = time.Now().Add(time.Hour * JWTExpirationHours).Unix()
	claims["email"] = email
	claims["tenantId"] = tenantId

	if jwtKey == "" {
		return ""
	}

	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		logger.ProcLog.Errorf("JWT err: %+v", err)
		return ""
	}

	return tokenString
}

func Login(c *gin.Context) {
	setCorsHeader(c)

	login := LoginRequest{}
	err := json.NewDecoder(c.Request.Body).Decode(&login)
	if err != nil {
		logger.ProcLog.Warnln("JSON decode error", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	filterEmail := bson.M{"email": login.Username}
	userData, err := mongoapi.RestfulAPIGetOne(userDataColl, filterEmail)
	if err != nil {
		logger.ProcLog.Errorf("Login err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if len(userData) == 0 {
		logger.ProcLog.Warnln("Can't find user email", login.Username)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	hash := userData["encryptedPassword"].(string)

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(login.Password))
	if err != nil {
		logger.ProcLog.Warnln("Password incorrect", login.Username)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	userId := userData["userId"].(string)
	tenantId := userData["tenantId"].(string)

	logger.ProcLog.Infoln("Login success {",
		"username:", login.Username,
		", userid:", userId,
		", tenantid:", tenantId,
		"}")

	token := JWT(login.Username, userId, tenantId)
	if token == "" {
		logger.ProcLog.Errorln("token is empty")
	}

	oauth := OAuth{}
	oauth.AccessToken = token
	c.JSON(http.StatusOK, oauth)
}

// Placeholder to handle logout.
func Logout(c *gin.Context) {
	setCorsHeader(c)
	// Needs to invalidate access_token.
	c.JSON(http.StatusOK, gin.H{})
}

type AuthSub struct {
	WebAuthenticationSubscription
	TenantId string `json:"tenantId" bson:"tenantId"`
}

// Parse JWT
func ParseJWT(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Tolerance for User's Time quicker than Server's time
		mapClaims := token.Claims.(jwt.MapClaims)
		delete(mapClaims, "iat")
		return []byte(jwtKey), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "ParseJWT error")
	}

	claims, _ := token.Claims.(jwt.MapClaims)

	return claims, nil
}

// Check of admin user. This should be done with proper JWT token.
func CheckAuth(c *gin.Context) bool {
	tokenStr := c.GetHeader("Token")
	claims, err := ParseJWT(tokenStr)
	if err != nil {
		return false
	}
	if claims["email"].(string) == "admin" {
		return true
	} else {
		return false
	}
}

// Tenant ID
func GetTenantId(c *gin.Context) (string, error) {
	tokenStr := c.GetHeader("Token")
	if tokenStr == "" {
		return "", fmt.Errorf("no token in header")
	}
	claims, err := ParseJWT(tokenStr)
	if err != nil {
		return "", errors.Wrap(err, "GetTenantId error")
	}
	return claims["tenantId"].(string), nil
}

// Tenant
func GetTenants(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantDataInterface, err := mongoapi.RestfulAPIGetMany(tenantDataColl, bson.M{})
	if err != nil {
		logger.ProcLog.Errorf("GetTenants err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var tenantData []Tenant
	err = json.Unmarshal(sliceToByte(tenantDataInterface), &tenantData)
	if err != nil {
		logger.ProcLog.Errorf("GetTenants err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	c.JSON(http.StatusOK, tenantData)
}

func GetTenantByID(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantId := c.Param("tenantId")

	filterTenantIdOnly := bson.M{"tenantId": tenantId}
	tenantDataInterface, err := mongoapi.RestfulAPIGetOne(tenantDataColl, filterTenantIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("GetTenantByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if len(tenantDataInterface) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}

	var tenantData Tenant
	err = json.Unmarshal(mapToByte(tenantDataInterface), &tenantData)
	if err != nil {
		logger.ProcLog.Errorf("GetTenantByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	c.JSON(http.StatusOK, tenantData)
}

func PostTenant(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	var tenantData Tenant
	if err := c.ShouldBindJSON(&tenantData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	// fmt.Println("in Post Tenant")
	if tenantData.TenantId == "" {
		tenantData.TenantId = uuid.Must(uuid.NewRandom()).String()
	}

	tenantBsonM := toBsonM(tenantData)
	filterTenantIdOnly := bson.M{"tenantId": tenantData.TenantId}
	if _, err := mongoapi.RestfulAPIPost(tenantDataColl, filterTenantIdOnly, tenantBsonM); err != nil {
		logger.ProcLog.Errorf("PostTenant err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, tenantData)
}

func PutTenantByID(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantId := c.Param("tenantId")

	filterTenantIdOnly := bson.M{"tenantId": tenantId}
	tenantDataInterface, err := mongoapi.RestfulAPIGetOne(tenantDataColl, filterTenantIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("PutTenantByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if len(tenantDataInterface) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}

	var tenantData Tenant
	if err = c.ShouldBindJSON(&tenantData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	tenantData.TenantId = tenantId

	tenantBsonM := toBsonM(tenantData)
	filterTenantIdOnly = bson.M{"tenantId": tenantId}
	if _, err_post := mongoapi.RestfulAPIPost(tenantDataColl, filterTenantIdOnly, tenantBsonM); err_post != nil {
		logger.ProcLog.Errorf("PutTenantByID err: %+v", err_post)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func DeleteTenantByID(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantId := c.Param("tenantId")
	filterTenantIdOnly := bson.M{"tenantId": tenantId}

	if err := mongoapi.RestfulAPIDeleteMany(amDataColl, filterTenantIdOnly); err != nil {
		logger.ProcLog.Errorf("DeleteTenantByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if err := mongoapi.RestfulAPIDeleteMany(userDataColl, filterTenantIdOnly); err != nil {
		logger.ProcLog.Errorf("DeleteTenantByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if err := mongoapi.RestfulAPIDeleteOne(tenantDataColl, filterTenantIdOnly); err != nil {
		logger.ProcLog.Errorf("DeleteTenantByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// Utility function.
func GetTenantById(tenantId string) map[string]interface{} {
	filterTenantIdOnly := bson.M{"tenantId": tenantId}
	tenantData, err := mongoapi.RestfulAPIGetOne(tenantDataColl, filterTenantIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("GetTenantById err: %+v", err)
		return nil
	}
	return tenantData
}

// Users
func GetUsers(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantId := c.Param("tenantId")
	if len(GetTenantById(tenantId)) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}

	filterTenantIdOnly := bson.M{"tenantId": tenantId}
	userDataInterface, err := mongoapi.RestfulAPIGetMany(userDataColl, filterTenantIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("GetUsers err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	var userData []User
	err = json.Unmarshal(sliceToByte(userDataInterface), &userData)
	if err != nil {
		logger.ProcLog.Errorf("GetUsers err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	for pos := range userData {
		userData[pos].EncryptedPassword = ""
	}

	c.JSON(http.StatusOK, userData)
}

func GetUserByID(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantId := c.Param("tenantId")
	if len(GetTenantById(tenantId)) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}
	userId := c.Param("userId")

	filterUserIdOnly := bson.M{"tenantId": tenantId, "userId": userId}
	userDataInterface, err := mongoapi.RestfulAPIGetOne(userDataColl, filterUserIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("GetUserByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if len(userDataInterface) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}

	var userData User
	err = json.Unmarshal(mapToByte(userDataInterface), &userData)
	if err != nil {
		logger.ProcLog.Errorf("GetUserByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	userData.EncryptedPassword = ""

	c.JSON(http.StatusOK, userData)
}

func PostUserByID(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantId := c.Param("tenantId")
	if len(GetTenantById(tenantId)) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}

	var userData User
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	filterEmail := bson.M{"email": userData.Email}
	userWithEmailData, err := mongoapi.RestfulAPIGetOne(userDataColl, filterEmail)
	if err != nil {
		logger.ProcLog.Errorf("PostUserByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if len(userWithEmailData) != 0 {
		logger.ProcLog.Warnln("Email already exists", userData.Email)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	userData.TenantId = tenantId
	userData.UserId = uuid.Must(uuid.NewRandom()).String()
	hash, err := bcrypt.GenerateFromPassword([]byte(userData.EncryptedPassword), DefaultBcryptCost)
	if err != nil {
		logger.ProcLog.Errorf("PostUserByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	userData.EncryptedPassword = string(hash)

	userBsonM := toBsonM(userData)
	filterUserIdOnly := bson.M{"tenantId": userData.TenantId, "userId": userData.UserId}
	if _, err_post := mongoapi.RestfulAPIPost(userDataColl, filterUserIdOnly, userBsonM); err_post != nil {
		logger.ProcLog.Errorf("PostUserByID err: %+v", err_post)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, userData)
}

func PutUserByID(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantId := c.Param("tenantId")
	if len(GetTenantById(tenantId)) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}
	userId := c.Param("userId")

	var newUserData User
	if err := c.ShouldBindJSON(&newUserData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	filterUserIdOnly := bson.M{"tenantId": tenantId, "userId": userId}
	userDataInterface, err := mongoapi.RestfulAPIGetOne(userDataColl, filterUserIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("PutUserByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if len(userDataInterface) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}

	var userData User
	err = json.Unmarshal(mapToByte(userDataInterface), &userData)
	if err != nil {
		logger.ProcLog.Errorf("PutUserByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if newUserData.Email != "" && newUserData.Email != userData.Email {
		filterEmail := bson.M{"email": newUserData.Email}
		sameEmailInterface, err_get := mongoapi.RestfulAPIGetOne(userDataColl, filterEmail)
		if err_get != nil {
			logger.ProcLog.Errorf("PutUserByID err: %+v", err_get)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		if len(sameEmailInterface) != 0 {
			c.JSON(http.StatusBadRequest, bson.M{})
			return
		}
		userData.Email = newUserData.Email
	}

	if newUserData.EncryptedPassword != "" {
		hash, err_gen := bcrypt.GenerateFromPassword([]byte(newUserData.EncryptedPassword), DefaultBcryptCost)
		if err_gen != nil {
			logger.ProcLog.Errorf("PutUserByID err: %+v", err_gen)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		userData.EncryptedPassword = string(hash)
	}

	userBsonM := toBsonM(userData)
	if _, err_post := mongoapi.RestfulAPIPost(userDataColl, filterUserIdOnly, userBsonM); err_post != nil {
		logger.ProcLog.Errorf("PutUserByID err: %+v", err_post)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, userData)
}

func DeleteUserByID(c *gin.Context) {
	setCorsHeader(c)

	if !CheckAuth(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"cause": "Illegal Token"})
		return
	}

	tenantId := c.Param("tenantId")
	if len(GetTenantById(tenantId)) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}
	userId := c.Param("userId")

	filterUserIdOnly := bson.M{"tenantId": tenantId, "userId": userId}
	if err := mongoapi.RestfulAPIDeleteOne(userDataColl, filterUserIdOnly); err != nil {
		logger.ProcLog.Errorf("DeleteUserByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// Get all subscribers list
func GetSubscribers(c *gin.Context) {
	setCorsHeader(c)

	logger.ProcLog.Infoln("Get All Subscribers List")

	userTenantId, err := GetTenantId(c)
	if err != nil {
		logger.ProcLog.Errorln(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"cause": "Illegal Token (Relogin required)!",
		})
		return
	}

	isAdmin := CheckAuth(c)

	subsList := make([]SubsListIE, 0)
	amDataList, err := mongoapi.RestfulAPIGetMany(amDataColl, bson.M{})
	if err != nil {
		logger.ProcLog.Errorf("GetSubscribers err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	for _, amData := range amDataList {
		ueId := amData["ueId"]
		servingPlmnId := amData["servingPlmnId"]
		gpsi := getMsisdn(amData["gpsis"])
		tenantId := amData["tenantId"]

		filterUeIdOnly := bson.M{"ueId": ueId}
		authSubsDataInterface, err_get := mongoapi.RestfulAPIGetOne(authWebSubsDataColl, filterUeIdOnly)
		if err_get != nil {
			logger.ProcLog.Errorf("GetSubscribers err: %+v", err_get)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		var authSubsData AuthSub
		if err = json.Unmarshal(mapToByte(authSubsDataInterface), &authSubsData); err != nil {
			logger.ProcLog.Errorf("GetSubscribers err: %+v", err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		if isAdmin || userTenantId == tenantId {
			tmp := SubsListIE{
				PlmnID: servingPlmnId.(string),
				UeId:   ueId.(string),
				Gpsi:   gpsi,
			}
			subsList = append(subsList, tmp)
		}
	}

	c.JSON(http.StatusOK, subsList)
}

// Get subscriber by IMSI(ueId) and PlmnID(servingPlmnId)
func GetSubscriberByID(c *gin.Context) {
	setCorsHeader(c)

	logger.ProcLog.Infoln("Get One Subscriber Data")

	var subsData SubsData

	ueId := c.Param("ueId")
	ueId = gpsiToSupi(ueId)
	servingPlmnId := c.Param("servingPlmnId")
	// checking whether gpsi is successfully transformed to supi or not
	if ueId == "" {
		logger.ProcLog.Errorf("GetSubscriberByID err: gpsi does not exists")
		c.JSON(http.StatusNotFound, gin.H{
			"cause": "gpsi does not exists",
		})
		return
	}
	filterUeIdOnly := bson.M{"ueId": ueId}
	filter := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}

	authSubsDataInterface, err := mongoapi.RestfulAPIGetOne(authWebSubsDataColl, filterUeIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	amDataDataInterface, err := mongoapi.RestfulAPIGetOne(amDataColl, filter)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	smDataDataInterface, err := mongoapi.RestfulAPIGetMany(smDataColl, filter)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	smfSelDataInterface, err := mongoapi.RestfulAPIGetOne(smfSelDataColl, filter)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	amPolicyDataInterface, err := mongoapi.RestfulAPIGetOne(amPolicyDataColl, filterUeIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	smPolicyDataInterface, err := mongoapi.RestfulAPIGetOne(smPolicyDataColl, filterUeIdOnly)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	flowRuleDataInterface, err := mongoapi.RestfulAPIGetMany(flowRuleDataColl, filter)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	qosFlowInterface, err := mongoapi.RestfulAPIGetMany(qosFlowDataColl, filter)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	chargingDatasInterface, err := mongoapi.RestfulAPIGetMany(chargingDataColl, filter)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	var authSubsData WebAuthenticationSubscription
	if err = json.Unmarshal(mapToByte(authSubsDataInterface), &authSubsData); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var amDataData models.AccessAndMobilitySubscriptionData
	if err = json.Unmarshal(mapToByte(amDataDataInterface), &amDataData); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var smDataData []models.SessionManagementSubscriptionData
	if err = json.Unmarshal(sliceToByte(smDataDataInterface), &smDataData); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var smfSelData models.SmfSelectionSubscriptionData
	if err = json.Unmarshal(mapToByte(smfSelDataInterface), &smfSelData); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var amPolicyData models.AmPolicyData
	if err = json.Unmarshal(mapToByte(amPolicyDataInterface), &amPolicyData); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var smPolicyData models.SmPolicyData
	if err = json.Unmarshal(mapToByte(smPolicyDataInterface), &smPolicyData); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var flowRules []FlowRule
	if err = json.Unmarshal(sliceToByte(flowRuleDataInterface), &flowRules); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var qosFlows []QosFlow
	if err = json.Unmarshal(sliceToByte(qosFlowInterface), &qosFlows); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	var chargingDatas []ChargingData
	if err = json.Unmarshal(sliceToByte(chargingDatasInterface), &chargingDatas); err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if flowRules == nil {
		flowRules = make([]FlowRule, 0)
	}
	if qosFlows == nil {
		qosFlows = make([]QosFlow, 0)
	}
	if chargingDatas == nil {
		chargingDatas = make([]ChargingData, 0)
	}

	for key, SnssaiData := range smPolicyData.SmPolicySnssaiData {
		tmpSmPolicyDnnData := make(map[string]models.SmPolicyDnnData)
		for escapedDnn, dnn := range SnssaiData.SmPolicyDnnData {
			dnnKey := UnescapeDnn(escapedDnn)
			tmpSmPolicyDnnData[dnnKey] = dnn
		}
		SnssaiData.SmPolicyDnnData = tmpSmPolicyDnnData
		smPolicyData.SmPolicySnssaiData[key] = SnssaiData
	}

	subsData = SubsData{
		PlmnID:                            servingPlmnId,
		UeId:                              ueId,
		WebAuthenticationSubscription:     authSubsData,
		AccessAndMobilitySubscriptionData: amDataData,
		SessionManagementSubscriptionData: smDataData,
		SmfSelectionSubscriptionData:      smfSelData,
		AmPolicyData:                      amPolicyData,
		SmPolicyData:                      smPolicyData,
		FlowRules:                         flowRules,
		QosFlows:                          qosFlows,
		ChargingDatas:                     chargingDatas,
	}

	c.JSON(http.StatusOK, subsData)
}

// Post subscriber by IMSI(ueId) and PlmnID(servingPlmnId)
// PostSubscriberByID godoc
// @Summary     CreateSubscriberByID
// @Description Create subscriber by IMSI(ueId) and PlmnID(servingPlmnId)
// @Accept       json
// @Produce      json
// @Param ueId path string true "imsi"
// @Param servingPlmnId path string true "servingPlmnId"
// @Param subdata body SubsData true "sub data"
// @Success      201 "Create subscription success"
// @Failure 400 {object} HTTPError "JSON format incorrect"
// @Router  /subscriber/{ueId}/{servingPlmnId}/{userNumber} [post]
func PostSubscriberByID(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Post One Subscriber Data")

	tokenStr := c.GetHeader("Token")
	claims, err := ParseJWT(tokenStr)
	if err != nil {
		logger.ProcLog.Errorln(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"cause": "Illegal Token (Relogin required)!",
		})
		return
	}

	var subsData SubsData
	err = c.ShouldBindJSON(&subsData)
	if err != nil {
		logger.ProcLog.Errorf("PostSubscriberByID err: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"cause": "JSON format incorrect",
		})
		return
	}
	ueId := strings.Split(c.Param("ueId"), "-")[1]
	servingPlmnId := c.Param("servingPlmnId")
	userNumber := c.Param("userNumber")
	if userNumber == "" {
		userNumber = "1"
	}
	userNumberTemp, err := strconv.Atoi(userNumber)
	if err != nil {
		logger.ProcLog.Errorf("PostSubscriberByID err: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"cause": "userNumber format incorrect",
		})
		return
	}
	gpsi := getMsisdn(toBsonM(subsData.AccessAndMobilitySubscriptionData)["gpsis"])
	gpsiInt := 0
	if gpsi != "" {
		if len(gpsi) > 7 && gpsi[:7] == "msisdn-" { // case: MSISDN
			gpsiInt, err = strconv.Atoi(gpsi[7:])
		}

		if err != nil {
			logger.ProcLog.Errorf("PostSubscriberByID err: %+v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"cause": "gpsi format incorrect (now only support MSISDN format)",
			})
			return
		}
	}

	ueIdTemp, err := strconv.Atoi(ueId)
	if err != nil {
		logger.ProcLog.Errorf("PostSubscriberByID err: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"cause": "ueId format incorrect",
		})
		return
	}

	for i := 0; i < userNumberTemp; i++ {
		ueId = fmt.Sprintf("imsi-%015d", ueIdTemp)
		if gpsiInt != 0 {
			if !validate(ueId, gpsi) {
				logger.ProcLog.Errorf("duplicate gpsi: %v", gpsi)
				c.JSON(http.StatusBadRequest, gin.H{
					"cause": "duplicate gpsi",
				})
				return
			}
			gpsiInt += 1
		}
		ueIdTemp += 1

		subsData.AccessAndMobilitySubscriptionData.Gpsis[0] = gpsi
		// create a msisdn-supi map
		logger.ProcLog.Infof("PostSubscriberByID gpsi: %+v", subsData.AccessAndMobilitySubscriptionData.Gpsis[0])
		identityDataOperation(ueId, gpsi, "post")
		filterUeIdOnly := bson.M{"ueId": ueId}

		// Lookup same UE ID of other tenant's subscription.
		if claims != nil {
			authSubsDataInterface, err_get := mongoapi.RestfulAPIGetOne(authWebSubsDataColl, filterUeIdOnly)
			if err_get != nil {
				logger.ProcLog.Errorf("PostSubscriberByID err: %+v", err_get)
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}
			if len(authSubsDataInterface) > 0 {
				c.JSON(http.StatusConflict, gin.H{"cause": ueId + " already exists"})
				return
			}
		}
		dbOperation(ueId, servingPlmnId, "post", &subsData, nil, claims, false)
	}
	c.JSON(http.StatusCreated, gin.H{})
}

// ctfang

func validate(supi string, gpsi string) bool {
	// If GPSI is empty, skip to check it.
	if gpsi == "msisdn-" {
		return true
	}
	filter := bson.M{"gpsi": gpsi}
	identityData, err := mongoapi.RestfulAPIGetOne(identityDataColl, filter)
	if err != nil {
		logger.ProcLog.Errorf("GetSubscriberByID err: %+v", err)
	}
	if identityData != nil && identityData["ueId"] != supi {
		return false
	} else {
		return true
	}
}

func identityDataOperation(supi string, gpsi string, method string) {
	filter := bson.M{"ueId": supi}
	data := bson.M{"ueId": supi, "gpsi": gpsi}

	if method == "put" || method == "post" {
		if gpsi != "" {
			if _, err := mongoapi.RestfulAPIPutOne(identityDataColl, filter, data); err != nil {
				logger.ProcLog.Errorf("PutIdentityData err: %+v", err)
			}
		} else {
			// delete
			if err := mongoapi.RestfulAPIDeleteOne(identityDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteIdentityData err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteOne(identityDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteIdentityData err: %+v", err)
			}
		}
	}
}

func sendRechargeNotification(ueId string, rg int32) {
	logger.ProcLog.Infoln("Send Notification to CHF due to quota change")
	webuiSelf := webui_context.GetSelf()
	webuiSelf.UpdateNfProfiles()

	requestUri := fmt.Sprintf("%s/nchf-convergedcharging/v3/recharging/%s_%d", "http://127.0.0.113:8000", ueId, rg)
	req, err := http.NewRequest(http.MethodPut, requestUri, nil)
	if err != nil {
		logger.ProcLog.Error(err)
	}
	defer func() {
		if err = req.Body.Close(); err != nil {
			logger.ProcLog.Error(err)
		}
	}()
	req.Header.Add("Content-Type", "application/json")

	resp, err1 := http.DefaultClient.Do(req)
	if err != nil {
		logger.ProcLog.Errorf("Send Charging Notification err: %+v", err1)
	}
	if err = resp.Body.Close(); err != nil {
		logger.ProcLog.Error(err)
	}
}

func dbOperation(
	ueId string,
	servingPlmnId string,
	method string,
	subsData *SubsData,
	subsDatas []*SubsListIE,
	claims jwt.MapClaims,
	multiple bool,
) {
	filterUeIdOnly := bson.M{"ueId": ueId}
	filter := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId}

	// Replace all data with new one
	switch method {
	case "put":
		if err := mongoapi.RestfulAPIDeleteMany(flowRuleDataColl, filter); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
		if err := mongoapi.RestfulAPIDeleteMany(qosFlowDataColl, filter); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
		if err := mongoapi.RestfulAPIDeleteMany(chargingDataColl, filter); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
	case "delete":
		if multiple {
			multipleFilterUeIdOnlyConditions, multipleFliterConditions := []bson.M{}, []bson.M{}
			for _, subsData := range subsDatas {
				multipleFilterUeIdOnlyConditions = append(multipleFilterUeIdOnlyConditions, bson.M{
					"ueId": subsData.UeId,
				})
				multipleFliterConditions = append(multipleFliterConditions, bson.M{
					"ueId":          subsData.UeId,
					"servingPlmnId": subsData.PlmnID,
				})
			}

			multipleFliterUeIdOnly, multipleFliter := bson.M{
				"$or": multipleFilterUeIdOnlyConditions,
			}, bson.M{
				"$or": multipleFliterConditions,
			}

			if err := mongoapi.RestfulAPIDeleteMany(authSubsDataColl, multipleFliterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(authWebSubsDataColl, multipleFliterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(amDataColl, multipleFliter); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(smDataColl, multipleFliter); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(flowRuleDataColl, multipleFliter); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(smfSelDataColl, multipleFliter); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(amPolicyDataColl, multipleFliterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(smPolicyDataColl, multipleFliterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(qosFlowDataColl, multipleFliter); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(chargingDataColl, multipleFliter); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(identityDataColl, multipleFliterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleIdnetityDatas err: %+v", err)
			}
		} else {
			if err := mongoapi.RestfulAPIDeleteOne(authSubsDataColl, filterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteOne(authWebSubsDataColl, filterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteOne(amDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(smDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(flowRuleDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteOne(smfSelDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteOne(amPolicyDataColl, filterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteOne(smPolicyDataColl, filterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(qosFlowDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteMany(chargingDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteSubscriberByID err: %+v", err)
			}
			if err := mongoapi.RestfulAPIDeleteOne(identityDataColl, filterUeIdOnly); err != nil {
				logger.ProcLog.Errorf("DeleteIdnetityData err: %+v", err)
			}
		}
	}
	if method == "post" || method == "put" {
		webAuthSubsBsonM := toBsonM(subsData.WebAuthenticationSubscription)
		webAuthSubsBsonM["ueId"] = ueId

		authSubs, errModel := WebAuthSubToModels(subsData.WebAuthenticationSubscription)
		if errModel != nil {
			logger.ProcLog.Errorf("WebAuthSubToModels err: %+v", errModel)
		}
		authSubsBsonM := toBsonM(authSubs)
		authSubsBsonM["ueId"] = ueId

		if claims != nil {
			webAuthSubsBsonM["tenantId"] = claims["tenantId"].(string)
			authSubsBsonM["tenantId"] = claims["tenantId"].(string)
		}

		amDataBsonM := toBsonM(subsData.AccessAndMobilitySubscriptionData)
		amDataBsonM["ueId"] = ueId
		amDataBsonM["servingPlmnId"] = servingPlmnId
		if claims != nil {
			amDataBsonM["tenantId"] = claims["tenantId"].(string)
		}

		// Replace all data with new one
		if err := mongoapi.RestfulAPIDeleteMany(smDataColl, filter); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
		for _, data := range subsData.SessionManagementSubscriptionData {
			smDataBsonM := toBsonM(data)
			smDataBsonM["ueId"] = ueId
			smDataBsonM["servingPlmnId"] = servingPlmnId
			filterSmData := bson.M{"ueId": ueId, "servingPlmnId": servingPlmnId, "snssai": data.SingleNssai}
			if _, err := mongoapi.RestfulAPIPutOne(smDataColl, filterSmData, smDataBsonM); err != nil {
				logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
			}
		}

		for key, SnssaiData := range subsData.SmPolicyData.SmPolicySnssaiData {
			tmpSmPolicyDnnData := make(map[string]models.SmPolicyDnnData)
			for dnnKey, dnn := range SnssaiData.SmPolicyDnnData {
				escapedDnn := EscapeDnn(dnnKey)
				tmpSmPolicyDnnData[escapedDnn] = dnn
			}
			SnssaiData.SmPolicyDnnData = tmpSmPolicyDnnData
			subsData.SmPolicyData.SmPolicySnssaiData[key] = SnssaiData
		}

		smfSelSubsBsonM := toBsonM(subsData.SmfSelectionSubscriptionData)
		smfSelSubsBsonM["ueId"] = ueId
		smfSelSubsBsonM["servingPlmnId"] = servingPlmnId
		amPolicyDataBsonM := toBsonM(subsData.AmPolicyData)
		amPolicyDataBsonM["ueId"] = ueId
		smPolicyDataBsonM := toBsonM(subsData.SmPolicyData)
		smPolicyDataBsonM["ueId"] = ueId

		if len(subsData.FlowRules) == 0 {
			logger.ProcLog.Infoln("No Flow Rule")
		} else {
			flowRulesBsonA := make([]interface{}, 0, len(subsData.FlowRules))
			for _, flowRule := range subsData.FlowRules {
				flowRuleBsonM := toBsonM(flowRule)
				flowRuleBsonM["ueId"] = ueId
				flowRuleBsonM["servingPlmnId"] = servingPlmnId
				flowRulesBsonA = append(flowRulesBsonA, flowRuleBsonM)
			}
			if err := mongoapi.RestfulAPIPostMany(flowRuleDataColl, filter, flowRulesBsonA); err != nil {
				logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
			}
		}

		if len(subsData.QosFlows) == 0 {
			logger.ProcLog.Infoln("No QoS Flow")
		} else {
			qosFlowBsonA := make([]interface{}, 0, len(subsData.QosFlows))
			for _, qosFlow := range subsData.QosFlows {
				qosFlowBsonM := toBsonM(qosFlow)
				qosFlowBsonM["ueId"] = ueId
				qosFlowBsonM["servingPlmnId"] = servingPlmnId
				qosFlowBsonA = append(qosFlowBsonA, qosFlowBsonM)
			}
			if err := mongoapi.RestfulAPIPostMany(qosFlowDataColl, filter, qosFlowBsonA); err != nil {
				logger.ProcLog.Errorf("PostSubscriberByID err: %+v", err)
			}
		}

		if len(subsData.ChargingDatas) == 0 {
			logger.ProcLog.Infoln("No Charging Data")
		} else {
			for _, chargingData := range subsData.ChargingDatas {
				var previousChargingData ChargingData
				var chargingFilter primitive.M

				chargingDataBsonM := toBsonM(chargingData)
				// Clear quota for offline charging flow
				if chargingData.ChargingMethod == ChargingOffline {
					chargingDataBsonM["quota"] = "0"
				}

				if chargingData.Dnn != "" && chargingData.Filter != "" {
					// Flow-level charging
					chargingFilter = bson.M{
						"ueId": ueId, "servingPlmnId": servingPlmnId,
						"snssai": chargingData.Snssai,
						"dnn":    chargingData.Dnn,
						"qosRef": chargingData.QosRef,
						"filter": chargingData.Filter,
					}
				} else {
					// Slice-level charging
					chargingFilter = bson.M{
						"ueId": ueId, "servingPlmnId": servingPlmnId,
						"snssai": chargingData.Snssai,
						"qosRef": chargingData.QosRef,
						"dnn":    "",
						"filter": "",
					}
					chargingDataBsonM["dnn"] = ""
					chargingDataBsonM["filter"] = ""
				}

				previousChargingDataInterface, err := mongoapi.RestfulAPIGetOne(chargingDataColl, chargingFilter)
				if err != nil {
					logger.ProcLog.Errorf("PostSubscriberByID err: %+v", err)
				}
				err = json.Unmarshal(mapToByte(previousChargingDataInterface), &previousChargingData)
				if err != nil {
					logger.ProcLog.Error(err)
				}

				ratingGroup := previousChargingDataInterface["ratingGroup"]
				if ratingGroup != nil {
					rg := ratingGroup.(int32)
					chargingDataBsonM["ratingGroup"] = rg
					if previousChargingData.Quota != chargingData.Quota {
						sendRechargeNotification(ueId, rg)
					}
				}

				chargingDataBsonM["ueId"] = ueId
				chargingDataBsonM["servingPlmnId"] = servingPlmnId

				if _, err_put := mongoapi.RestfulAPIPutOne(chargingDataColl, chargingFilter, chargingDataBsonM); err != nil {
					logger.ProcLog.Errorf("PostSubscriberByID err: %+v", err_put)
				}
			}
		}

		if _, err := mongoapi.RestfulAPIPutOne(authWebSubsDataColl, filterUeIdOnly, webAuthSubsBsonM); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
		if _, err := mongoapi.RestfulAPIPutOne(authSubsDataColl, filterUeIdOnly, authSubsBsonM); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
		if _, err := mongoapi.RestfulAPIPutOne(amDataColl, filter, amDataBsonM); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
		if _, err := mongoapi.RestfulAPIPutOne(smfSelDataColl, filter, smfSelSubsBsonM); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
		if _, err := mongoapi.RestfulAPIPutOne(amPolicyDataColl, filterUeIdOnly, amPolicyDataBsonM); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
		if _, err := mongoapi.RestfulAPIPutOne(smPolicyDataColl, filterUeIdOnly, smPolicyDataBsonM); err != nil {
			logger.ProcLog.Errorf("PutSubscriberByID err: %+v", err)
		}
	}
}

// Put subscriber by IMSI(ueId) and PlmnID(servingPlmnId)
func PutSubscriberByID(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Put One Subscriber Data")
	var subsData SubsData
	if err := c.ShouldBindJSON(&subsData); err != nil {
		logger.ProcLog.Errorf("PutSubscriberByID err: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"cause": "JSON format incorrect",
		})
		return
	}
	ueId := c.Param("ueId")
	servingPlmnId := c.Param("servingPlmnId")
	// modify a gpsi-supi map
	gpsi := getMsisdn(toBsonM(subsData.AccessAndMobilitySubscriptionData)["gpsis"])
	if !validate(ueId, gpsi) {
		logger.ProcLog.Errorf("duplicate gpsi: %v", gpsi)
		c.JSON(http.StatusBadRequest, gin.H{
			"cause": "duplicate gpsi",
		})
		return
	}

	logger.ProcLog.Infof("PutSubscriberByID gpsi: %+v", gpsi)
	identityDataOperation(ueId, gpsi, "put")

	var claims jwt.MapClaims = nil
	dbOperation(ueId, servingPlmnId, "put", &subsData, nil, claims, false)
	c.Status(http.StatusNoContent)
}

// Patch subscriber by IMSI(ueId) and PlmnID(servingPlmnId)
func PatchSubscriberByID(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Patch One Subscriber Data")

	var subsData SubsData
	if err := c.ShouldBindJSON(&subsData); err != nil {
		logger.ProcLog.Errorf("PatchSubscriberByID err: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"cause": "JSON format incorrect",
		})
		return
	}

	ueId := c.Param("ueId")
	supi := gpsiToSupi(ueId)
	servingPlmnId := c.Param("servingPlmnId")
	// checking whether gpsi is successfully transformed to supi or not
	if ueId == "" {
		logger.ProcLog.Errorf("PatchSubscriberByID err: gpsi does not exists")
		c.JSON(http.StatusNotFound, gin.H{
			"cause": "gpsi does not exists",
		})
		return
	}
	filterUeIdOnly := bson.M{"ueId": supi}
	filter := bson.M{"ueId": supi, "servingPlmnId": servingPlmnId}

	webAuthSubsBsonM := toBsonM(subsData.WebAuthenticationSubscription)
	webAuthSubsBsonM["ueId"] = supi

	authSubs, errModels := WebAuthSubToModels(subsData.WebAuthenticationSubscription)
	if errModels != nil {
		logger.ProcLog.Errorf("WebAuthSubToModels err: %+v", errModels)
	}
	authSubsBsonM := toBsonM(authSubs)
	authSubsBsonM["ueId"] = ueId

	amDataBsonM := toBsonM(subsData.AccessAndMobilitySubscriptionData)
	amDataBsonM["ueId"] = supi
	amDataBsonM["servingPlmnId"] = servingPlmnId

	// Replace all data with new one
	if err := mongoapi.RestfulAPIDeleteMany(smDataColl, filter); err != nil {
		logger.ProcLog.Errorf("PatchSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	for _, data := range subsData.SessionManagementSubscriptionData {
		smDataBsonM := toBsonM(data)
		smDataBsonM["ueId"] = supi
		smDataBsonM["servingPlmnId"] = servingPlmnId
		filterSmData := bson.M{"ueId": supi, "servingPlmnId": servingPlmnId, "snssai": data.SingleNssai}
		if err := mongoapi.RestfulAPIMergePatch(smDataColl, filterSmData, smDataBsonM); err != nil {
			logger.ProcLog.Errorf("PatchSubscriberByID err: %+v", err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
	}

	smfSelSubsBsonM := toBsonM(subsData.SmfSelectionSubscriptionData)
	smfSelSubsBsonM["ueId"] = supi
	smfSelSubsBsonM["servingPlmnId"] = servingPlmnId
	amPolicyDataBsonM := toBsonM(subsData.AmPolicyData)
	amPolicyDataBsonM["ueId"] = supi
	smPolicyDataBsonM := toBsonM(subsData.SmPolicyData)
	smPolicyDataBsonM["ueId"] = supi

	if err := mongoapi.RestfulAPIMergePatch(authWebSubsDataColl, filterUeIdOnly, webAuthSubsBsonM); err != nil {
		logger.ProcLog.Errorf("PatchSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if err := mongoapi.RestfulAPIMergePatch(authSubsDataColl, filterUeIdOnly, authSubsBsonM); err != nil {
		logger.ProcLog.Errorf("PatchSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if err := mongoapi.RestfulAPIMergePatch(amDataColl, filter, amDataBsonM); err != nil {
		logger.ProcLog.Errorf("PatchSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if err := mongoapi.RestfulAPIMergePatch(smfSelDataColl, filter, smfSelSubsBsonM); err != nil {
		logger.ProcLog.Errorf("PatchSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if err := mongoapi.RestfulAPIMergePatch(amPolicyDataColl, filterUeIdOnly, amPolicyDataBsonM); err != nil {
		logger.ProcLog.Errorf("PatchSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if err := mongoapi.RestfulAPIMergePatch(smPolicyDataColl, filterUeIdOnly, smPolicyDataBsonM); err != nil {
		logger.ProcLog.Errorf("PatchSubscriberByID err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.Status(http.StatusNoContent)
}

func removeCdrFile(cdrFilePath string) {
	files, err := filepath.Glob(cdrFilePath + "*.cdr")
	if err != nil {
		logger.BillingLog.Warnf("CDR file not found in %s", cdrFilePath)
	}

	for _, file := range files {
		if _, err = os.Stat(file); err == nil {
			logger.BillingLog.Infof("Remove CDR file: %s", file)
			if err = os.Remove(file); err != nil {
				logger.BillingLog.Warnf("Failed to remove CDR file: %s\n", file)
			}
		}
	}
}

// Delete subscriber by IMSI(ueId) and PlmnID(servingPlmnId)
func DeleteSubscriberByID(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Delete One Subscriber Data")
	ueId := c.Param("ueId")
	supi := gpsiToSupi(ueId)
	servingPlmnId := c.Param("servingPlmnId")
	// checking whether supi is successfully transformed to supi or not
	if supi == "" {
		logger.ProcLog.Errorf("DeleteSubscriberByID err: supi does not exists")
		c.JSON(http.StatusNotFound, gin.H{
			"cause": "supi does not exists",
		})
		return
	}
	var claims jwt.MapClaims = nil
	dbOperation(supi, servingPlmnId, "delete", nil, nil, claims, false)

	CdrFilePath := "/tmp/webconsole/"
	removeCdrFile(CdrFilePath)

	c.Status(http.StatusNoContent)
}

// Delete multiple subscribers
func DeleteMultipleSubscribers(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Delete Multiple Subscribers")
	var subsDatas []*SubsListIE
	if err := c.ShouldBindJSON(&subsDatas); err != nil {
		logger.ProcLog.Errorf("DeleteMultipleSubscribers err: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"cause": "JSON format incorrect",
		})
		return
	}

	var claims jwt.MapClaims = nil

	dbOperation("", "", "delete", nil, subsDatas, claims, true)

	c.Status(http.StatusNoContent)
}

func GetRegisteredUEContext(c *gin.Context) {
	setCorsHeader(c)

	logger.ProcLog.Infoln("Get Registered UE Context")

	webuiSelf := webui_context.GetSelf()
	webuiSelf.UpdateNfProfiles()

	isAdmin := CheckAuth(c)

	supi, supiExists := c.Params.Get("supi")
	// TODO: support fetching data from multiple AMFs
	if amfUris := webuiSelf.GetOamUris(models.NrfNfManagementNfType_AMF); amfUris != nil {
		var requestUri string

		if supiExists {
			requestUri = fmt.Sprintf("%s/namf-oam/v1/registered-ue-context/%s", amfUris[0], supi)
		} else {
			requestUri = fmt.Sprintf("%s/namf-oam/v1/registered-ue-context", amfUris[0])
		}

		ctx, pd, tokerErr := webui_context.GetSelf().GetTokenCtx(
			models.ServiceName_NAMF_OAM, models.NrfNfManagementNfType_AMF)
		if tokerErr != nil {
			logger.ProcLog.Errorf("GetTokenCtx error: %+v", tokerErr)
			c.JSON(http.StatusInternalServerError, pd)
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUri, nil)
		if err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		if err = webui_context.GetSelf().RequestBindToken(req, ctx); err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		resp, err := httpsClient.Do(req)
		if err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				logger.ProcLog.Error(closeErr)
			}
		}()

		// Filter by tenant.
		tenantId, err := GetTenantId(c)
		if err != nil {
			logger.ProcLog.Errorln(err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"cause": "Illegal or Nil token",
			})
			return
		}

		if isAdmin {
			sendResponseToClient(c, resp)
		} else {
			sendResponseToClientFilterTenant(c, resp, tenantId)
		}
	} else {
		logger.ProcLog.Warningln("No AMF found!")
		c.JSON(http.StatusNotFound, gin.H{
			"cause": "No AMF Found",
		})
	}
}

func GetUEPDUSessionInfo(c *gin.Context) {
	setCorsHeader(c)

	logger.ProcLog.Infoln("Get UE PDU Session Info")

	webuiSelf := webui_context.GetSelf()
	webuiSelf.UpdateNfProfiles()

	smContextRef, smContextRefExists := c.Params.Get("smContextRef")
	if !smContextRefExists {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	// TODO: support fetching data from multiple SMF
	if smfUris := webuiSelf.GetOamUris(models.NrfNfManagementNfType_SMF); smfUris != nil {
		requestUri := fmt.Sprintf("%s/nsmf-oam/v1/ue-pdu-session-info/%s", smfUris[0], smContextRef)

		ctx, pd, tokerErr := webui_context.GetSelf().GetTokenCtx(
			models.ServiceName_NSMF_OAM, models.NrfNfManagementNfType_SMF)
		if tokerErr != nil {
			logger.ProcLog.Errorf("GetTokenCtx error: %+v", tokerErr)
			c.JSON(http.StatusInternalServerError, pd)
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUri, nil)
		if err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		if err = webui_context.GetSelf().RequestBindToken(req, ctx); err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		resp, err := httpsClient.Do(req)
		if err != nil {
			logger.ProcLog.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				logger.ProcLog.Error(closeErr)
			}
		}()
		sendResponseToClient(c, resp)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"cause": "No SMF Found",
		})
	}
}

func ChangePasswordInfo(c *gin.Context) {
	setCorsHeader(c)

	// Need to get tenantId.
	tenantId, err := GetTenantId(c)
	if err != nil {
		logger.ProcLog.Errorln(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"cause": "Illegal or Nil token",
		})
		return
	}

	var newUserData User
	if err = c.ShouldBindJSON(&newUserData); err != nil {
		logger.ProcLog.Errorln(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	filterEmailOnly := bson.M{"tenantId": tenantId, "email": newUserData.Email}
	userDataInterface, err := mongoapi.RestfulAPIGetOne(userDataColl, filterEmailOnly)
	if err != nil {
		logger.ProcLog.Errorf("ChangePassword err: %+v", err)
	}
	if len(userDataInterface) == 0 {
		c.JSON(http.StatusNotFound, bson.M{})
		return
	}

	var userData User
	err = json.Unmarshal(mapToByte(userDataInterface), &userData)
	if err != nil {
		logger.ProcLog.Errorf("JSON Unmarshal err: %+v", err)
	}

	if newUserData.EncryptedPassword != "" {
		hash, err_gen := bcrypt.GenerateFromPassword([]byte(newUserData.EncryptedPassword), DefaultBcryptCost)
		if err_gen != nil {
			logger.ProcLog.Errorf("GenerateFromPassword err: %+v", err_gen)
		}
		userData.EncryptedPassword = string(hash)
	}

	userBsonM := toBsonM(userData)
	if _, err_put := mongoapi.RestfulAPIPost(userDataColl, filterEmailOnly, userBsonM); err_put != nil {
		logger.ProcLog.Errorf("PutUserByID err: %+v", err_put)
	}

	c.JSON(http.StatusOK, userData)
}

func OptionsSubscribers(c *gin.Context) {
	setCorsHeader(c)

	c.Status(http.StatusNoContent)
}

// Delete profile by profileName
func DeleteProfile(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Delete One Profile Data")

	profileName := c.Param("profileName")
	pf, err := mongoapi.RestfulAPIGetOne(profileDataColl, bson.M{"profileName": profileName})
	if err != nil {
		logger.ProcLog.Errorf("DeleteProfile err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	if len(pf) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"cause": profileName + " does not exist"})
		return
	}

	if err = dbProfileOperation(profileName, "delete", nil, nil, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": profileName + " has already been deleted"})
}

// Delete multiple profiles
func DeleteMultipleProfiles(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Delete Multiple Profiles")
	var profileDatas []*Profile
	if err := c.ShouldBindJSON(&profileDatas); err != nil {
		logger.ProcLog.Errorf("DeleteMultipleProfiles err: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"cause": "JSON format incorrect",
		})
		return
	}

	if err := dbProfileOperation("", "delete", nil, profileDatas, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Get profile list
func GetProfiles(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Get All Profiles List")

	_, err := GetTenantId(c)
	if err != nil {
		logger.ProcLog.Errorln(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"cause": "Illegal Token"})
		return
	}

	pfs := make([]string, 0)
	profileList, err := mongoapi.RestfulAPIGetMany(profileDataColl, bson.M{})
	if err != nil {
		logger.ProcLog.Errorf("GetProfiles err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	for _, profile := range profileList {
		profileName := profile["profileName"]

		pfs = append(pfs, profileName.(string))
	}
	c.JSON(http.StatusOK, pfs)
}

// Get profile by profileName
func GetProfile(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Get One Profile Data")

	profileName := c.Param("profileName")

	profile, err := mongoapi.RestfulAPIGetOne(profileDataColl, bson.M{"profileName": profileName})
	if err != nil {
		logger.ProcLog.Errorf("GetProfile err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	var pf Profile
	err = json.Unmarshal(mapToByte(profile), &pf)
	if err != nil {
		logger.ProcLog.Errorf("JSON Unmarshal err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pf)
}

// Post profile
func PostProfile(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Post One Profile Data")

	tokenStr := c.GetHeader("Token")
	_, err := ParseJWT(tokenStr)
	if err != nil {
		logger.ProcLog.Errorln(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"cause": "Illegal Token"})
		return
	}

	var profile Profile
	if err = c.ShouldBindJSON(&profile); err != nil {
		logger.ProcLog.Errorf("PostProfile err: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"cause": "JSON format incorrect"})
		return
	}

	tenantData, err := mongoapi.RestfulAPIGetOne(tenantDataColl, bson.M{"tenantName": "admin"})
	if err != nil {
		logger.ProcLog.Errorf("GetProfile err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	profile.TenantId = tenantData["tenantId"].(string)

	pf, err := mongoapi.RestfulAPIGetOne(profileDataColl, bson.M{"profileName": profile.ProfileName})
	if err != nil {
		logger.ProcLog.Errorf("GetProfile err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	if len(pf) != 0 {
		c.JSON(http.StatusConflict, gin.H{"cause": profile.ProfileName + " already exists"})
		return
	}

	logger.ProcLog.Infof("PostProfile: %+v", profile.ProfileName)
	if err = dbProfileOperation(profile.ProfileName, "post", &profile, nil, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// Put profile by profileName
func PutProfile(c *gin.Context) {
	setCorsHeader(c)
	logger.ProcLog.Infoln("Put One Profile Data")

	profileName := c.Param("profileName")

	var profile Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		logger.ProcLog.Errorf("PutProfile err: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"cause": "JSON format incorrect"})
		return
	}

	pf, err := mongoapi.RestfulAPIGetOne(profileDataColl, bson.M{"profileName": profile.ProfileName})
	if err != nil {
		logger.ProcLog.Errorf("PutProfile err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	if len(pf) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"cause": profileName + " does not exist"})
		return
	}

	tenantData, err := mongoapi.RestfulAPIGetOne(tenantDataColl, bson.M{"tenantName": "admin"})
	if err != nil {
		logger.ProcLog.Errorf("GetProfile err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	profile.TenantId = tenantData["tenantId"].(string)

	logger.ProcLog.Infof("PutProfile: %+v", profile.ProfileName)
	if err = dbProfileOperation(profileName, "put", &profile, nil, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"cause": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func dbProfileOperation(
	profileName string,
	method string,
	profile *Profile,
	profileDatas []*Profile,
	multiple bool,
) (err error) {
	err = nil
	filter := bson.M{"profileName": profileName}

	// Replace all data with new one
	switch method {
	case "put":
		if err = mongoapi.RestfulAPIDeleteOne(profileDataColl, filter); err != nil {
			logger.ProcLog.Errorf("DeleteProfile err: %+v", err)
		}
	case "delete":
		if multiple {
			multipleFilterConditions := []bson.M{}
			for _, profileData := range profileDatas {
				multipleFilterConditions = append(multipleFilterConditions, bson.M{
					"profileName": profileData.ProfileName,
				})
			}
			multipleFilter := bson.M{"$or": multipleFilterConditions}
			if err = mongoapi.RestfulAPIDeleteMany(profileDataColl, multipleFilter); err != nil {
				logger.ProcLog.Errorf("DeleteMultipleProfiles err: %+v", err)
			}
		} else {
			if err = mongoapi.RestfulAPIDeleteOne(profileDataColl, filter); err != nil {
				logger.ProcLog.Errorf("DeleteProfile err: %+v", err)
			}
		}
	}

	// Deal with error & early return
	if err != nil {
		return err
	}

	// Insert data
	if method == "post" || method == "put" {
		profileBsonM := toBsonM(profile)
		if _, err = mongoapi.RestfulAPIPost(profileDataColl, filter, profileBsonM); err != nil {
			logger.ProcLog.Errorf("PostProfile err: %+v", err)
		}
	}

	return err
}
