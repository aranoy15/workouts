import { useEffect, useState } from 'react';
import { CreateUserInput, UpdateUserInput, User, UserRole } from '../types';
import { userService } from '../services/users';
import './UsersList.css';

export const UsersList = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [creatingUser, setCreatingUser] = useState(false);
  const [editForm, setEditForm] = useState<UpdateUserInput>({});
  const [createForm, setCreateForm] = useState<CreateUserInput>({
    username: '',
    email: '',
    password: '',
    role: 'user',
  });

  const loadUsers = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await userService.getAll();
      setUsers(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки пользователей');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
    const interval = setInterval(loadUsers, 30000);
    return () => clearInterval(interval);
  }, []);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const handleEdit = (user: User) => {
    setEditingUser(user);
    setEditForm({
      username: user.username,
      email: user.email,
      role: user.role,
      is_active: user.is_active,
      password: '',
    });
  };

  const handleCreate = async () => {
    if (!createForm.username || !createForm.password) {
      setError('Имя пользователя и пароль обязательны');
      return;
    }

    if (createForm.password.length < 6) {
      setError('Пароль должен содержать минимум 6 символов');
      return;
    }

    setError('');
    try {
      await userService.create(createForm);
      setCreatingUser(false);
      setCreateForm({ username: '', email: '', password: '', role: 'user' });
      await loadUsers();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка создания пользователя');
    }
  };

  const handleSaveEdit = async () => {
    if (!editingUser) return;

    if (editForm.password && editForm.password.length < 6) {
      setError('Пароль должен содержать минимум 6 символов');
      return;
    }

    const updateData: UpdateUserInput = { ...editForm };
    if (!updateData.password || updateData.password.trim() === '') {
      delete updateData.password;
    }

    setError('');
    try {
      await userService.update(editingUser.id, updateData);
      setEditingUser(null);
      setEditForm({});
      await loadUsers();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка обновления пользователя');
    }
  };

  return (
    <div className="users-list-container">
      <div className="users-header">
        <h1>Управление пользователями</h1>
        <div className="header-actions">
          <button onClick={() => setCreatingUser(true)} className="create-button">
            Создать пользователя
          </button>
          <button onClick={loadUsers} className="refresh-button">
            Обновить
          </button>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      {loading ? (
        <div className="loading">Загрузка...</div>
      ) : users.length === 0 ? (
        <div className="empty-state">Пользователи не найдены</div>
      ) : (
        <div className="users-table-container">
          <table className="users-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Имя пользователя</th>
                <th>Email</th>
                <th>Роль</th>
                <th>Дата создания</th>
                <th>Статус</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr
                  key={user.id}
                  className={selectedUser?.id === user.id ? 'selected' : ''}
                  onClick={() => setSelectedUser(user.id === selectedUser?.id ? null : user)}
                >
                  <td className="user-id">{user.id.substring(0, 8)}...</td>
                  <td>{user.username}</td>
                  <td>{user.email || '—'}</td>
                  <td>
                    <span className={`role-badge ${user.role}`}>
                      {user.role === 'admin' ? 'Администратор' : 'Пользователь'}
                    </span>
                  </td>
                  <td>{formatDate(user.created_at)}</td>
                  <td>
                    {user.is_active ? (
                      <span className="status-badge active">Активен</span>
                    ) : (
                      <span className="status-badge deleted">Неактивен</span>
                    )}
                  </td>
                  <td onClick={(e) => e.stopPropagation()}>
                    <div className="user-actions">
                      <button
                        onClick={() => handleEdit(user)}
                        className="action-button edit-button"
                        title="Редактировать"
                      >
                        Edit
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {selectedUser && (
        <div className="user-details-overlay" onClick={() => setSelectedUser(null)}>
          <div className="user-details" onClick={(e) => e.stopPropagation()}>
            <div className="user-details-header">
              <h2>Детали пользователя</h2>
              <button onClick={() => setSelectedUser(null)} className="close-button" title="Закрыть">
                ×
              </button>
            </div>
            <div className="user-details-content">
              <div className="detail-row">
                <span className="detail-label">ID:</span>
                <span className="detail-value">{selectedUser.id}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Имя пользователя:</span>
                <span className="detail-value">{selectedUser.username}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Email:</span>
                <span className="detail-value">{selectedUser.email || '—'}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Роль:</span>
                <span className={`role-badge ${selectedUser.role}`}>
                  {selectedUser.role === 'admin' ? 'Администратор' : 'Пользователь'}
                </span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Создан:</span>
                <span className="detail-value">{formatDate(selectedUser.created_at)}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Обновлён:</span>
                <span className="detail-value">{formatDate(selectedUser.updated_at)}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Статус:</span>
                {selectedUser.is_active ? (
                  <span className="status-badge active">Активен</span>
                ) : (
                  <span className="status-badge deleted">Неактивен</span>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {creatingUser && (
        <div
          className="user-edit-overlay"
          onClick={() => {
            setCreatingUser(false);
            setCreateForm({ username: '', email: '', password: '', role: 'user' });
            setError('');
          }}
        >
          <div className="user-edit-container" onClick={(e) => e.stopPropagation()}>
            <div className="user-edit-header">
              <h2>Создать пользователя</h2>
              <button
                onClick={() => {
                  setCreatingUser(false);
                  setCreateForm({ username: '', email: '', password: '', role: 'user' });
                  setError('');
                }}
                className="close-button"
                title="Закрыть"
              >
                ×
              </button>
            </div>
            <div className="user-edit-content">
              <div className="form-group">
                <label htmlFor="create-username">Имя пользователя *</label>
                <input
                  id="create-username"
                  type="text"
                  value={createForm.username}
                  onChange={(e) => setCreateForm({ ...createForm, username: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label htmlFor="create-email">Email</label>
                <input
                  id="create-email"
                  type="email"
                  value={createForm.email || ''}
                  onChange={(e) => setCreateForm({ ...createForm, email: e.target.value })}
                />
              </div>
              <div className="form-group">
                <label htmlFor="create-password">Пароль *</label>
                <input
                  id="create-password"
                  type="password"
                  value={createForm.password}
                  onChange={(e) => setCreateForm({ ...createForm, password: e.target.value })}
                  required
                  minLength={6}
                />
              </div>
              <div className="form-group">
                <label htmlFor="create-role">Роль *</label>
                <select
                  id="create-role"
                  value={createForm.role || 'user'}
                  onChange={(e) =>
                    setCreateForm({ ...createForm, role: e.target.value as UserRole })
                  }
                >
                  <option value="user">Пользователь</option>
                  <option value="admin">Администратор</option>
                </select>
              </div>
              {error && <div className="error-message">{error}</div>}
              <div className="form-actions">
                <button
                  onClick={() => {
                    setCreatingUser(false);
                    setCreateForm({ username: '', email: '', password: '', role: 'user' });
                    setError('');
                  }}
                  className="cancel-button"
                >
                  Отмена
                </button>
                <button onClick={handleCreate} className="save-button">
                  Создать
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {editingUser && (
        <div
          className="user-edit-overlay"
          onClick={() => {
            setEditingUser(null);
            setEditForm({});
          }}
        >
          <div className="user-edit-container" onClick={(e) => e.stopPropagation()}>
            <div className="user-edit-header">
              <h2>Редактировать пользователя</h2>
              <button
                onClick={() => {
                  setEditingUser(null);
                  setEditForm({});
                }}
                className="close-button"
                title="Закрыть"
              >
                ×
              </button>
            </div>
            <div className="user-edit-content">
              <div className="form-group">
                <label htmlFor="edit-username">Имя пользователя *</label>
                <input
                  id="edit-username"
                  type="text"
                  value={editForm.username || ''}
                  onChange={(e) => setEditForm({ ...editForm, username: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label htmlFor="edit-email">Email</label>
                <input
                  id="edit-email"
                  type="email"
                  value={editForm.email || ''}
                  onChange={(e) => setEditForm({ ...editForm, email: e.target.value })}
                />
              </div>
              <div className="form-group">
                <label htmlFor="edit-password">Пароль (оставьте пустым, чтобы не менять)</label>
                <input
                  id="edit-password"
                  type="password"
                  value={editForm.password || ''}
                  onChange={(e) => setEditForm({ ...editForm, password: e.target.value })}
                  minLength={6}
                />
              </div>
              <div className="form-group">
                <label htmlFor="edit-role">Роль *</label>
                <select
                  id="edit-role"
                  value={editForm.role || 'user'}
                  onChange={(e) => setEditForm({ ...editForm, role: e.target.value as UserRole })}
                >
                  <option value="user">Пользователь</option>
                  <option value="admin">Администратор</option>
                </select>
              </div>
              <div className="form-group">
                <label htmlFor="edit-active">Статус</label>
                <select
                  id="edit-active"
                  value={editForm.is_active === false ? 'inactive' : 'active'}
                  onChange={(e) =>
                    setEditForm({ ...editForm, is_active: e.target.value === 'active' })
                  }
                >
                  <option value="active">Активен</option>
                  <option value="inactive">Неактивен</option>
                </select>
              </div>
              {error && <div className="error-message">{error}</div>}
              <div className="form-actions">
                <button
                  onClick={() => {
                    setEditingUser(null);
                    setEditForm({});
                  }}
                  className="cancel-button"
                >
                  Отмена
                </button>
                <button onClick={handleSaveEdit} className="save-button">
                  Сохранить
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
