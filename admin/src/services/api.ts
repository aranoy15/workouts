import {
  ApiResponse,
  AuthResponse,
  CreateUserInput,
  Exercise,
  ExerciseInput,
  NamedCatalogItem,
  UpdateUserInput,
  User,
  VideoUploadResult,
} from '../types';

const API_BASE_URL =
  (import.meta.env.VITE_API_URL && import.meta.env.VITE_API_URL.trim()) ||
  'http://localhost:8080/api';

class ApiService {
  private getAuthToken(): string | null {
    return localStorage.getItem('authToken');
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    skipAuth = false
  ): Promise<T> {
    const token = this.getAuthToken();

    if (!skipAuth && !token) {
      throw new Error('Требуется авторизация. Пожалуйста, войдите в систему.');
    }

    const customHeaders = (options.headers as Record<string, string>) || {};
    const headers: Record<string, string> = {
      ...customHeaders,
    };

    if (!(options.body instanceof FormData)) {
      headers['Content-Type'] = headers['Content-Type'] || 'application/json';
    }

    if (!skipAuth && token) {
      headers['X-Auth-Token'] = token;
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      let errorMessage = `Ошибка ${response.status}: ${response.statusText}`;

      try {
        const errorData = await response.json();
        if (errorData.error) {
          errorMessage = errorData.error;
        } else if (errorData.message) {
          errorMessage = errorData.message;
        }
      } catch {
        if (response.status === 401) {
          errorMessage = 'Требуется авторизация. Пожалуйста, войдите в систему.';
        } else if (response.status === 403) {
          errorMessage = 'Недостаточно прав доступа. Требуется роль администратора.';
        }
      }

      throw new Error(errorMessage);
    }

    return response.json();
  }

  async login(identifier: string, password: string) {
    const body = identifier.includes('@')
      ? { email: identifier, password }
      : { username: identifier, password };

    return this.request<ApiResponse<AuthResponse>>(
      '/auth/login',
      {
        method: 'POST',
        body: JSON.stringify(body),
      },
      true
    );
  }

  async getUsers() {
    return this.request<ApiResponse<User[]>>('/users');
  }

  async createUser(user: CreateUserInput) {
    return this.request<ApiResponse<User>>('/users', {
      method: 'POST',
      body: JSON.stringify(user),
    });
  }

  async updateUser(id: string, user: UpdateUserInput) {
    return this.request<ApiResponse<User>>(`/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(user),
    });
  }

  async getExercises() {
    return this.request<ApiResponse<Exercise[]>>('/exercises', {}, true);
  }

  async getMuscleGroups() {
    return this.request<ApiResponse<NamedCatalogItem[]>>('/muscle-groups', {}, true);
  }

  async createMuscleGroup(name: string) {
    return this.request<ApiResponse<NamedCatalogItem>>('/muscle-groups', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  }

  async getLevels() {
    return this.request<ApiResponse<NamedCatalogItem[]>>('/levels', {}, true);
  }

  async createLevel(name: string) {
    return this.request<ApiResponse<NamedCatalogItem>>('/levels', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  }

  async createExercise(exercise: ExerciseInput) {
    return this.request<ApiResponse<Exercise>>('/exercises', {
      method: 'POST',
      body: JSON.stringify(exercise),
    });
  }

  async updateExercise(id: string, exercise: ExerciseInput) {
    return this.request<ApiResponse<Exercise>>(`/exercises/${id}`, {
      method: 'PUT',
      body: JSON.stringify(exercise),
    });
  }

  async deleteExercise(id: string) {
    return this.request<ApiResponse<unknown>>(`/exercises/${id}`, {
      method: 'DELETE',
    });
  }

  async uploadVideo(file: File): Promise<VideoUploadResult> {
    const formData = new FormData();
    formData.append('video', file);

    const response = await this.request<ApiResponse<VideoUploadResult>>('/videos', {
      method: 'POST',
      body: formData,
    });

    if (!response.data?.video_url) {
      throw new Error(response.error || 'Неверный ответ от сервера при загрузке видео');
    }

    return response.data;
  }

  async deleteVideo(payload: { key?: string; video_url?: string }) {
    return this.request<ApiResponse<unknown>>('/videos', {
      method: 'DELETE',
      body: JSON.stringify(payload),
    });
  }
}

export const apiService = new ApiService();
