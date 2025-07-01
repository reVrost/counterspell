import axios from "axios";

const api = axios.create({
  baseURL: "/api",
  timeout: 30000,
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // Handle common errors
    if (error.response?.status === 404) {
      console.error("Resource not found:", error.config?.url);
    } else if (error.response?.status >= 500) {
      console.error("Server error:", error.response?.data);
    }
    return Promise.reject(error);
  },
);

export { api };
