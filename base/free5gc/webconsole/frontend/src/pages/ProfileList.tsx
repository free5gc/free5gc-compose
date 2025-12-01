import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { config } from "../constants/config";
import Dashboard from "../Dashboard";
import axios from "../axios";
import {
  Alert,
  Box,
  Button,
  Grid,
  Snackbar,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TablePagination,
  TableRow,
  TextField,
  Checkbox,
} from "@mui/material";
import { ReportProblemRounded } from "@mui/icons-material";
import { MultipleDeleteProfileData, formatMultipleDeleteProfileToJson } from "../lib/jsonFormating";

interface Props {
  refresh: boolean;
  setRefresh: (v: boolean) => void;
}

function ProfileList(props: Props) {
  const navigation = useNavigate();
  const [data, setData] = useState<string[]>([]);
  const [limit, setLimit] = useState(50);
  const [page, setPage] = useState(0);
  const [searchTerm, setSearchTerm] = useState<string>("");
  const [isLoadError, setIsLoadError] = useState(false);
  const [isDeleteError, setIsDeleteError] = useState(false);
  const [selected, setSelected] = useState<MultipleDeleteProfileData[]>([]);

  useEffect(() => {
    axios
      .get("/api/profile")
      .then((res) => {
        setData(res.data);
      })
      .catch((e) => {
        setIsLoadError(true);
      });
  }, [props.refresh, limit, page]);

  if (isLoadError) {
    return (
      <Stack sx={{ mx: "auto", py: "2rem" }} spacing={2} alignItems={"center"}>
        <ReportProblemRounded sx={{ fontSize: "3rem" }} />
        <Box fontWeight={700}>Something went wrong</Box>
      </Stack>
    );
  }

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

  const onCreate = () => {
    navigation("/profile/create");
  };

  const onDelete = (profileName: string) => {
    const result = window.confirm("Delete profile?");
    if (!result) {
      return;
    }
    axios
      .delete("/api/profile/" + profileName)
      .then((res) => {
        props.setRefresh(!props.refresh);
      })
      .catch((err) => {
        setIsDeleteError(true);
        console.error(err.response.data.message);
      });
  };

  const handleModify = (profile: string) => {
    navigation("/profile/" + profile);
  };

  const handleEdit = (profile: string) => {
    navigation("/profile/create/" + profile);
  };

  const filteredData = data.filter((profile) =>
    profile.toLowerCase().includes(searchTerm.toLowerCase()),
  );

  const handleSearch = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
  };

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const newSelected = filteredData.map(row => ({
        profileName: row.toString()
      }));
      setSelected(newSelected);
      return;
    }
    setSelected([]);
  };

  const handleClick = (item: MultipleDeleteProfileData) => {
    const selectedIndex = selected.findIndex(
      s => s.profileName === item.profileName
    );
    let newSelected: MultipleDeleteProfileData[] = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selected, item);
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selected.slice(1));
    } else if (selectedIndex === selected.length - 1) {
      newSelected = newSelected.concat(selected.slice(0, -1));
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selected.slice(0, selectedIndex),
        selected.slice(selectedIndex + 1),
      );
    }

    setSelected(newSelected);
  };

  const isSelected = (item: MultipleDeleteProfileData) => 
    selected.some(s => s.profileName === item.profileName);

  const onDeleteSelected = () => {
    const selectedItems = selected.map(item => 
      `Profile Name: ${item.profileName}`
    );

    const confirmMessage = `Are you sure you want to delete the following profiles?\n\n${selectedItems.join('\n')}`;
    const result = window.confirm(confirmMessage);
    if (!result) {
      return;
    }

    const data = formatMultipleDeleteProfileToJson(selected);
    axios.delete("/api/profile", { data })
      .then(() => {
        props.setRefresh(!props.refresh);
        setSelected([]);
      })
      .catch((err) => {
        setIsDeleteError(true);
        console.error(err.response.data.message);
      });
  };

  if (data.length === 0) {
    return (
      <>
        <br />
        <div>
          No Profiles
          <br />
          <br />
          <Grid item xs={12}>
            <Button color="primary" variant="contained" onClick={() => onCreate()} sx={{ m: 1 }}>
              CREATE
            </Button>
          </Grid>
        </div>
      </>
    );
  }

  return (
    <>
      <br />
      <TextField
        label="Search Profile"
        variant="outlined"
        value={searchTerm}
        onChange={handleSearch}
        fullWidth
        margin="normal"
      />
      {selected.length > 0 && (
        <Box sx={{ mb: 2 }}>
          <Button
            color="error"
            variant="contained"
            onClick={onDeleteSelected}
          >
            Delete Selected ({selected.length})
          </Button>
        </Box>
      )}
      <Table>
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox">
              <Checkbox
                color="primary"
                indeterminate={selected.length > 0 && selected.length < filteredData.length}
                checked={filteredData.length > 0 && selected.length === filteredData.length}
                onChange={handleSelectAllClick}
              />
            </TableCell>
            <TableCell>Name</TableCell>
            <TableCell>Delete</TableCell>
            <TableCell>View</TableCell>
            <TableCell>Edit</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {filteredData.map((row, index) => {
            const item = { profileName: row.toString() };
            const isItemSelected = isSelected(item);
            return (
              <TableRow 
                key={index}
                hover
                onClick={() => handleClick(item)}
                role="checkbox"
                aria-checked={isItemSelected}
                selected={isItemSelected}
              >
                <TableCell padding="checkbox">
                  <Checkbox
                    color="primary"
                    checked={isItemSelected}
                  />
                </TableCell>
                <TableCell>{row.toString()}</TableCell>
                <TableCell>
                  <Button
                    color="primary"
                    variant="contained"
                    onClick={() => onDelete(row.toString())}
                  >
                    DELETE
                  </Button>
                </TableCell>
                <TableCell>
                  <Button color="primary" variant="contained" onClick={() => handleModify(row.toString())}>
                    VIEW
                  </Button>
                </TableCell>
                <TableCell>
                  <Button color="primary" variant="contained" onClick={() => handleEdit(row.toString())}>
                    EDIT
                  </Button>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
      {pager()}
      <Grid item xs={12}>
        <Button color="primary" variant="contained" onClick={() => onCreate()} sx={{ m: 1 }}>
          CREATE
        </Button>
      </Grid>

      <Snackbar
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
        open={isDeleteError}
        autoHideDuration={6000}
        onClose={() => setIsDeleteError(false)}
      >
        <Alert severity="error">Failed to delete profile</Alert>
      </Snackbar>
    </>
  );
}

function WithDashboard(Component: React.ComponentType<any>) {
  return function (props: any) {
    const [refresh, setRefresh] = useState<boolean>(false);

    return (
      <Dashboard title="PROFILE" refreshAction={() => setRefresh(!refresh)}>
        <Component {...props} refresh={refresh} setRefresh={(v: boolean) => setRefresh(v)} />
      </Dashboard>
    );
  };
}

export default WithDashboard(ProfileList);
