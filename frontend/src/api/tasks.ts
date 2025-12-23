import apiClient from './client';
import type { Task, TaskDetailResponse } from '@/types';

export async function fetchTasks(): Promise<Task[]> {
  const { data } = await apiClient.get<{ tasks: Task[]; total: number }>('/tasks');
  return data.tasks;
}

export async function fetchTaskDetail(taskId: string | number): Promise<TaskDetailResponse> {
  const { data } = await apiClient.get<TaskDetailResponse>(`/tasks/${taskId}`);
  return data;
}

// 从文本创建任务
export async function createTaskFromText(rawText: string): Promise<TaskDetailResponse> {
  const { data } = await apiClient.post<TaskDetailResponse>('/tasks/from-text', {
    raw_text: rawText,
  });
  return data;
}
