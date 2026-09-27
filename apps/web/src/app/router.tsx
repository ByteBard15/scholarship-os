import { createBrowserRouter } from "react-router-dom";
import { AppLayout } from "../components/AppLayout";
import { DashboardPage } from "../pages/DashboardPage";
import { ImportReviewPage } from "../pages/ImportReviewPage";
import { ProfileDetailPage } from "../pages/ProfileDetailPage";
import { ProfilesPage } from "../pages/ProfilesPage";
import { ProfileComparisonPage } from "../pages/ProfileComparisonPage";
import { ApplicationsPage } from "../pages/ApplicationsPage";
import { ApplicationDetailPage } from "../pages/ApplicationDetailPage";
import { InformationRequestsPage } from "../pages/InformationRequestsPage";
import { QuestionnairePage } from "../pages/QuestionnairePage";
import { ResearchDetailPage } from "../pages/ResearchDetailPage";
import { ResearchPage } from "../pages/ResearchPage";
import { LoginPage } from "../pages/LoginPage";
import { ProtectedRoute } from "../components/ProtectedRoute";
import { WritingLibraryPage } from "../pages/WritingLibraryPage";
import { RouteErrorPage } from "../components/RouteErrorPage";
export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage />, errorElement: <RouteErrorPage /> },
  {
    element: <ProtectedRoute />,
    errorElement: <RouteErrorPage />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: "/", element: <DashboardPage /> },
          { path: "/profiles", element: <ProfilesPage /> },
          { path: "/applications", element: <ApplicationsPage /> },
          { path: "/applications/:id", element: <ApplicationDetailPage /> },
          {
            path: "/applications/:applicationId/questionnaires/:questionnaireId",
            element: <QuestionnairePage />,
          },
          { path: "/research", element: <ResearchPage /> },
          { path: "/research/:id", element: <ResearchDetailPage /> },
          { path: "/writing", element: <WritingLibraryPage /> },
          {
            path: "/information-requests",
            element: <InformationRequestsPage />,
          },
          { path: "/profiles/:id", element: <ProfileDetailPage /> },
          {
            path: "/profiles/:id/imports/:importId",
            element: <ImportReviewPage />,
          },
          {
            path: "/profiles/:id/compare/:otherId",
            element: <ProfileComparisonPage />,
          },
        ],
      },
    ],
  },
]);
