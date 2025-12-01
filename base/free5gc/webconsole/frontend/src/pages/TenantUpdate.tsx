import React from "react";
import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";

import axios from "../axios";
import { Tenant } from "../api/api";

import Dashboard from "../Dashboard";

import { Button, Grid, TextField, Table, TableBody, TableCell, TableRow } from "@mui/material";

export default function TenantUpdate() {
  const { id } = useParams<{
    id: string;
  }>();
  const navigation = useNavigate();
  const [tenant, setTenant] = useState<Tenant>({tenantName: ""});

  useEffect(() => {
    axios.get("/api/tenant/" + id).then((res) => {
      setTenant(res.data);
    });
  }, [id]);

  const onUpdate = () => {
    axios.put("/api/tenant/" + id, tenant).then((res) => {
      console.log("put result:" + res);
      navigation("/tenant");
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
                label="Tenant Id"
                variant="outlined"
                fullWidth
                value={tenant.tenantId}
                InputLabelProps={{ shrink: true }}
                inputProps={{ readonly: true, disabled: true }}
              />
            </TableCell>
          </TableRow>
        </TableBody>
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
                InputLabelProps={{ shrink: true }}
              />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
      <br />
      <Grid item xs={12}>
        <Button color="primary" variant="contained" onClick={onUpdate} sx={{ m: 1 }}>
          Update
        </Button>
      </Grid>
    </Dashboard>
  );
}
