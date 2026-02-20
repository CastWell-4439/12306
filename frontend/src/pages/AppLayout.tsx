import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAppState } from "../state/AppState";

const navItems = [
  { to: "/app/dashboard", label: "总览" },
  { to: "/app/booking", label: "购票流程" },
  { to: "/app/inventory", label: "库存操作" },
  { to: "/app/orders", label: "订单查询" }
];

export function AppLayout() {
  const navigate = useNavigate();
  const { session, setSession } = useAppState();

  return (
    <div className="layout">
      <header className="header">
        <h1>Ticketing Frontend Console</h1>
        <p>业务流程、系统检查与请求排障一体化控制台。</p>
      </header>

      <div className="topbar card">
        <nav className="nav-tabs">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => (isActive ? "tab active" : "tab")}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="session-meta">
          <span>当前用户：{session?.username ?? "-"}</span>
          <button
            className="btn-secondary"
            onClick={() => {
              setSession(null);
              navigate("/login");
            }}
          >
            退出登录
          </button>
        </div>
      </div>

      <Outlet />
    </div>
  );
}


