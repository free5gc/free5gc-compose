import React from "react";
import { useState, useContext } from "react";
import Button from "@mui/material/Button";
import CssBaseline from "@mui/material/CssBaseline";
import TextField from "@mui/material/TextField";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Container from "@mui/material/Container";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import axios from "../axios";
import { useNavigate } from "react-router-dom";
import { LoginContext } from "../LoginContext";

const theme = createTheme();

export default function SignIn() {
  const navigation = useNavigate();
  const [error, setError] = useState<string>("");
  const context = useContext(LoginContext);
  if (context === undefined) {
    throw new Error("LoginContext must be used within a LoginContext.Provider");
  }
  const { setUser } = context;

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    axios
      .post("/api/login", { username: data.get("email"), password: data.get("password") })
      .then((res) => {
        if (data.get("email") !== null) {
          setUser({ username: data.get("email")!.toString(), token: res.data.access_token });
        }
        setError("");
        navigation("/");
      })
      .catch((err) => {
        console.log(err.message);
        setError("Wrong credentials");
      });
  };

  return (
    <ThemeProvider theme={theme}>
      <Container component="main" maxWidth="xs">
        <CssBaseline />
        <Box
          sx={{
            marginTop: 8,
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
          }}
        >
          <img src="/free5gc_logo.png" className="App-logo" alt="logo" />
          <br />
          <Typography component="h1" variant="h6" color="red">
            {error}
          </Typography>
          <Box component="form" onSubmit={handleSubmit} noValidate sx={{ mt: 1 }}>
            <TextField
              margin="normal"
              required
              fullWidth
              id="email"
              label="Username"
              name="email"
              autoComplete="email"
              autoFocus
            />
            <TextField
              margin="normal"
              required
              fullWidth
              name="password"
              label="Password"
              type="password"
              id="password"
              autoComplete="current-password"
            />
            <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }}>
              Sign In
            </Button>
          </Box>
        </Box>
      </Container>
    </ThemeProvider>
  );
}
