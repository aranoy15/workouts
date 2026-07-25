import { apiService } from './api';
import { User } from '../types';

export interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
}

class AuthService {
  private getToken(): string | null {
    return localStorage.getItem('authToken');
  }

  private getUser(): User | null {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    try {
      return JSON.parse(userStr) as User;
    } catch {
      return null;
    }
  }

  getAuthState(): AuthState {
    const token = this.getToken();
    const user = this.getUser();
    return {
      token,
      user,
      isAuthenticated: !!token && !!user,
    };
  }

  async login(identifier: string, password: string): Promise<AuthState> {
    const response = await apiService.login(identifier, password);
    if (response.data?.token && response.data.user) {
      const { token, user } = response.data;
      if (user.role !== 'admin') {
        throw new Error('Доступ только для администраторов');
      }
      localStorage.setItem('authToken', token);
      localStorage.setItem('user', JSON.stringify(user));
      return { token, user, isAuthenticated: true };
    }
    throw new Error(response.error || 'Login failed');
  }

  logout(): void {
    localStorage.removeItem('authToken');
    localStorage.removeItem('user');
  }

  isAdmin(): boolean {
    return this.getUser()?.role === 'admin';
  }
}

export const authService = new AuthService();
