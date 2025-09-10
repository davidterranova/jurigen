import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { ApiResponse, ApiError } from '../types/api';

// Create axios instance with base configuration
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
    timeout: 10000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Request interceptor for adding auth headers if needed
  client.interceptors.request.use(
    (config) => {
      // Add auth token if available
      const token = localStorage.getItem('authToken');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    },
    (error) => {
      return Promise.reject(error);
    }
  );

  // Response interceptor for handling common errors
  client.interceptors.response.use(
    (response: AxiosResponse) => response,
    (error) => {
      const apiError: ApiError = {
        message: error.response?.data?.message || error.message || 'An error occurred',
        status: error.response?.status || 500,
        details: error.response?.data || {},
      };

      // Handle specific error cases
      if (error.response?.status === 401) {
        // Handle unauthorized - clear token and redirect to login if needed
        localStorage.removeItem('authToken');
      }

      return Promise.reject(apiError);
    }
  );

  return client;
};

export const apiClient = createApiClient();

// Generic API response handler
export const handleApiResponse = <T>(response: AxiosResponse): ApiResponse<T> => {
  return {
    data: response.data,
    success: response.status >= 200 && response.status < 300,
    message: response.data?.message,
  };
};

// API utility functions
export const api = {
  get: async <T>(url: string, params?: Record<string, unknown>): Promise<ApiResponse<T>> => {
    const response = await apiClient.get(url, { params });
    return handleApiResponse<T>(response);
  },

  post: async <T>(url: string, data?: unknown): Promise<ApiResponse<T>> => {
    const response = await apiClient.post(url, data);
    return handleApiResponse<T>(response);
  },

  put: async <T>(url: string, data?: unknown): Promise<ApiResponse<T>> => {
    const response = await apiClient.put(url, data);
    return handleApiResponse<T>(response);
  },

  patch: async <T>(url: string, data?: unknown): Promise<ApiResponse<T>> => {
    const response = await apiClient.patch(url, data);
    return handleApiResponse<T>(response);
  },

  delete: async <T>(url: string): Promise<ApiResponse<T>> => {
    const response = await apiClient.delete(url);
    return handleApiResponse<T>(response);
  },
};
