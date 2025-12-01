import React from "react";
import {
    Box,
    Card,
    Table,
    TableBody,
    TableRow,
    TableCell,
} from "@mui/material";
import { DnnConfiguration } from "../../api/api";

const UpSecurity = ({
    dnn,
}: {
    dnn: DnnConfiguration;
}) => {
    if (dnn.upSecurity === undefined) {
        return <div></div>;
    }

    const security = dnn.upSecurity;
    return (
        <div key={security.upIntegr}>
            <Box sx={{ m: 2 }}>
                <h4>UP Security</h4>
                <Card variant="outlined">
                    <Table>
                        <TableBody>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>Integrity of UP Security</TableCell>
                                <TableCell>{security.upIntegr}</TableCell>
                            </TableRow>
                            <TableRow>
                                <TableCell style={{ width: "40%" }}>Confidentiality of UP Security</TableCell>
                                <TableCell>{security.upConfid}</TableCell>
                            </TableRow>
                        </TableBody>
                    </Table>
                </Card>
            </Box>
        </div>
    );
};

export default UpSecurity; 