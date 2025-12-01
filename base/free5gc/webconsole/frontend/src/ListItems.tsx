import React, { useContext } from "react";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import BarChartIcon from "@mui/icons-material/BarChart";
import PhoneAndroid from "@mui/icons-material/PhoneAndroid";
import FontDownload from "@mui/icons-material/FontDownload";
import SupervisorAccountOutlinedIcon from "@mui/icons-material/SupervisorAccountOutlined";
import AttachMoneyOutlinedIcon from "@mui/icons-material/AttachMoneyOutlined";
import PersonIcon from "@mui/icons-material/Person";

import { Link } from "react-router-dom";
import { LoginContext } from "./LoginContext";
import { config } from "./constants/config";

export const MainListItems = () => {
  const { user } = useContext(LoginContext);

  const isAdmin = () => {
    if (config.enableCognitoAuth) {
      return true;
    }
    if (user !== null && user.username === "admin") {
      return true;
    }
    return false;
  };

  return (
    <React.Fragment>
      <Link to="/status" style={{ color: "inherit", textDecoration: "inherit" }}>
        <ListItemButton>
          <ListItemIcon>
            <BarChartIcon />
          </ListItemIcon>
          <ListItemText primary="REALTIME STATUS" />
        </ListItemButton>
      </Link>
      <Link to="/subscriber" style={{ color: "inherit", textDecoration: "inherit" }}>
        <ListItemButton>
          <ListItemIcon>
            <PhoneAndroid />
          </ListItemIcon>
          <ListItemText primary="SUBSCRIBERS" />
        </ListItemButton>
      </Link>
      <Link to="/profile" style={{ color: "inherit", textDecoration: "inherit" }}>
        <ListItemButton>
          <ListItemIcon>
            <PersonIcon />
          </ListItemIcon>
          <ListItemText primary="PROFILE" />
        </ListItemButton>
      </Link>
      <Link to="/analysis" style={{ color: "inherit", textDecoration: "inherit" }}>
        <ListItemButton>
          <ListItemIcon>
            <FontDownload />
          </ListItemIcon>
          <ListItemText primary="ANALYSIS" />
        </ListItemButton>
      </Link>
      {isAdmin() ? (
        <Link to="/tenant" style={{ color: "inherit", textDecoration: "inherit" }}>
          <ListItemButton>
            <ListItemIcon>
              <SupervisorAccountOutlinedIcon />
            </ListItemIcon>
            <ListItemText primary="TENANT AND USER" />
          </ListItemButton>
        </Link>
      ) : (
        <div />
      )}
      <Link to="/charging" style={{ color: "inherit", textDecoration: "inherit" }}>
        <ListItemButton>
          <ListItemIcon>
            <AttachMoneyOutlinedIcon />
          </ListItemIcon>
          <ListItemText primary="UE CHARGING" />
        </ListItemButton>
      </Link>
    </React.Fragment>
  );
};
