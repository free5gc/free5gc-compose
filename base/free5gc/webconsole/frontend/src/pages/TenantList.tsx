import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import { config } from "../constants/config";

import axios from "../axios";
import { Tenant } from "../api/api";

import Dashboard from "../Dashboard";
import {
  Button,
  Grid,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TablePagination,
} from "@mui/material";

export default function TenantList() {
  const navigation = useNavigate();
  const [data, setData] = useState<Tenant[]>([]);
  const [limit, setLimit] = useState(50);
  const [page, setPage] = useState(0);
  const [refresh, setRefresh] = useState<boolean>(false);

  useEffect(() => {
    axios
      .get("/api/tenant")
      .then((res) => {
        setData(res.data);
      })
      .catch((e) => {
        console.log(e.message);
      });
  }, [limit, page, refresh]);

  const handlePageChange = (
    _event: React.MouseEvent<HTMLButtonElement> | null,
    newPage?: number,
  ) => {
    if (newPage !== null) {
      setPage(newPage!);
    }
  };

  const handleLimitChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setLimit(Number(event.target.value));
  };

  const count = () => {
    return 0;
  };

  const pager = () => {
    if (config.enablePagination) {
      return (
        <TablePagination
          component="div"
          count={count()}
          onPageChange={handlePageChange}
          onRowsPerPageChange={handleLimitChange}
          page={page}
          rowsPerPage={limit}
          rowsPerPageOptions={[50, 100, 200]}
        />
      );
    } else {
      return <br />;
    }
  };

  const onDelete = (id: string | undefined) => {
    const result = window.confirm("Delete tenant?");
    if (!result) {
      return;
    }
    axios
      .delete("/api/tenant/" + id)
      .then((res) => {
        console.log(res);
        setRefresh(!refresh);
      })
      .catch((err) => {
        alert(err.response.data.message);
      });
  };

  const onModify = (id: string | undefined) => {
    navigation("/tenant/update/" + id);
  };

  const onCreate = () => {
    navigation("/tenant/create");
  };

  const createButton = () => {
    return (
      <Grid item xs={12}>
        <Button color="primary" variant="contained" onClick={() => onCreate()} sx={{ m: 1 }}>
          CREATE
        </Button>
      </Grid>
    );
  };

  const tableView = (
    <React.Fragment>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Tenant ID</TableCell>
            <TableCell>Tenant Name</TableCell>
            <TableCell>Delete</TableCell>
            <TableCell>Modify</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data?.map((row) => (
            <TableRow key={row.tenantId}>
              <TableCell>{row.tenantId}</TableCell>
              <TableCell>
                <Link to={"/tenant/" + row.tenantId + "/user"}>{row.tenantName}</Link>
              </TableCell>
              <TableCell>
                <Button color="primary" variant="contained" onClick={() => onDelete(row.tenantId)}>
                  DELETE
                </Button>
              </TableCell>
              <TableCell>
                <Button color="primary" variant="contained" onClick={() => onModify(row.tenantId)}>
                  MODIFY
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {pager()}
      {createButton()}
    </React.Fragment>
  );

  return (
    <Dashboard title="Tenants" refreshAction={() => setRefresh(!refresh)}>
      <br />
      {data == null || data.length === 0 ? (
        <div>
          No Tenant
          <br />
          <br />
          {createButton()}
        </div>
      ) : (
        tableView
      )}
    </Dashboard>
  );
}
