import React from "react";
import { useState, useEffect } from "react";
import { useLocation, useParams } from "react-router-dom";

import axios from "../axios";
import { UeContext, PduSessionInfo } from "../api/api";

import Dashboard from "../Dashboard";
import { Card, Table, TableBody, TableCell, TableRow } from "@mui/material";

export default function StatusRead() {
  const location = useLocation();
  const ueContext = location.state as UeContext;
  const [refresh, setRefresh] = useState<boolean>(false);

  const { id } = useParams<{
    id: string;
  }>();

  const [data, setData] = useState<PduSessionInfo[]>([]);

  useEffect(() => {
    const endpoints = [];
    for (let i = 0; i < ueContext.PduSessions!.length; i++) {
      endpoints.push("/api/ue-pdu-session-info/" + ueContext.PduSessions![i].SmContextRef);
    }
    const fetchData = endpoints.map((endpoint) => axios.get(endpoint));
    Promise.all([...fetchData]).then((res) => {
      const pdus: PduSessionInfo[] = res.map((item) => item.data);
      setData(pdus);
    });
  }, [id, refresh]);

  return (
    <Dashboard title="Registered UEs" refreshAction={() => setRefresh(!refresh)}>
      <h3>AMF Information [SUPI:{ueContext.Supi}]</h3>
      <Card variant="outlined">
        <Table>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>AccessType</TableCell>
              <TableCell>{ueContext.AccessType}</TableCell>
            </TableRow>
          </TableBody>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>CmState</TableCell>
              <TableCell>{ueContext.CmState}</TableCell>
            </TableRow>
          </TableBody>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>GUTI</TableCell>
              <TableCell>{ueContext.Guti}</TableCell>
            </TableRow>
          </TableBody>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>Mcc</TableCell>
              <TableCell>{ueContext.Mcc}</TableCell>
            </TableRow>
          </TableBody>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>Mnc</TableCell>
              <TableCell>{ueContext.Mnc}</TableCell>
            </TableRow>
          </TableBody>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>Supi</TableCell>
              <TableCell>{ueContext.Supi}</TableCell>
            </TableRow>
          </TableBody>
          <TableBody>
            <TableRow>
              <TableCell style={{ width: "40%" }}>Tac</TableCell>
              <TableCell>{ueContext.Tac}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        {ueContext.PduSessions!.map((row) => (
          <Table key={row.PduSessionId}>
            <TableBody>
              <TableRow>
                <TableCell style={{ width: "40%" }}>Dnn</TableCell>
                <TableCell>{row.Dnn}</TableCell>
              </TableRow>
            </TableBody>
            <TableBody>
              <TableRow>
                <TableCell style={{ width: "40%" }}>PduSessionId</TableCell>
                <TableCell>{row.PduSessionId}</TableCell>
              </TableRow>
            </TableBody>
            <TableBody>
              <TableRow>
                <TableCell style={{ width: "40%" }}>Sd</TableCell>
                <TableCell>{row.Sd}</TableCell>
              </TableRow>
            </TableBody>
            <TableBody>
              <TableRow>
                <TableCell style={{ width: "40%" }}>SmContextRef</TableCell>
                <TableCell>{row.SmContextRef}</TableCell>
              </TableRow>
            </TableBody>
            <TableBody>
              <TableRow>
                <TableCell style={{ width: "40%" }}>Sst</TableCell>
                <TableCell>{row.Sst}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        ))}
      </Card>

      <br />

      {data.map((row) => (
        <div key="row.PDUSsessionId">
          <h3>SMF Information [SUPI:{ueContext.Supi}]</h3>
          <Card variant="outlined">
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>AnType</TableCell>
                  <TableCell>{row.AnType}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>Dnn</TableCell>
                  <TableCell>{row.Dnn}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>LocalSEID</TableCell>
                  <TableCell>{row.LocalSEID}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>PDUAddress</TableCell>
                  <TableCell>{row.PDUAddress}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>PDUSessionId</TableCell>
                  <TableCell>{row.PDUSessionID}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>RemoteSEID</TableCell>
                  <TableCell>{row.RemoteSEID}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>Sd</TableCell>
                  <TableCell>{row.Sd}</TableCell>
                </TableRow>
              </TableBody>
              <TableBody>
                <TableRow>
                  <TableCell style={{ width: "40%" }}>Sst</TableCell>
                  <TableCell>{row.Sst}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </Card>
        </div>
      ))}
    </Dashboard>
  );
}
