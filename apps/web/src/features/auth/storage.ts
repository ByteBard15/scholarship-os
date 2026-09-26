const tokenKey = "scholarship-os.access-token";

export const authStorage = {
  get: () => sessionStorage.getItem(tokenKey),
  set: (token: string) => sessionStorage.setItem(tokenKey, token),
  clear: () => sessionStorage.removeItem(tokenKey),
};

export const unauthorizedEvent = "scholarship-os:unauthorized";
