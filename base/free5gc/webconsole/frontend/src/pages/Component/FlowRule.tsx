import React from "react";
import {
    Box,
    Card,
    Table,
    TableBody,
    TableRow,
    TableCell,
} from "@mui/material";
import { Nssai } from "../../api/api";

const qosFlow = (qosflows: any, sstSd: string, dnn: string, qosRef: number | undefined) => {
    if (qosflows === null) {
        return undefined;
    }
    for (const qos of qosflows) {
        if (qos.snssai === sstSd && qos.dnn === dnn && qos.qosRef === qosRef) {
            return qos;
        }
    }
    return undefined;
}

const FlowRule = ({
    dnn,
    flow,
    data,
    chargingConfig,
}: {
    dnn: string;
    flow: any;
    data: any;
    chargingConfig: (dnn: string, snssai: Nssai, filter: string | undefined) => JSX.Element | undefined;
}) => {
    return (
        <div key={flow.snssai}>
            <Box sx={{ m: 2 }}>
                <h4>Flow Rules</h4>
                <Card variant="outlined">
                    <Table>
                        <TableBody>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>IP Filter</TableCell>
                                <TableCell>{flow.filter}</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>Precedence</TableCell>
                                <TableCell>{flow.precedence}</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>5QI</TableCell>
                                <TableCell>{qosFlow(data.QosFlows, flow.snssai, dnn, flow.qosRef!)?.["5qi"]}</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>Uplink GBR</TableCell>
                                <TableCell>{qosFlow(data.QosFlows, flow.snssai, dnn, flow.qosRef!)?.gbrUL}</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>Downlink GBR</TableCell>
                                <TableCell>{qosFlow(data.QosFlows, flow.snssai, dnn, flow.qosRef!)?.gbrDL}</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>Uplink MBR</TableCell>
                                <TableCell>{qosFlow(data.QosFlows, flow.snssai, dnn, flow.qosRef!)?.mbrUL}</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>Downlink MBR</TableCell>
                                <TableCell>{qosFlow(data.QosFlows, flow.snssai, dnn, flow.qosRef!)?.mbrDL}</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>Charging Characteristics</TableCell>
                                <TableCell>{chargingConfig(dnn, flow.snssai, flow.filter)}</TableCell>
                            </TableRow>
                        </TableBody>
                    </Table>
                </Card>
            </Box>
        </div>
    );
};

export default FlowRule;
