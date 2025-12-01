import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { config } from "../constants/config";

import axios from "../axios";
import { UeContext } from "../api/api";

import Dashboard from "../Dashboard";
import {
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TablePagination,
} from "@mui/material";

export default function AnalysisList() {
  const navigation = useNavigate();
  const [data, setData] = useState<UeContext[]>([]);
  const [limit, setLimit] = useState(50);
  const [page, setPage] = useState(0);
  const [refresh, setRefresh] = useState<boolean>(false);

  useEffect(() => {
    axios
      .get("/api")
      .then((res) => {
        setData(res.data);
      })
      .catch((e) => {
        console.log(e.message);
      });
  }, [refresh, limit, page]);

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

  const handleShowInfo = (ueContext: UeContext) => {
    if (ueContext.PduSessions!.length > 0) {
      navigation("/status/" + ueContext.PduSessions![0].SmContextRef!, { state: ueContext });
    }
  };

  const tableView = (
    <React.Fragment>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>SUPI</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Details</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.map((row) => (
            <TableRow key={row.Supi}>
              <TableCell>{row.Supi}</TableCell>
              <TableCell>{row.CmState}</TableCell>
              <TableCell>
                <Button color="primary" variant="contained" onClick={() => handleShowInfo(row)}>
                  Show Info
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {pager()}
    </React.Fragment>
  );

  return (
    <Dashboard title="Analysis" refreshAction={() => setRefresh(!refresh)}>
      <br />
      {data == null || data.length === 0 ? <div>No App Data</div> : tableView}
    </Dashboard>
  );
}
