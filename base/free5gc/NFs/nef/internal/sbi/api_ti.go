package sbi

import (
	"net/http"

	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/metrics/sbi"
	"github.com/gin-gonic/gin"
)

func (s *Server) getTrafficInfluenceRoutes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Pattern: "/:afID/subscriptions",
			APIFunc: s.apiGetTrafficInfluenceSubscription,
		},
		{
			Method:  http.MethodPost,
			Pattern: "/:afID/subscriptions",
			APIFunc: s.apiPostTrafficInfluenceSubscription,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/:afID/subscriptions/:subID",
			APIFunc: s.apiGetIndividualTrafficInfluenceSubscription,
		},
		{
			Method:  http.MethodPut,
			Pattern: "/:afID/subscriptions/:subID",
			APIFunc: s.apiPutIndividualTrafficInfluenceSubscription,
		},
		{
			Method:  http.MethodPatch,
			Pattern: "/:afID/subscriptions/:subID",
			APIFunc: s.apiPatchIndividualTrafficInfluenceSubscription,
		},
		{
			Method:  http.MethodDelete,
			Pattern: "/:afID/subscriptions/:subID",
			APIFunc: s.apiDeleteIndividualTrafficInfluenceSubscription,
		},
	}
}

func (s *Server) apiGetTrafficInfluenceSubscription(gc *gin.Context) {
	s.Processor().GetTrafficInfluenceSubscription(
		gc, gc.Param("afID"))
}

func (s *Server) apiPostTrafficInfluenceSubscription(gc *gin.Context) {
	var tiSub models.NefTrafficInfluSub
	reqBody, err := gc.GetRawData()
	if err != nil {
		logger.SBILog.Errorf("Get Request Body error: %+v", err)
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError, pd)
		return
	}

	err = openapi.Deserialize(&tiSub, reqBody, "application/json")
	if err != nil {
		logger.SBILog.Errorf("Deserialize Request Body error: %+v", err)
		pd := openapi.ProblemDetailsMalformedReqSyntax(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	s.Processor().PostTrafficInfluenceSubscription(
		gc, gc.Param("afID"), &tiSub)
}

func (s *Server) apiGetIndividualTrafficInfluenceSubscription(gc *gin.Context) {
	s.Processor().GetIndividualTrafficInfluenceSubscription(
		gc, gc.Param("afID"), gc.Param("subID"))
}

func (s *Server) apiPutIndividualTrafficInfluenceSubscription(gc *gin.Context) {
	var tiSub models.NefTrafficInfluSub
	reqBody, err := gc.GetRawData()
	if err != nil {
		logger.SBILog.Errorf("Get Request Body error: %+v", err)
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError, pd)
		return
	}

	err = openapi.Deserialize(&tiSub, reqBody, "application/json")
	if err != nil {
		logger.SBILog.Errorf("Deserialize Request Body error: %+v", err)
		pd := openapi.ProblemDetailsMalformedReqSyntax(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	s.Processor().PutIndividualTrafficInfluenceSubscription(
		gc, gc.Param("afID"), gc.Param("subID"), &tiSub)
}

func (s *Server) apiPatchIndividualTrafficInfluenceSubscription(gc *gin.Context) {
	var tiSubPatch models.NefTrafficInfluSubPatch
	reqBody, err := gc.GetRawData()
	if err != nil {
		logger.SBILog.Errorf("Get Request Body error: %+v", err)
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError, pd)
		return
	}

	err = openapi.Deserialize(&tiSubPatch, reqBody, "application/json")
	if err != nil {
		logger.SBILog.Errorf("Deserialize Request Body error: %+v", err)
		pd := openapi.ProblemDetailsMalformedReqSyntax(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	s.Processor().PatchIndividualTrafficInfluenceSubscription(
		gc, gc.Param("afID"), gc.Param("subID"), &tiSubPatch)
}

func (s *Server) apiDeleteIndividualTrafficInfluenceSubscription(gc *gin.Context) {
	s.Processor().DeleteIndividualTrafficInfluenceSubscription(
		gc, gc.Param("afID"), gc.Param("subID"))
}
