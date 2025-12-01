import React from "react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import axios from "../axios";
import { Tenant } from "../api/api";

import Dashboard from "../Dashboard";
import { Button, Grid, TextField, Table, TableBody, TableCell, TableRow } from "@mui/material";

export default function TenantCreate() {
  const navigation = useNavigate();
  const [tenant, setTenant] = useState<Tenant>({tenantName: ""});

  const handleCreate = () => {
    console.log("Create");
    axios
      .post("/api/tenant", tenant)
      .then((res) => {
        console.log("post result:" + res);
        navigation("/tenant");
      })
      .catch((err) => {
        alert(err.response.data.message);
      });
  };

  const handleChangeTenantName = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ): void => {
    setTenant({ ...tenant, tenantName: event.target.value });
  };

  return (
    <Dashboard title="Tenant" refreshAction={() => {}}>
      <Table>
        <TableBody>
          <TableRow>
            <TableCell>
              <TextField
                label="Tenant Name"
                variant="outlined"
                required
                fullWidth
                value={tenant.tenantName}
                onChange={handleChangeTenantName}
              />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
      <br />
      <Grid item xs={12}>
        <Button color="primary" variant="contained" onClick={handleCreate} sx={{ m: 1 }}>
          CREATE
        </Button>
      </Grid>
    </Dashboard>
  );
}
