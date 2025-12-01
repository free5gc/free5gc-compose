import React from "react";
import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import axios from "../axios";
import { User } from "../api/api";

import Dashboard from "../Dashboard";
import {
  Button,
  Grid,
  InputAdornment,
  IconButton,
  TextField,
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "@mui/material";
import Visibility from "@mui/icons-material/Visibility";
import VisibilityOff from "@mui/icons-material/VisibilityOff";

export interface Password {
  password?: string;
  passwordConfirm?: string;
}

export default function UserCreate() {
  const navigation = useNavigate();
  const [user, setUser] = useState<User>({ email: "", encryptedPassword: "" });

  const [password, setPassword] = useState<Password>({});

  const [showPassword, setShowPassword] = useState(false);
  const handleClickShowPassword = () => setShowPassword(!showPassword);
  const handleMouseDownPassword = () => setShowPassword(!showPassword);

  const [showPasswordConfirm, setShowPasswordConfirm] = useState(false);
  const handleClickShowPasswordConfirm = () => setShowPasswordConfirm(!showPasswordConfirm);
  const handleMouseDownPasswordConfirm = () => setShowPasswordConfirm(!showPasswordConfirm);

  const { id } = useParams<{
    id: string;
  }>();

  const handleCreate = () => {
    if (password.password === undefined || password.password === "") {
      alert("Password can't be empty");
      return;
    }
    if (password.passwordConfirm === undefined || password.passwordConfirm === "") {
      alert("Password can't be empty");
      return;
    }
    if (password.password !== password.passwordConfirm) {
      alert("Password mismatch");
      return;
    }
    user.encryptedPassword = password.password;
    axios
      .post("/api/tenant/" + id + "/user", user)
      .then((res) => {
        console.log("post result:" + res);
        navigation("/tenant/" + id + "/user");
      })
      .catch((err) => {
        alert(err.response.data.message);
      });
  };

  const handleChangeEmail = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ): void => {
    setUser({ ...user, email: event.target.value });
  };

  const handleChangePassword = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ): void => {
    setPassword({ ...password, password: event.target.value });
  };

  const handleChangePasswordConfirm = (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ): void => {
    setPassword({ ...password, passwordConfirm: event.target.value });
  };

  return (
    <Dashboard title="User" refreshAction={() => {}}>
      <Table>
        <TableBody>
          <TableRow>
            <TableCell>
              <TextField
                label="User Email"
                variant="outlined"
                required
                fullWidth
                value={user.email}
                onChange={handleChangeEmail}
              />
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody>
          <TableRow>
            <TableCell>
              <TextField
                label="Password"
                variant="outlined"
                required
                fullWidth
                type={showPassword ? "text" : "password"}
                onChange={handleChangePassword}
                InputProps={{
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton
                        aria-label="toggle password visibility"
                        onClick={handleClickShowPassword}
                        onMouseDown={handleMouseDownPassword}
                      >
                        {showPassword ? <Visibility /> : <VisibilityOff />}
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
              />
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody>
          <TableRow>
            <TableCell>
              <TextField
                label="Confirm Password"
                variant="outlined"
                required
                fullWidth
                type={showPasswordConfirm ? "text" : "password"}
                onChange={handleChangePasswordConfirm}
                InputProps={{
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton
                        aria-label="toggle password visibility"
                        onClick={handleClickShowPasswordConfirm}
                        onMouseDown={handleMouseDownPasswordConfirm}
                      >
                        {showPasswordConfirm ? <Visibility /> : <VisibilityOff />}
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
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
