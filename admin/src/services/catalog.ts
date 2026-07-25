import { apiService } from './api';
import { NamedCatalogItem } from '../types';

class CatalogService {
  async getMuscleGroups(): Promise<NamedCatalogItem[]> {
    const response = await apiService.getMuscleGroups();
    if (response.error) {
      throw new Error(response.error);
    }
    return response.data || [];
  }

  async createMuscleGroup(name: string): Promise<NamedCatalogItem> {
    const response = await apiService.createMuscleGroup(name);
    if (response.error || !response.data) {
      throw new Error(response.error || 'Failed to create muscle group');
    }
    return response.data;
  }

  async getLevels(): Promise<NamedCatalogItem[]> {
    const response = await apiService.getLevels();
    if (response.error) {
      throw new Error(response.error);
    }
    return response.data || [];
  }

  async createLevel(name: string): Promise<NamedCatalogItem> {
    const response = await apiService.createLevel(name);
    if (response.error || !response.data) {
      throw new Error(response.error || 'Failed to create level');
    }
    return response.data;
  }
}

export const catalogService = new CatalogService();
