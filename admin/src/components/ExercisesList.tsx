import { useEffect, useState, useCallback, useRef } from 'react';
import { Exercise } from '../types';
import { exerciseService } from '../services/exercises';
import { ExerciseForm } from './ExerciseForm';
import './ExercisesList.css';

interface ExercisesListProps {
  onAddExercise: () => void;
}

export const ExercisesList = ({ onAddExercise }: ExercisesListProps) => {
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [editingExercise, setEditingExercise] = useState<Exercise | null>(null);

  const loadExercises = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await exerciseService.getAll();
      setExercises(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки упражнений');
    } finally {
      setLoading(false);
    }
  }, []);

  const loadExercisesRef = useRef(loadExercises);
  useEffect(() => {
    loadExercisesRef.current = loadExercises;
  }, [loadExercises]);

  useEffect(() => {
    loadExercises();
  }, [loadExercises]);

  useEffect(() => {
    const interval = setInterval(() => {
      loadExercisesRef.current();
    }, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleDelete = async (id: string) => {
    if (!confirm('Удалить это упражнение?')) {
      return;
    }

    setDeletingId(id);
    setError('');
    try {
      await exerciseService.delete(id);
      await loadExercises();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка удаления упражнения');
    } finally {
      setDeletingId(null);
    }
  };

  const filtered = exercises.filter((exercise) => {
    const q = search.trim().toLowerCase();
    if (!q) return true;
    return (
      exercise.name.toLowerCase().includes(q) ||
      (exercise.muscle_group?.name || '').toLowerCase().includes(q) ||
      (exercise.level?.name || '').toLowerCase().includes(q) ||
      exercise.description.toLowerCase().includes(q)
    );
  });

  return (
    <div className="products-list-container">
      <div className="products-header">
        <h1>Управление упражнениями</h1>
        <button onClick={onAddExercise} className="add-button">
          + Добавить упражнение
        </button>
      </div>

      <div className="filters">
        <input
          type="text"
          placeholder="Поиск по названию, группе мышц, уровню..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="search-input"
        />
      </div>

      {error && <div className="error-message">{error}</div>}

      {loading ? (
        <div className="loading">Загрузка...</div>
      ) : filtered.length === 0 ? (
        <div className="empty-state">Упражнения не найдены</div>
      ) : (
        <div className="products-grid">
          {filtered.map((exercise) => (
            <div key={exercise.id} className="product-card">
              {exercise.video_urls?.length ? (
                <div className="exercise-videos">
                  {exercise.video_urls.map((url) => (
                    <video
                      key={url}
                      className="exercise-video"
                      src={url}
                      controls
                      preload="metadata"
                    />
                  ))}
                </div>
              ) : (
                <div className="exercise-video-placeholder">Нет видео</div>
              )}
              <div className="product-info">
                <div className="product-header">
                  <h3>{exercise.name}</h3>
                  <div className="product-actions">
                    <button
                      onClick={() => setEditingExercise(exercise)}
                      className="edit-button"
                      title="Редактировать упражнение"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(exercise.id)}
                      disabled={deletingId === exercise.id}
                      className="delete-button"
                      title="Удалить упражнение"
                    >
                      {deletingId === exercise.id ? '...' : '×'}
                    </button>
                  </div>
                </div>
                {exercise.description && <p>{exercise.description}</p>}
                <div className="product-meta">
                  {exercise.muscle_group?.name && (
                    <span className="category">Группа: {exercise.muscle_group.name}</span>
                  )}
                  {exercise.level?.name && (
                    <span className="stock in-stock">Уровень: {exercise.level.name}</span>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {editingExercise && (
        <ExerciseForm
          exercise={editingExercise}
          onSuccess={() => {
            setEditingExercise(null);
            loadExercises();
          }}
          onCancel={() => setEditingExercise(null)}
        />
      )}
    </div>
  );
};
