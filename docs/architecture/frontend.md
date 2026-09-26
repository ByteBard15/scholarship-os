# Frontend architecture

The Vite application uses React Router for page routing and TanStack Query for server state. `src/api` contains the HTTP abstraction, while `src/features/profile` contains profile types, endpoint functions, query keys, and hooks. Pages compose those feature APIs; shared layout belongs in `src/components`.

Feature packages exist for authentication, profiles, applications, and controlled workflows. Pages include login, the operational dashboard, hierarchy-aware profile management, reviewed imports and overrides, application list/creation, and application detail tabs for requirements, deadlines, funding, tasks, research review, contacts, supervisors, profile lineage, and links. There is intentionally no client-side global state store. `VITE_API_BASE_URL` selects the API origin. The authenticated user is restored through `/auth/me`; no user ID is accepted from frontend configuration.
