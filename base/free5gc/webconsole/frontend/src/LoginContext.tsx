import { Dispatch, SetStateAction, createContext } from "react";

export interface User {
  username: string;
  token: string;
}

export interface UserContext {
  user: User | null;
  setUser: Dispatch<SetStateAction<User | null>>;
}

export const LoginContext = createContext<UserContext | undefined>(undefined);
