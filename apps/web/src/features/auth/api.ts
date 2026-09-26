import { getJSON, requestJSON } from "../../api/client";
import type { AuthUser, Envelope, LoginResult } from "./types";

export const loginRequest = (email: string, password: string) =>
  requestJSON<Envelope<LoginResult>>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });

export const meRequest = () => getJSON<Envelope<AuthUser>>("/api/v1/auth/me");

export const logoutRequest = () =>
  requestJSON<void>("/api/v1/auth/logout", { method: "POST" });
