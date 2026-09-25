import { NavLink, Outlet } from "react-router-dom";

export function AppLayout() {
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
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-10">
        <Outlet />
      </main>
    </div>
  );
}
