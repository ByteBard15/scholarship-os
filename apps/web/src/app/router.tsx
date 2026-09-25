import { createBrowserRouter } from "react-router-dom";
import { AppLayout } from "../components/AppLayout";
import { DashboardPage } from "../pages/DashboardPage";
import { ProfileDetailPage } from "../pages/ProfileDetailPage";
import { ProfilesPage } from "../pages/ProfilesPage";

export const router = createBrowserRouter([
  {
    element: <AppLayout />,
    children: [
      { path: "/", element: <DashboardPage /> },
      { path: "/profiles", element: <ProfilesPage /> },
      { path: "/profiles/:id", element: <ProfileDetailPage /> },
    ],
  },
]);
