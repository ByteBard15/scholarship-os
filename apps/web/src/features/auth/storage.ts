const tokenKey = "scholarship-os.access-token";

export const authStorage = {
  key: tokenKey,
  get: () => {
    const sharedToken = localStorage.getItem(tokenKey);
    if (sharedToken) return sharedToken;

    // Preserve existing sessions after switching from per-tab storage. New tabs
    // can then restore the same authenticated session from localStorage.
    const legacyToken = sessionStorage.getItem(tokenKey);
    if (legacyToken) {
      localStorage.setItem(tokenKey, legacyToken);
      sessionStorage.removeItem(tokenKey);
    }
    return legacyToken;
  },
  set: (token: string) => {
    localStorage.setItem(tokenKey, token);
    sessionStorage.removeItem(tokenKey);
  },
  clear: () => {
    localStorage.removeItem(tokenKey);
    sessionStorage.removeItem(tokenKey);
  },
};

export const unauthorizedEvent = "scholarship-os:unauthorized";
