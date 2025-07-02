import axios from "axios";

// Dynamically set the base URL based on the environment
const API_BASE = "api";
const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
});

// Request interceptor with logging
api.interceptors.request.use(
  (config) => {
    console.log("Making request to:", config.url); // Log to confirm request is triggered
    return config;
  },
  (error) => {
    console.error("Request error:", error);
    return Promise.reject(error);
  },
);

// Response interceptor with enhanced logging
// api.interceptors.response.use(
//   (response) => {
//     return response;
//   },
//   (error) => {
//     console.error("Response error:", error.message);
//     if (error.response) {
//       console.error("Status:", error.response.status);
//       console.error("Data:", error.response.data);
//       console.error("URL:", error.config.url);
//     } else if (error.request) {
//       console.error("No response received:", error.request);
//     }
//     return Promise.reject(error);
//   },
// );

export { api };
