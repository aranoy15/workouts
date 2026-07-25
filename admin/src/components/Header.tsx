import { authService } from '../services/auth';
import { View } from '../App';
import './Header.css';

interface HeaderProps {
  onLogout: () => void;
  currentView?: View;
  onViewChange?: (view: View) => void;
}

export const Header = ({ onLogout, currentView, onViewChange }: HeaderProps) => {
  const authState = authService.getAuthState();
  const user = authState.user;

  return (
    <header className="admin-header">
      <div className="header-content">
        <h1>Workouts Admin</h1>
        <nav className="header-nav">
          <button
            onClick={() => onViewChange?.('exercises')}
            className={`nav-button ${currentView === 'exercises' || currentView === 'add-exercise' ? 'active' : ''}`}
          >
            Упражнения
          </button>
          <button
            onClick={() => onViewChange?.('users')}
            className={`nav-button ${currentView === 'users' ? 'active' : ''}`}
          >
            Пользователи
          </button>
        </nav>
        <div className="header-user">
          <span className="user-info">
            {user?.username}
            {user?.email ? ` (${user.email})` : ''}
          </span>
          <button onClick={onLogout} className="logout-button">
            Выйти
          </button>
        </div>
      </div>
    </header>
  );
};
