import { apiService } from './api';
import { CreateUserInput, UpdateUserInput, User } from '../types';

class UserService {
  async getAll(): Promise<User[]> {
    const response = await apiService.getUsers();
    if (response.error) {
      throw new Error(response.error);
    }
    return response.data || [];
  }

  async create(user: CreateUserInput): Promise<User> {
    const response = await apiService.createUser(user);
    if (response.error || !response.data) {
      throw new Error(response.error || 'Failed to create user');
    }
    return response.data;
  }

  async update(id: string, user: UpdateUserInput): Promise<User> {
    const response = await apiService.updateUser(id, user);
    if (response.error || !response.data) {
      throw new Error(response.error || 'Failed to update user');
    }
    return response.data;
  }
}

export const userService = new UserService();
