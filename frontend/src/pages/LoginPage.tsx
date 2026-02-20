import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAppState } from "../state/AppState";

export function LoginPage() {
  const navigate = useNavigate();
  const { setSession } = useAppState();
  const [username, setUsername] = useState("demo-user");
  const [password, setPassword] = useState("123456");
  const [error, setError] = useState("");

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!username.trim() || !password.trim()) {
      setError("用户名和密码不能为空");
      return;
    }
    setSession({
      username: username.trim(),
      token: `mock-${Date.now()}`
    });
    navigate("/app/dashboard", { replace: true });
  };

  return (
    <div className="auth-layout">
      <section className="card auth-card">
        <h2>Ticketing 登录</h2>
        <p>当前为前端演示登录（Mock Session），用于流程联调。</p>
        <form className="form" onSubmit={onSubmit}>
          <label>
            用户名
            <input value={username} onChange={(e) => setUsername(e.target.value)} />
          </label>
          <label>
            密码
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </label>
          {error ? <p className="error-text">{error}</p> : null}
          <button type="submit">登录进入控制台</button>
        </form>
      </section>
    </div>
  );
}


