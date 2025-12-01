package WebUI_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/free5gc/openapi/models"
	smf_factory "github.com/free5gc/smf/pkg/factory"
	"github.com/free5gc/webconsole/backend/WebUI"
)

func TestGetStaticIpPoolsFromUserPlaneInfomation(t *testing.T) {
	testcases := []struct {
		Name          string
		Snssai        models.Snssai
		Dnn           string
		Userplaneinfo smf_factory.UserPlaneInformation
		ExpectIpPools []string
	}{
		{
			Name: "Simple",
			Snssai: models.Snssai{
				Sst: 1,
				Sd:  "010203",
			},
			Dnn: "internet1",
			Userplaneinfo: smf_factory.UserPlaneInformation{
				UPNodes: map[string]*smf_factory.UPNode{
					"UPF": {
						Type: "UPF",
						SNssaiInfos: []*smf_factory.SnssaiUpfInfoItem{
							{
								SNssai: &models.Snssai{
									Sst: 1,
									Sd:  "010203",
								},
								DnnUpfInfoList: []*smf_factory.DnnUpfInfoItem{
									{
										Dnn: "internet1",
										StaticPools: []*smf_factory.UEIPPool{
											{
												Cidr: "10.60.100.0/24",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectIpPools: []string{
				"10.60.100.0/24",
			},
		},
		{
			Name: "From second UPF",
			Snssai: models.Snssai{
				Sst: 1,
				Sd:  "112233",
			},
			Dnn: "internet",
			Userplaneinfo: smf_factory.UserPlaneInformation{
				UPNodes: map[string]*smf_factory.UPNode{
					"UPF1": {
						Type: "UPF",
						SNssaiInfos: []*smf_factory.SnssaiUpfInfoItem{
							{
								SNssai: &models.Snssai{
									Sst: 1,
									Sd:  "010203",
								},
								DnnUpfInfoList: []*smf_factory.DnnUpfInfoItem{
									{
										Dnn: "internet1",
										StaticPools: []*smf_factory.UEIPPool{
											{
												Cidr: "10.60.100.0/24",
											},
										},
									},
								},
							},
						},
					},
					"UPF2": {
						Type: "UPF",
						SNssaiInfos: []*smf_factory.SnssaiUpfInfoItem{
							{
								SNssai: &models.Snssai{
									Sst: 1,
									Sd:  "112233",
								},
								DnnUpfInfoList: []*smf_factory.DnnUpfInfoItem{
									{
										Dnn: "internet",
										StaticPools: []*smf_factory.UEIPPool{
											{
												Cidr: "10.61.100.0/24",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectIpPools: []string{
				"10.61.100.0/24",
			},
		},
		{
			Name: "Two Pools",
			Snssai: models.Snssai{
				Sst: 1,
				Sd:  "010203",
			},
			Dnn: "internet",
			Userplaneinfo: smf_factory.UserPlaneInformation{
				UPNodes: map[string]*smf_factory.UPNode{
					"UPF": {
						Type: "UPF",
						SNssaiInfos: []*smf_factory.SnssaiUpfInfoItem{
							{
								SNssai: &models.Snssai{
									Sst: 1,
									Sd:  "010203",
								},
								DnnUpfInfoList: []*smf_factory.DnnUpfInfoItem{
									{
										Dnn: "internet",
										StaticPools: []*smf_factory.UEIPPool{
											{
												Cidr: "10.60.100.0/24",
											},
											{
												Cidr: "10.60.101.0/24",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectIpPools: []string{
				"10.60.100.0/24",
				"10.60.101.0/24",
			},
		},
	}

	for _, tc := range testcases {
		pools := []netip.Prefix{}
		for _, pool := range tc.ExpectIpPools {
			net, err := netip.ParsePrefix(pool)
			require.NoError(t, err)

			pools = append(pools, net)
		}

		resultPools, err := WebUI.GetStaticIpPoolsFromUserPlaneInfomation(&tc.Userplaneinfo, tc.Snssai, tc.Dnn)
		require.NoError(t, err)

		require.Equal(t, pools, resultPools)
	}
}

func TestVerifyStaticIpProcedure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testcases := []struct {
		Name    string
		Scope   WebUI.VerifyScope
		IpPools []string
		Result  bool
	}{
		{
			Name: "One Static Pool - PASS",
			Scope: WebUI.VerifyScope{
				Supi:   "imsi-",
				Sst:    1,
				Sd:     "010203",
				Dnn:    "internet",
				Ipaddr: "10.163.100.100",
			},
			IpPools: []string{"10.163.100.0/24"},
			Result:  true,
		},
		{
			Name: "One Static Pool - Not in pool",
			Scope: WebUI.VerifyScope{
				Supi:   "imsi-",
				Sst:    1,
				Dnn:    "internet",
				Ipaddr: "10.163.163.1",
			},
			IpPools: []string{"10.163.100.0/24"},
			Result:  false,
		},
		{
			Name: "Two Static Pools - PASS (In first)",
			Scope: WebUI.VerifyScope{
				Supi:   "imsi-",
				Sst:    1,
				Sd:     "010203",
				Dnn:    "internet",
				Ipaddr: "10.163.100.100",
			},
			IpPools: []string{
				"10.163.100.0/24",
				"10.163.101.0/24",
			},
			Result: true,
		},
		{
			Name: "Two Static Pools - PASS (In Second)",
			Scope: WebUI.VerifyScope{
				Supi:   "imsi-",
				Sst:    1,
				Sd:     "010203",
				Dnn:    "internet",
				Ipaddr: "10.163.101.100",
			},
			IpPools: []string{
				"10.163.100.0/24",
				"10.163.101.0/24",
			},
			Result: true,
		},
		{
			Name: "Two Static Pools - Not in pools",
			Scope: WebUI.VerifyScope{
				Supi:   "imsi-",
				Sst:    1,
				Dnn:    "internet",
				Ipaddr: "10.163.163.1",
			},
			IpPools: []string{
				"10.163.100.0/24",
				"10.163.101.0/24",
			},
			Result: false,
		},
	}

	for _, tc := range testcases {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		pools := []netip.Prefix{}
		for _, pool := range tc.IpPools {
			net, err := netip.ParsePrefix(pool)
			require.NoError(t, err)

			pools = append(pools, net)
		}

		WebUI.VerifyStaticIpProcedure(ctx, tc.Scope, pools)
		require.Equal(t, http.StatusOK, w.Code)

		var result gin.H
		rawByte := w.Body
		errUnmarshal := json.Unmarshal(rawByte.Bytes(), &result)
		require.NoError(t, errUnmarshal)

		valid, exist := result["valid"]
		require.True(t, exist)

		require.Equal(t, tc.Result, valid)
	}
}
