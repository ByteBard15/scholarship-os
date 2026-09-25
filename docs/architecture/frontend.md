# Frontend architecture

The Vite application uses React Router for page routing and TanStack Query for server state. `src/api` contains the HTTP abstraction, while `src/features/profile` contains profile types, endpoint functions, query keys, and hooks. Pages compose those feature APIs; shared layout belongs in `src/components`.

The current pages are a dashboard placeholder, a profile list, and a full-profile overview. There are intentionally no CRUD forms or client-side global state store. `VITE_API_BASE_URL` selects the API origin, and `VITE_DEMO_USER_ID` provides a development-only account selector until real authentication is introduced.
