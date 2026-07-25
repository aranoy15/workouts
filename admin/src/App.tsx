import { useState, useEffect } from 'react';
import { authService } from './services/auth';
import { Login } from './components/Login';
import { Header } from './components/Header';
import { ExercisesList } from './components/ExercisesList';
import { ExerciseForm } from './components/ExerciseForm';
import { UsersList } from './components/UsersList';
import './App.css';

export type View = 'exercises' | 'add-exercise' | 'users';

export const App = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [currentView, setCurrentView] = useState<View>('exercises');
  const [loading, setLoading] = useState(true);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    const authState = authService.getAuthState();
    setIsAuthenticated(authState.isAuthenticated);
    setLoading(false);
  }, []);

  const handleLogin = () => {
    setIsAuthenticated(true);
    setCurrentView('exercises');
  };

  const handleLogout = () => {
    authService.logout();
    setIsAuthenticated(false);
  };

  const handleExerciseCreated = () => {
    setCurrentView('exercises');
    setRefreshKey((prev) => prev + 1);
  };

  if (loading) {
    return <div className="loading-screen">Загрузка...</div>;
  }

  if (!isAuthenticated) {
    return <Login onLogin={handleLogin} />;
  }

  return (
    <div className="app">
      <Header onLogout={handleLogout} currentView={currentView} onViewChange={setCurrentView} />
      <main className="main-content">
        {currentView === 'exercises' && (
          <ExercisesList key={refreshKey} onAddExercise={() => setCurrentView('add-exercise')} />
        )}
        {currentView === 'add-exercise' && (
          <ExerciseForm
            onSuccess={handleExerciseCreated}
            onCancel={() => setCurrentView('exercises')}
          />
        )}
        {currentView === 'users' && <UsersList />}
      </main>
    </div>
  );
};
