import { apiService } from './api';
import { Exercise, ExerciseInput, VideoUploadResult } from '../types';

class ExerciseService {
  async getAll(): Promise<Exercise[]> {
    const response = await apiService.getExercises();
    if (response.error) {
      throw new Error(response.error);
    }
    return response.data || [];
  }

  async create(exercise: ExerciseInput): Promise<Exercise> {
    const response = await apiService.createExercise(exercise);
    if (response.error || !response.data) {
      throw new Error(response.error || 'Failed to create exercise');
    }
    return response.data;
  }

  async update(id: string, exercise: ExerciseInput): Promise<Exercise> {
    const response = await apiService.updateExercise(id, exercise);
    if (response.error || !response.data) {
      throw new Error(response.error || 'Failed to update exercise');
    }
    return response.data;
  }

  async delete(id: string): Promise<void> {
    const response = await apiService.deleteExercise(id);
    if (response.error) {
      throw new Error(response.error);
    }
  }

  async uploadVideo(file: File): Promise<VideoUploadResult> {
    return apiService.uploadVideo(file);
  }

  async deleteVideo(payload: { key?: string; video_url?: string }): Promise<void> {
    const response = await apiService.deleteVideo(payload);
    if (response.error) {
      throw new Error(response.error);
    }
  }
}

export const exerciseService = new ExerciseService();
