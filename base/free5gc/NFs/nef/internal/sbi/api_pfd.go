package sbi

import (
	"net/http"

	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/metrics/sbi"
	"github.com/gin-gonic/gin"
)

func (s *Server) getPFDManagementRoutes() []Route {
	return []Route{
		{
			Method:  http.MethodGet,
			Pattern: "/:scsAsID/transactions",
			APIFunc: s.apiGetPFDManagementTransactions,
		},
		{
			Method:  http.MethodPost,
			Pattern: "/:scsAsID/transactions",
			APIFunc: s.apiPostPFDManagementTransactions,
		},
		{
			Method:  http.MethodDelete,
			Pattern: "/:scsAsID/transactions",
			APIFunc: s.apiDeletePFDManagementTransactions,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/:scsAsID/transactions/:transID",
			APIFunc: s.apiGetIndividualPFDManagementTransaction,
		},
		{
			Method:  http.MethodPut,
			Pattern: "/:scsAsID/transactions/:transID",
			APIFunc: s.apiPutIndividualPFDManagementTransaction,
		},
		{
			Method:  http.MethodDelete,
			Pattern: "/:scsAsID/transactions/:transID",
			APIFunc: s.apiDeleteIndividualPFDManagementTransaction,
		},
		{
			Method:  http.MethodGet,
			Pattern: "/:scsAsID/transactions/:transID/applications/:appID",
			APIFunc: s.apiGetIndividualApplicationPFDManagement,
		},
		{
			Method:  http.MethodDelete,
			Pattern: "/:scsAsID/transactions/:transID/applications/:appID",
			APIFunc: s.apiDeleteIndividualApplicationPFDManagement,
		},
		{
			Method:  http.MethodPut,
			Pattern: "/:scsAsID/transactions/:transID/applications/:appID",
			APIFunc: s.apiPutIndividualApplicationPFDManagement,
		},
		{
			Method:  http.MethodPatch,
			Pattern: "/:scsAsID/transactions/:transID/applications/:appID",
			APIFunc: s.apiPatchIndividualApplicationPFDManagement,
		},
	}
}

func (s *Server) apiGetPFDManagementTransactions(gc *gin.Context) {
	s.Processor().GetPFDManagementTransactions(gc, gc.Param("scsAsID"))
}

func (s *Server) apiPostPFDManagementTransactions(gc *gin.Context) {
	var pfdMng models.PfdManagement
	reqBody, err := gc.GetRawData()
	if err != nil {
		logger.SBILog.Errorf("Get Request Body error: %+v", err)
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError, pd)
		return
	}

	err = openapi.Deserialize(&pfdMng, reqBody, "application/json")
	if err != nil {
		logger.SBILog.Errorf("Deserialize Request Body error: %+v", err)
		pd := openapi.ProblemDetailsMalformedReqSyntax(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	s.Processor().PostPFDManagementTransactions(gc, gc.Param("scsAsID"), &pfdMng)
}

func (s *Server) apiDeletePFDManagementTransactions(gc *gin.Context) {
	s.Processor().DeletePFDManagementTransactions(gc, gc.Param("scsAsID"))
}

func (s *Server) apiGetIndividualPFDManagementTransaction(gc *gin.Context) {
	s.Processor().GetIndividualPFDManagementTransaction(
		gc, gc.Param("scsAsID"), gc.Param("transID"))
}

func (s *Server) apiPutIndividualPFDManagementTransaction(gc *gin.Context) {
	var pfdMng models.PfdManagement
	reqBody, err := gc.GetRawData()
	if err != nil {
		logger.SBILog.Errorf("Get Request Body error: %+v", err)
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError,
			openapi.ProblemDetailsSystemFailure(err.Error()))
		return
	}

	err = openapi.Deserialize(&pfdMng, reqBody, "application/json")
	if err != nil {
		logger.SBILog.Errorf("Deserialize Request Body error: %+v", err)
		pd := openapi.ProblemDetailsMalformedReqSyntax(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	s.Processor().PutIndividualPFDManagementTransaction(
		gc, gc.Param("scsAsID"), gc.Param("transID"), &pfdMng)
}

func (s *Server) apiDeleteIndividualPFDManagementTransaction(gc *gin.Context) {
	s.Processor().DeleteIndividualPFDManagementTransaction(
		gc, gc.Param("scsAsID"), gc.Param("transID"))
}

func (s *Server) apiGetIndividualApplicationPFDManagement(gc *gin.Context) {
	s.Processor().GetIndividualApplicationPFDManagement(
		gc, gc.Param("scsAsID"), gc.Param("transID"), gc.Param("appID"))
}

func (s *Server) apiDeleteIndividualApplicationPFDManagement(gc *gin.Context) {
	s.Processor().DeleteIndividualApplicationPFDManagement(
		gc, gc.Param("scsAsID"), gc.Param("transID"), gc.Param("appID"))
}

func (s *Server) apiPutIndividualApplicationPFDManagement(gc *gin.Context) {
	var pfdData models.PfdData
	reqBody, err := gc.GetRawData()
	if err != nil {
		logger.SBILog.Errorf("Get Request Body error: %+v", err)
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError, pd)
		return
	}

	err = openapi.Deserialize(&pfdData, reqBody, "application/json")
	if err != nil {
		logger.SBILog.Errorf("Deserialize Request Body error: %+v", err)
		pd := openapi.ProblemDetailsMalformedReqSyntax(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	s.Processor().PutIndividualApplicationPFDManagement(
		gc, gc.Param("scsAsID"), gc.Param("transID"), gc.Param("appID"), &pfdData)
}

func (s *Server) apiPatchIndividualApplicationPFDManagement(gc *gin.Context) {
	var pfdData models.PfdData
	reqBody, err := gc.GetRawData()
	if err != nil {
		logger.SBILog.Errorf("Get Request Body error: %+v", err)
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError, pd)
		return
	}

	err = openapi.Deserialize(&pfdData, reqBody, "application/json")
	if err != nil {
		logger.SBILog.Errorf("Deserialize Request Body error: %+v", err)
		pd := openapi.ProblemDetailsMalformedReqSyntax(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	s.Processor().PatchIndividualApplicationPFDManagement(
		gc, gc.Param("scsAsID"), gc.Param("transID"), gc.Param("appID"), &pfdData)
}
