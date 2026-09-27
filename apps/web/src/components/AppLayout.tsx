import { NavLink, Outlet } from "react-router-dom";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../features/auth/AuthProvider";

export function AppLayout() {
  const auth = useAuth();
  const navigate = useNavigate();
  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <NavLink className="font-semibold text-slate-950" to="/">
            Scholarship OS
          </NavLink>
          <nav className="flex gap-5 text-sm">
            <NavLink to="/">Dashboard</NavLink>
            <NavLink to="/profiles">Profiles</NavLink>
            <NavLink to="/applications">Applications</NavLink>
            <NavLink to="/research">Research</NavLink>
            <NavLink to="/writing">Writing</NavLink>
            <NavLink to="/information-requests">Requests</NavLink>
            <button
              type="button"
              onClick={() =>
                void auth
                  .logout()
                  .then(() => navigate("/login", { replace: true }))
              }
              className="text-slate-600"
            >
              Logout
            </button>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-10">
        <Outlet />
      </main>
    </div>
  );
}
