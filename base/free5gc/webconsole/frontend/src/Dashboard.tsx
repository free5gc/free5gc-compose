import React, { useContext, useState, useEffect } from "react";
import { styled, createTheme, ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import MuiDrawer from "@mui/material/Drawer";
import Box from "@mui/material/Box";
import MuiAppBar, { AppBarProps as MuiAppBarProps } from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import List from "@mui/material/List";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import IconButton from "@mui/material/IconButton";
import Container from "@mui/material/Container";
import Grid from "@mui/material/Grid";
import Paper from "@mui/material/Paper";
import MenuIcon from "@mui/icons-material/Menu";
import ChevronLeftIcon from "@mui/icons-material/ChevronLeft";
import { MainListItems } from "./ListItems";
import { LoginContext } from "./LoginContext";
import SimpleListMenu from "./SimpleListMenu";
import { useNavigate } from "react-router-dom";

const drawerWidth = 300;

interface AppBarProps extends MuiAppBarProps {
  open?: boolean;
}

const AppBar = styled(MuiAppBar, {
  shouldForwardProp: (prop) => prop !== "open",
})<AppBarProps>(({ theme, open }) => ({
  zIndex: theme.zIndex.drawer + 1,
  transition: theme.transitions.create(["width", "margin"], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  ...(open && {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth}px)`,
    transition: theme.transitions.create(["width", "margin"], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  }),
}));

const Drawer = styled(MuiDrawer, { shouldForwardProp: (prop) => prop !== "open" })(
  ({ theme, open }) => ({
    "& .MuiDrawer-paper": {
      position: "relative",
      whiteSpace: "nowrap",
      width: drawerWidth,
      transition: theme.transitions.create("width", {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.enteringScreen,
      }),
      boxSizing: "border-box",
      ...(!open && {
        overflowX: "hidden",
        transition: theme.transitions.create("width", {
          easing: theme.transitions.easing.sharp,
          duration: theme.transitions.duration.leavingScreen,
        }),
        width: theme.spacing(7),
        [theme.breakpoints.up("sm")]: {
          width: theme.spacing(9),
        },
      }),
    },
  }),
);

const mdTheme = createTheme();

export interface DashboardProps {
  children: React.ReactNode;
  title: string;
  refreshAction: () => void;
}

function Dashboard(props: DashboardProps) {
  const [open, setOpen] = React.useState(true);
  const toggleDrawer = () => {
    setOpen(!open);
  };
  const context = useContext(LoginContext);
  if (context === undefined) {
    throw new Error("LoginContext must be used within a LoginContext.Provider");
  }
  const { user } = context;

  const navigation = useNavigate();

  const [time, setTime] = useState<Date>(new Date());
  const [refreshInterval, setRefreshInterval] = useState(0);
  const [refreshString, setRefreshString] = useState("manual");

  // execute every time the refreshInterval changes to set the interval correctly
  // update the time value every x ms, which triggers refresh (see below)
  useEffect(() => {
    if (refreshInterval === 0) {
      console.log("refreshInterval is 0");
      return;
    }
    const interval = setInterval(() => setTime(new Date()), refreshInterval);
    return () => {
      console.log("clear refreshInterval");
      clearInterval(interval);
    };
  }, [refreshInterval]);

  // refresh every time the 'time' value changes
  useEffect(() => {
    console.log("reload page at", time.toISOString());
    props.refreshAction();
  }, [time]);

  const handleUserNameClick = (event: React.MouseEvent<HTMLElement>, index: number) => {
    switch (index) {
      case 0:
        navigation("/password");
        break;
      case 1:
        // setUser(null);
        navigation("/login");
        break;
      default:
        break;
    }
  };

  const refreshStrings = ["manual", "1s", "5s", "10s", "30s"];

  const handleRefreshClick = (event: React.MouseEvent<HTMLElement>, index: number) => {
    switch (index) {
      case 0: // manual
        setRefreshInterval(0);
        setRefreshString(refreshStrings.at(index)!);
        break;
      case 1: // 1s
        setRefreshInterval(1000);
        setRefreshString(refreshStrings.at(index)!);
        break;
      case 2: // 5s
        setRefreshInterval(5000);
        setRefreshString(refreshStrings.at(index)!);
        break;
      case 3: // 10s
        setRefreshInterval(10000);
        setRefreshString(refreshStrings.at(index)!);
        break;
      case 4: // 30s
        setRefreshInterval(30000);
        setRefreshString(refreshStrings.at(index)!);
        break;
      default:
        break;
    }
  };
  return (
    <ThemeProvider theme={mdTheme}>
      <Box sx={{ display: "flex" }}>
        <CssBaseline />
        <AppBar position="absolute" open={open}>
          <Toolbar
            sx={{
              pr: "24px", // keep right padding when drawer closed
            }}
          >
            <IconButton
              edge="start"
              color="inherit"
              aria-label="open drawer"
              onClick={toggleDrawer}
              sx={{
                marginRight: "36px",
                ...(open && { display: "none" }),
              }}
            >
              <MenuIcon />
            </IconButton>
            <Box sx={{ display: "flex", alignItems: "center", flexGrow: 1 }}>
              <Typography component="h1" variant="h6" color="inherit" noWrap>
                {props.title}
              </Typography>
              <Divider
                orientation="vertical"
                flexItem
                sx={{
                  //height: "100%",
                  //alignSelf: "center",
                  mx: 2,
                  borderColor: "white",
                }}
              />
              <SimpleListMenu
                title={`Refresh: ${refreshString}`}
                options={refreshStrings}
                handleMenuItemClick={handleRefreshClick}
              />
            </Box>
            <SimpleListMenu
              title={user?.username}
              options={["Change Password", "Logout"]}
              handleMenuItemClick={handleUserNameClick}
            />
          </Toolbar>
        </AppBar>
        <Drawer variant="permanent" open={open}>
          <Toolbar
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "flex-end",
              px: [1],
            }}
          >
            <IconButton onClick={toggleDrawer}>
              <ChevronLeftIcon />
            </IconButton>
          </Toolbar>
          <Divider />
          <List component="nav">
            <MainListItems />
            <Divider sx={{ my: 1 }} />
            {/* Moved to drop down menu */}
            {/* <Logout /> */}
          </List>
        </Drawer>
        <Box
          component="main"
          sx={{
            backgroundColor: (theme) =>
              theme.palette.mode === "light" ? theme.palette.grey[100] : theme.palette.grey[900],
            flexGrow: 1,
            height: "100vh",
            overflow: "auto",
          }}
        >
          <Toolbar />
          <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
            <Grid container spacing={3}>
              <Grid item xs={12}>
                <Paper sx={{ p: 2, display: "flex", flexDirection: "column" }}>
                  {props.children}
                </Paper>
              </Grid>
            </Grid>
          </Container>
        </Box>
      </Box>
    </ThemeProvider>
  );
}

export default Dashboard;
