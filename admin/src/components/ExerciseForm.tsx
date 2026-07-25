import { useState, FormEvent, useEffect } from 'react';
import { Exercise, ExerciseInput, NamedCatalogItem } from '../types';
import { exerciseService } from '../services/exercises';
import { catalogService } from '../services/catalog';
import './ExerciseForm.css';

interface ExerciseFormProps {
  exercise?: Exercise | null;
  onSuccess: () => void;
  onCancel: () => void;
}

export const ExerciseForm = ({ exercise, onSuccess, onCancel }: ExerciseFormProps) => {
  const [formData, setFormData] = useState<ExerciseInput>({
    name: '',
    description: '',
    muscle_group_id: '',
    level_id: '',
    video_urls: [],
  });
  const [muscleGroups, setMuscleGroups] = useState<NamedCatalogItem[]>([]);
  const [levels, setLevels] = useState<NamedCatalogItem[]>([]);
  const [urlDraft, setUrlDraft] = useState('');
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingCatalogs, setLoadingCatalogs] = useState(true);
  const isEdit = !!exercise;
  const videoURLs = formData.video_urls || [];

  const loadCatalogs = async () => {
    setLoadingCatalogs(true);
    try {
      const [groups, levelItems] = await Promise.all([
        catalogService.getMuscleGroups(),
        catalogService.getLevels(),
      ]);
      setMuscleGroups(groups);
      setLevels(levelItems);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки справочников');
    } finally {
      setLoadingCatalogs(false);
    }
  };

  useEffect(() => {
    void loadCatalogs();
  }, []);

  useEffect(() => {
    if (exercise) {
      setFormData({
        name: exercise.name || '',
        description: exercise.description || '',
        muscle_group_id: exercise.muscle_group_id || exercise.muscle_group?.id || '',
        level_id: exercise.level_id || exercise.level?.id || '',
        video_urls: [...(exercise.video_urls || [])],
      });
    } else {
      setFormData({
        name: '',
        description: '',
        muscle_group_id: '',
        level_id: '',
        video_urls: [],
      });
    }
    setUrlDraft('');
    setError('');
  }, [exercise]);

  const handleChange = (field: 'name' | 'description', value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const setVideoURLs = (urls: string[]) => {
    setFormData((prev) => ({ ...prev, video_urls: urls }));
  };

  const addVideoURL = (url: string) => {
    const trimmed = url.trim();
    if (!trimmed) return;
    if (videoURLs.includes(trimmed)) {
      setError('Это видео уже добавлено');
      return;
    }
    setError('');
    setVideoURLs([...videoURLs, trimmed]);
  };

  const removeVideoURL = (url: string) => {
    setVideoURLs(videoURLs.filter((item) => item !== url));
  };

  const handleUpload = async (file: File | null) => {
    if (!file) return;
    setUploading(true);
    setError('');
    try {
      const uploaded = await exerciseService.uploadVideo(file);
      addVideoURL(uploaded.video_url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки видео');
    } finally {
      setUploading(false);
    }
  };

  const createMuscleGroup = async () => {
    const name = window.prompt('Название группы мышц');
    if (!name?.trim()) return;
    try {
      const created = await catalogService.createMuscleGroup(name.trim());
      setMuscleGroups((prev) => [...prev, created].sort((a, b) => a.name.localeCompare(b.name)));
      setFormData((prev) => ({ ...prev, muscle_group_id: created.id }));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка создания группы мышц');
    }
  };

  const createLevel = async () => {
    const name = window.prompt('Название уровня');
    if (!name?.trim()) return;
    try {
      const created = await catalogService.createLevel(name.trim());
      setLevels((prev) => [...prev, created].sort((a, b) => a.name.localeCompare(b.name)));
      setFormData((prev) => ({ ...prev, level_id: created.id }));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка создания уровня');
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    const previousURLs = exercise?.video_urls || [];
    const nextURLs = formData.video_urls || [];
    const removedURLs = previousURLs.filter((url) => !nextURLs.includes(url));

    try {
      const payload: ExerciseInput = {
        name: formData.name,
        description: formData.description,
        muscle_group_id: formData.muscle_group_id || null,
        level_id: formData.level_id || null,
        video_urls: nextURLs,
      };

      if (isEdit && exercise) {
        await exerciseService.update(exercise.id, payload);
        await Promise.all(
          removedURLs.map(async (url) => {
            try {
              await exerciseService.deleteVideo({ video_url: url });
            } catch {
              // ignore cleanup errors
            }
          })
        );
      } else {
        await exerciseService.create(payload);
      }
      onSuccess();
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : isEdit
            ? 'Ошибка обновления упражнения'
            : 'Ошибка создания упражнения'
      );
    } finally {
      setLoading(false);
    }
  };

  const formCard = (
    <div className="product-form-card">
      <h2>{isEdit ? 'Редактировать упражнение' : 'Добавить упражнение'}</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="name">Название *</label>
          <input
            id="name"
            type="text"
            value={formData.name}
            onChange={(e) => handleChange('name', e.target.value)}
            required
            disabled={loading}
          />
        </div>

        <div className="form-group">
          <label htmlFor="description">Описание</label>
          <textarea
            id="description"
            value={formData.description}
            onChange={(e) => handleChange('description', e.target.value)}
            rows={4}
            disabled={loading}
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="muscle_group_id">Группа мышц</label>
            <div className="catalog-row">
              <select
                id="muscle_group_id"
                value={formData.muscle_group_id || ''}
                onChange={(e) => setFormData((prev) => ({ ...prev, muscle_group_id: e.target.value }))}
                disabled={loading || loadingCatalogs}
              >
                <option value="">Не выбрано</option>
                {muscleGroups.map((group) => (
                  <option key={group.id} value={group.id}>
                    {group.name}
                  </option>
                ))}
              </select>
              <button type="button" className="add-url-button" onClick={createMuscleGroup} disabled={loading}>
                +
              </button>
            </div>
          </div>
          <div className="form-group">
            <label htmlFor="level_id">Уровень</label>
            <div className="catalog-row">
              <select
                id="level_id"
                value={formData.level_id || ''}
                onChange={(e) => setFormData((prev) => ({ ...prev, level_id: e.target.value }))}
                disabled={loading || loadingCatalogs}
              >
                <option value="">Не выбрано</option>
                {levels.map((level) => (
                  <option key={level.id} value={level.id}>
                    {level.name}
                  </option>
                ))}
              </select>
              <button type="button" className="add-url-button" onClick={createLevel} disabled={loading}>
                +
              </button>
            </div>
          </div>
        </div>

        <div className="form-group">
          <label>Видео</label>
          <div className="video-list">
            {videoURLs.length === 0 ? (
              <div className="video-list-empty">Видео пока нет</div>
            ) : (
              videoURLs.map((url) => (
                <div key={url} className="video-list-item">
                  <video src={url} controls preload="metadata" />
                  <div className="video-list-item-meta">
                    <a href={url} target="_blank" rel="noreferrer">
                      {url}
                    </a>
                    <button
                      type="button"
                      className="remove-video-button"
                      disabled={loading || uploading}
                      onClick={() => removeVideoURL(url)}
                    >
                      Удалить
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="video">Загрузить видео (до 100 МБ)</label>
          <input
            id="video"
            type="file"
            accept="video/*"
            disabled={loading || uploading}
            onChange={(e) => {
              const file = e.target.files?.[0] || null;
              e.target.value = '';
              void handleUpload(file);
            }}
          />
          {uploading && <div className="upload-status">Загрузка...</div>}
        </div>

        <div className="form-group">
          <label htmlFor="video_url">Или добавить URL</label>
          <div className="url-add-row">
            <input
              id="video_url"
              type="url"
              value={urlDraft}
              onChange={(e) => setUrlDraft(e.target.value)}
              disabled={loading || uploading}
              placeholder="https://..."
            />
            <button
              type="button"
              className="add-url-button"
              disabled={loading || uploading || !urlDraft.trim()}
              onClick={() => {
                addVideoURL(urlDraft);
                setUrlDraft('');
              }}
            >
              Добавить
            </button>
          </div>
        </div>

        {error && <div className="error-message">{error}</div>}

        <div className="form-actions">
          <button type="button" onClick={onCancel} className="cancel-button" disabled={loading || uploading}>
            Отмена
          </button>
          <button type="submit" className="submit-button" disabled={loading || uploading}>
            {loading ? 'Сохранение...' : isEdit ? 'Сохранить' : 'Создать'}
          </button>
        </div>
      </form>
    </div>
  );

  if (isEdit) {
    return (
      <div className="exercise-form-overlay" onClick={onCancel}>
        <div className="exercise-form-modal" onClick={(e) => e.stopPropagation()}>
          {formCard}
        </div>
      </div>
    );
  }

  return <div className="product-form-container">{formCard}</div>;
};
