import { Outlet, Link, useLocation, useNavigate } from "react-router-dom";
import { useState } from "react";

function NavLink({ to, label }: { to: string; label: string }) {
  const { pathname } = useLocation();
  const active = pathname === to || (to !== "/" && pathname.startsWith(to));

  return (
    <Link
      to={to}
      style={{
        padding: "8px 10px",
        borderRadius: 10,
        background: active ? "var(--accent-soft)" : "transparent",
        color: active ? "var(--accent-ink)" : "inherit",
        border: active ? "1px solid rgba(124, 92, 255, 0.22)" : "1px solid transparent",
      }}
    >
      {label}
    </Link>
  );
}

export default function App() {
  const nav = useNavigate();
  const [pin, setPin] = useState("");

  function goJoin() {
    const cleaned = pin.trim();
    if (!cleaned) return;
    nav(`/join?pin=${encodeURIComponent(cleaned)}`);
  }

  return (
    <div className="container">
      <header
        className="card card-pad"
        style={{
          display: "flex",
          justifyContent: "space-between",
          gap: 16,
          alignItems: "center",
          marginBottom: 14,
        }}
      >
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <div style={{ fontWeight: 900, fontSize: 18 }}>Quiz Platform</div>
            <span
              style={{
                fontSize: 12,
                padding: "4px 8px",
                borderRadius: 999,
                border: "1px solid var(--border)",
                color: "var(--muted)",
              }}
            >
              MVP
            </span>
          </div>
          <div className="small">Интерактивные квизы и опросы для лекций и семинаров</div>
        </div>

        <div style={{ display: "flex", gap: 10, alignItems: "center", flexWrap: "wrap" }}>
          {/* Join by PIN */}
          <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <input
              className="input"
              placeholder="Код комнаты (PIN)"
              value={pin}
              onChange={(e) => setPin(e.target.value)}
              style={{ width: 180 }}
            />
            <button className="btn btn-primary" onClick={goJoin}>
              Войти
            </button>
          </div>

          <Link to="/login" className="btn btn-soft" style={{ display: "inline-flex", alignItems: "center" }}>
            Вход преподавателя
          </Link>

          <nav className="nav">
            <NavLink to="/" label="Главная" />
            <NavLink to="/teacher" label="Кабинет" />
          </nav>
        </div>
      </header>

      <Outlet />
    </div>
  );
}
