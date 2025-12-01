import { useContext } from "react";
import { Navigate } from "react-router-dom";

import { LoginContext } from "./LoginContext";

export const ProtectedRoute = (props: any) => {
  const context = useContext(LoginContext);
  if (context === undefined) {
    throw new Error("LoginContext must be used within a LoginContext.Provider");
  }
  const { user } = context;

  if (user === null) {
    return <Navigate to="/login" />;
  }
  return props.children;
};
