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

// 删除任务
export async function deleteTask(taskId: string | number): Promise<void> {
  await apiClient.delete(`/tasks/${taskId}`);
}

// 获取已完成任务列表
export async function fetchCompletedTasks(): Promise<Task[]> {
  const { data } = await apiClient.get<{ tasks: Task[]; total: number }>('/tasks/completed');
  return data.tasks;
}
