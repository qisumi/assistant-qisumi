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

// 更新任务字段
export interface UpdateTaskFields {
  title?: string;
  description?: string;
  status?: string;
  priority?: string;
  isFocusToday?: boolean;
  dueAt?: string | null;
  completedAt?: string | null;
}

export async function updateTask(taskId: string | number, fields: UpdateTaskFields): Promise<void> {
  await apiClient.patch(`/tasks/${taskId}`, fields);
}

// 更新步骤字段
export interface UpdateStepFields {
  title?: string;
  detail?: string;
  status?: string;
  blockingReason?: string;
  estimateMinutes?: number;
  orderIndex?: number;
  plannedStart?: string | null;
  plannedEnd?: string | null;
  completedAt?: string | null;
}

export async function updateTaskStep(
  taskId: string | number,
  stepId: string | number,
  fields: UpdateStepFields
): Promise<void> {
  await apiClient.patch(`/tasks/${taskId}/steps/${stepId}`, fields);
}
