import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { config } from "../constants/config";

import axios from "../axios";
import { User } from "../api/api";

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

export default function UserList() {
  const navigation = useNavigate();
  const [data, setData] = useState<User[]>([]);
  const [limit, setLimit] = useState(50);
  const [page, setPage] = useState(0);
  const [refresh, setRefresh] = useState<boolean>(false);

  const { id } = useParams<{
    id: string;
  }>();

  useEffect(() => {
    axios
      .get("/api/tenant/" + id + "/user")
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

  const onDelete = (uid: string | undefined) => {
    const result = window.confirm("Delete user?");
    if (!result) {
      return;
    }
    axios
      .delete("/api/tenant/" + id + "/user/" + uid)
      .then((res) => {
        console.log(res);
        setRefresh(!refresh);
      })
      .catch((err) => {
        alert(err.response.data.message);
      });
  };

  const onModify = (uid: string | undefined) => {
    navigation("/tenant/" + id + "/user/update/" + uid);
  };

  const onCreate = () => {
    navigation("/tenant/" + id + "/user/create");
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
            <TableCell>User ID</TableCell>
            <TableCell>User Email</TableCell>
            <TableCell>Delete</TableCell>
            <TableCell>Modify</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data?.map((row) => (
            <TableRow key={row.userId}>
              <TableCell>{row.userId}</TableCell>
              <TableCell>{row.email}</TableCell>
              <TableCell>
                <Button color="primary" variant="contained" onClick={() => onDelete(row.userId)}>
                  DELETE
                </Button>
              </TableCell>
              <TableCell>
                <Button color="primary" variant="contained" onClick={() => onModify(row.userId)}>
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
    <Dashboard title="Users" refreshAction={() => {}}>
      <br />
      {data == null || data.length === 0 ? (
        <div>
          No User
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
