import axios from "axios";

const apiConfig = {
  API_URL: "",
};

if (process.env.NODE_ENV === "development") {
  apiConfig.API_URL = "http://127.0.0.1:5000";
} else {
  apiConfig.API_URL = process.env.REACT_APP_HTTP_API_URL ? process.env.REACT_APP_HTTP_API_URL : "";
}

const instance = axios.create({
  baseURL: apiConfig.API_URL,
});

// attach the token to every request
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) {
      // add the token to the header
      console.log("adding token to axios header");
      config.headers.Token = `${token}`;
    } else {
      console.warn("no token in local storage!");
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

// Handle Unauthorized Error(401)
instance.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      console.warn("Unauthorized error, clearing token");
      localStorage.removeItem("token");
      window.location.href = "/"; // Redirect to index page
    }
    return Promise.reject(error);
  },
);

export default instance;
