import { useNavigate } from "react-router-dom";
import { useAppDispatch } from "../../../app/hooks";
import { setTeacherToken } from "../../auth/authSlice";

export default function LoginPage() {
  const nav = useNavigate();
  const dispatch = useAppDispatch();

  return (
    <div style={{ background: "var(--card)", borderRadius: "var(--radius)", padding: 20 }}>
      <h1 style={{ marginTop: 0 }}>Teacher Login</h1>
      <p style={{ color: "var(--muted)" }}>Пока заглушка: нажми — и попадёшь в кабинет.</p>

      <button
        onClick={() => {
          dispatch(setTeacherToken("stub-token"));
          nav("/teacher");
        }}
        style={{
          padding: "10px 14px",
          borderRadius: 12,
          border: 0,
          background: "var(--accent)",
          color: "white",
          fontWeight: 800,
          cursor: "pointer",
        }}
      >
        Login (stub)
      </button>
    </div>
  );
}
