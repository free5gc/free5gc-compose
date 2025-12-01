import React from "react";
import {
    Box,
    Grid,
    Table,
    TableBody,
    TableCell,
    TableRow,
} from "@mui/material";
import { ChargingData } from "../../api/api";

const ChargingCfg = ({
    chargingData,
}: {
    chargingData: ChargingData;
}) => {
    const isOnlineCharging = chargingData.chargingMethod === "Online";

    return (
        <Box sx={{ m: 2 }}>
            <Grid container spacing={2}>
                <Grid item xs={12}>
                    <h4>Charging Config</h4>
                </Grid>
            </Grid>
            <Table>
                <TableBody>
                    <TableRow>
                        <TableCell style={{ width: "40%" }}>Charging Method</TableCell>
                        <TableCell>{chargingData.chargingMethod}</TableCell>
                    </TableRow>
                </TableBody>
                {isOnlineCharging && (
                    <TableBody>
                        <TableRow>
                            <TableCell style={{ width: "40%" }}> Quota </TableCell>
                            <TableCell>{chargingData.quota}</TableCell>
                        </TableRow>
                    </TableBody>
                )}
                <TableBody>
                    <TableRow>
                        <TableCell style={{ width: "40%" }}> Unit Cost </TableCell>
                        <TableCell>{chargingData.unitCost}</TableCell>
                    </TableRow>
                </TableBody>
            </Table>
        </Box>
    );
};

export default ChargingCfg; 