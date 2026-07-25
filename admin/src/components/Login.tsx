import { useState, FormEvent, useEffect } from 'react';
import { authService } from '../services/auth';
import './Login.css';

interface LoginProps {
  onLogin: () => void;
}

const SAVED_IDENTIFIER_KEY = 'admin_saved_identifier';
const REMEMBER_ME_KEY = 'admin_remember_me';

export const Login = ({ onLogin }: LoginProps) => {
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const savedIdentifier = localStorage.getItem(SAVED_IDENTIFIER_KEY);
    const savedRememberMe = localStorage.getItem(REMEMBER_ME_KEY) === 'true';

    if (savedIdentifier) {
      setIdentifier(savedIdentifier);
    }
    if (savedRememberMe) {
      setRememberMe(true);
    }
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await authService.login(identifier.trim(), password);

      if (rememberMe) {
        localStorage.setItem(SAVED_IDENTIFIER_KEY, identifier.trim());
        localStorage.setItem(REMEMBER_ME_KEY, 'true');
      } else {
        localStorage.removeItem(SAVED_IDENTIFIER_KEY);
        localStorage.removeItem(REMEMBER_ME_KEY);
      }

      onLogin();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка входа');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <h1>Workouts Admin</h1>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="identifier">Email или username</label>
            <input
              id="identifier"
              type="text"
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              required
              disabled={loading}
              autoComplete="username"
            />
          </div>
          <div className="form-group">
            <label htmlFor="password">Пароль</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={loading}
              autoComplete="current-password"
            />
          </div>
          <div className="form-group remember-me">
            <label className="remember-me-label">
              <input
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                disabled={loading}
              />
              <span>Запомнить логин</span>
            </label>
          </div>
          {error && <div className="error-message">{error}</div>}
          <button type="submit" disabled={loading} className="login-button">
            {loading ? 'Вход...' : 'Войти'}
          </button>
        </form>
      </div>
    </div>
  );
};
