import axios, { AxiosResponse } from "axios";
import type { LoginRequest } from "@/models";

export const checkAuthStatusRequest = axios.get("/auth/status", { withCredentials: true });

export function loginRequest(loginCreds: LoginRequest): Promise<AxiosResponse> {
    return axios.post("/users/login", loginCreds, {withCredentials: true})
}