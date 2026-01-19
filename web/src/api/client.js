import axios from "axios";

export const api = axios.create({
    baseURL: "/api/v1",
    headers: {
        "Content-Type": "application/json",
    },
});

// Add interceptor to inject API Key
api.interceptors.request.use((config) => {
    config.headers["X-App-Key"] = "netengine_secret_key_123";
    return config;
});

// API Helper Functions
export const syncRouter = (id) => api.post(`/sync/${id}`);
export const isolateUser = (data, sync = true) => api.post(`/isolate${sync ? '?sync=true' : ''}`, data);
export const createSecret = (data) => api.post(`/secret`, data);
export const backupRouter = (id) => api.post(`/router/${id}/backup`);
export const updatePlan = (user, data) => api.put(`/secret/${user}`, data);
