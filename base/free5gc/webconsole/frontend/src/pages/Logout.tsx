import React, { useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import PhoneAndroid from "@mui/icons-material/PhoneAndroid";

import { LoginContext } from "../LoginContext";

export const Logout = () => {
  const navigation = useNavigate();
  const context = useContext(LoginContext);
  if (context === undefined) {
    throw new Error("LoginContext must be used within a LoginContext.Provider");
  }
  const { setUser } = context;

  function onLogout() {
    setUser(null);
    navigation("/login");
  }

  return (
    <React.Fragment>
      <ListItemButton onClick={() => onLogout()}>
        <ListItemIcon>
          <PhoneAndroid />
        </ListItemIcon>
        <ListItemText primary="LOGOUT" />
      </ListItemButton>
    </React.Fragment>
  );
};
