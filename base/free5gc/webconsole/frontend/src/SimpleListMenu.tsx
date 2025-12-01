import React, { useContext, useState } from "react";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import MenuItem from "@mui/material/MenuItem";
import Menu from "@mui/material/Menu";
import { useNavigate } from "react-router-dom";
import { LoginContext } from "./LoginContext";

export interface SimpleListMenuProps {
  title: string | undefined;
  options: string[];
  handleMenuItemClick: (event: React.MouseEvent<HTMLElement>, index: number) => void;
}

export default function SimpleListMenu(props: SimpleListMenuProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const open = Boolean(anchorEl);
  const handleClickListItem = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const navigation = useNavigate();
  const context = useContext(LoginContext);
  if (context === undefined) {
    throw new Error("LoginContext must be used within a LoginContext.Provider");
  }
  const { setUser } = context;

  function onChangePassword() {
    navigation("/password");
  }

  function onLogout() {
    setUser(null);
    navigation("/login");
  }

  const handleMenuItemClick = (event: React.MouseEvent<HTMLElement>, index: number) => {
    setSelectedIndex(index);
    setAnchorEl(null);
    switch (index) {
      case 0:
        onChangePassword();
        break;
      case 1:
        onLogout();
        break;
      default:
        break;
    }
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <div>
      <List component="nav" aria-label="Device settings">
        <ListItem
          button
          id="lock-button"
          aria-haspopup="listbox"
          aria-controls="lock-menu"
          aria-label="when device is locked"
          aria-expanded={open ? "true" : undefined}
          onClick={handleClickListItem}
        >
          <ListItemText primaryTypographyProps={{ fontSize: "18px" }} primary={props.title} />
        </ListItem>
      </List>
      <Menu
        id="lock-menu"
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        MenuListProps={{
          "aria-labelledby": "lock-button",
          role: "listbox",
        }}
      >
        {props.options.map((option, index) => (
          <MenuItem
            key={option}
            selected={index === selectedIndex}
            onClick={(event) => {
              setSelectedIndex(index);
              setAnchorEl(null);
              props.handleMenuItemClick(event, index);
            }}
          >
            {option}
          </MenuItem>
        ))}
      </Menu>
    </div>
  );
}
