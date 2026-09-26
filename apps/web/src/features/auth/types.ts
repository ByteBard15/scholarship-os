export interface AuthUser {
  id: string;
  email: string;
  displayName?: string;
  isActive: boolean;
  lastLoginAt?: string;
  createdAt: string;
}

export interface LoginResult {
  accessToken: string;
  tokenType: "Bearer";
  expiresAt: string;
  user: AuthUser;
}

export interface Envelope<T> {
  data: T;
}
