export type UserRole = 'admin' | 'user';

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

export interface NamedCatalogItem {
  id: string;
  name: string;
  created_at?: string;
  updated_at?: string;
}

export interface Exercise {
  id: string;
  name: string;
  description: string;
  muscle_group_id?: string | null;
  level_id?: string | null;
  muscle_group?: NamedCatalogItem | null;
  level?: NamedCatalogItem | null;
  video_urls: string[];
  created_at: string;
  updated_at: string;
}

export interface ExerciseInput {
  name: string;
  description?: string;
  muscle_group_id?: string | null;
  level_id?: string | null;
  video_urls?: string[];
}

export interface CreateUserInput {
  username: string;
  email?: string;
  password: string;
  role?: UserRole;
}

export interface UpdateUserInput {
  username?: string;
  email?: string;
  password?: string;
  role?: UserRole;
  is_active?: boolean;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ApiResponse<T> {
  data?: T;
  message?: string;
  error?: string;
}

export interface VideoUploadResult {
  key: string;
  video_url: string;
}
