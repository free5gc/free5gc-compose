import React from "react";
import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";

import axios from "../axios";
import {
  Nssai,
  Profile,
  DnnConfiguration,
} from "../api/api";

import Dashboard from "../Dashboard";
import {
  Button,
  Box,
  Card,
  Checkbox,
  Grid,
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "@mui/material";
import FlowRule from "./Component/FlowRule";
import ChargingCfg from "./Component/ChargingCfg";
import UpSecurity from "./Component/UpSecurity";

export default function ProfileRead() {
  const { profileName } = useParams<{ profileName: string }>();
  const navigation = useNavigate();

  const [data, setData] = useState<Profile | null>(null);

  function toHex(v: number | undefined): string {
    return ("00" + v?.toString(16).toUpperCase()).substr(-2);
  }

  useEffect(() => {
    axios.get("/api/profile/" + profileName).then((res) => {
      setData(res.data);
    });
  }, [profileName]);

  const handleEdit = () => {
    navigation("/profile/create/" + profileName);
  };

  const isDefaultNssai = (nssai: Nssai | undefined) => {
    if (nssai === undefined || data == null) {
      return false;
    } else {
      for (
        let i = 0;
        i < data.AccessAndMobilitySubscriptionData.nssai!.defaultSingleNssais!.length;
        i++
      ) {
        const defaultNssai = data.AccessAndMobilitySubscriptionData.nssai!.defaultSingleNssais![i];
        if (defaultNssai.sd === nssai.sd && defaultNssai.sst === nssai.sst) {
          return true;
        }
      }
      return false;
    }
  };

  const chargingConfig = (dnn: string, snssai: Nssai, filter: string | undefined) => {
    const flowKey = toHex(snssai.sst) + snssai.sd;
    for (const chargingData of data?.ChargingDatas ?? []) {
      if (
        chargingData.snssai === flowKey &&
        chargingData.dnn === dnn &&
        chargingData.filter === filter
      ) {
        return <ChargingCfg chargingData={chargingData} />;
      }
    }
  };

  const flowRule = (dnn: string, snssai: Nssai) => {
    const flowKey = toHex(snssai.sst) + snssai.sd;
    if (data?.FlowRules === undefined) {
      return <div></div>;
    }
    return data.FlowRules.filter((flow) => flow.dnn === dnn && flow.snssai === flowKey).map(
      (flow) => (
        <FlowRule
          flow={flow}
          dnn={dnn}
          data={data}
          chargingConfig={() => chargingConfig(dnn, snssai, flow.filter)}
        />
      ),
    );
  };

  const upSecurity = (dnn: DnnConfiguration | undefined) => {
    if (dnn === undefined || dnn.upSecurity === undefined) {
      return <div></div>;
    }
    return <UpSecurity dnn={dnn} />;
  };

  return (
    <Dashboard title="Profile" refreshAction={() => {}}>
      <Grid 
        item 
        xs={12} 
        sx={{ 
          display: 'flex', 
          justifyContent: 'flex-end' 
        }}
      >
        <Button color="primary" variant="contained" onClick={handleEdit} sx={{ m: 1 }}>
          EDIT
        </Button>
      </Grid>
      <br />
      <Card variant="outlined">
        <Table>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>Profile Name</TableCell>
              <TableCell>{data?.profileName}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Card>

      <h3>Subscribed UE AMBR</h3>
      <Card variant="outlined">
        <Table>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>Uplink</TableCell>
              <TableCell>
                {data?.AccessAndMobilitySubscriptionData.subscribedUeAmbr?.uplink}
              </TableCell>
            </TableRow>
          </TableBody>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>Downlink</TableCell>
              <TableCell>
                {data?.AccessAndMobilitySubscriptionData.subscribedUeAmbr?.downlink}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Card>

      {/* S-NSSAI Configurations */}
      {data?.SessionManagementSubscriptionData?.map((row, index) => (
        <div key={index}>
          <h3>S-NSSAI Configuration</h3>
          <Card variant="outlined">
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>SST</TableCell>
                  <TableCell>{row.singleNssai?.sst}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>SD</TableCell>
                  <TableCell>{row.singleNssai?.sd}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>Default S-NSSAI</TableCell>
                  <TableCell>
                    <Checkbox checked={isDefaultNssai(row.singleNssai)} />
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
            {row.dnnConfigurations &&
              Object.keys(row.dnnConfigurations!).map((dnn) => (
                <div key={dnn}>
                  <Box sx={{ m: 2 }}>
                    <h4>DNN Configurations</h4>
                    <Card variant="outlined">
                      <Table>
                        <TableBody>
                          <TableRow>
                            <TableCell style={{ width: "40%" }}>Data Network Name</TableCell>
                            <TableCell>{dnn}</TableCell>
                          </TableRow>
                        </TableBody>
                        <TableBody>
                          <TableRow>
                            <TableCell style={{ width: "40%" }}>Uplink AMBR</TableCell>
                            <TableCell>
                              {row.dnnConfigurations![dnn].sessionAmbr?.uplink} /{" "}
                              {row.dnnConfigurations![dnn].sessionAmbr?.downlink}
                            </TableCell>
                          </TableRow>
                        </TableBody>
                        <TableBody>
                          <TableRow>
                            <TableCell style={{ width: "40%" }}>Downlink AMBR</TableCell>
                            <TableCell>
                              {row.dnnConfigurations![dnn].sessionAmbr?.downlink}
                            </TableCell>
                          </TableRow>
                        </TableBody>
                        <TableBody>
                          <TableRow>
                            <TableCell style={{ width: "40%" }}>Default 5QI</TableCell>
                            <TableCell>
                              {row.dnnConfigurations![dnn]["5gQosProfile"]?.["5qi"]}
                            </TableCell>
                          </TableRow>
                        </TableBody>
                        <TableBody>
                          <TableRow>
                            <TableCell style={{ width: "40%" }}>Static IPv4 Address</TableCell>
                            <TableCell>
                              {row.dnnConfigurations![dnn]["staticIpAddress"] == null
                                ? "Not Set"
                                : row.dnnConfigurations![dnn]["staticIpAddress"]?.length == 0
                                  ? ""
                                  : row.dnnConfigurations![dnn]["staticIpAddress"]![0].ipv4Addr!}
                            </TableCell>
                          </TableRow>
                        </TableBody>
                      </Table>
                      {flowRule(dnn, row.singleNssai!)}
                      {upSecurity(row.dnnConfigurations![dnn])}
                      {chargingConfig("", row.singleNssai!, "")}
                    </Card>
                  </Box>
                </div>
              ))}
          </Card>
        </div>
      ))}
    </Dashboard>
  );
}
