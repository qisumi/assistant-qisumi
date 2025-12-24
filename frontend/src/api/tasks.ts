import apiClient from './client';
import type { Task, TaskDetailResponse, TaskStep } from '@/types';

export const fetchTasks = async (): Promise<Task[]> => {
  const { data } = await apiClient.get<{ tasks: Task[] }>('/tasks');
  return data.tasks;
};

export const fetchTaskDetail = async (taskId: string | number): Promise<TaskDetailResponse> => {
  const { data } = await apiClient.get(`/tasks/${taskId}`);
  return data;
};

export const createTaskFromText = async (rawText: string): Promise<TaskDetailResponse> => {
  const { data } = await apiClient.post('/tasks/from-text', { raw_text: rawText });
  // 清理可能的循环引用
  return JSON.parse(JSON.stringify(data));
};

export interface CreateTaskRequest {
  title: string;
  description?: string;
  priority?: 'low' | 'medium' | 'high';
  isFocusToday?: boolean;
  dueAt?: string | null;
}

export interface UpdateTaskFields {
  title?: string;
  description?: string;
  status?: string;
  priority?: string;
  isFocusToday?: boolean;
  dueAt?: string | null;
  completedAt?: string | null;
}

export const createTask = async (taskData: CreateTaskRequest): Promise<Task> => {
  const { data } = await apiClient.post<{ task: Task }>('/tasks', taskData);
  return data.task;
};

export const deleteTask = async (taskId: string | number): Promise<void> => {
  await apiClient.delete(`/tasks/${taskId}`);
};

export const fetchCompletedTasks = async (): Promise<Task[]> => {
  const { data } = await apiClient.get<{ tasks: Task[] }>('/tasks/completed');
  return data.tasks;
};

export const updateTask = async (taskId: string | number, fields: UpdateTaskFields): Promise<void> => {
  await apiClient.patch(`/tasks/${taskId}`, fields);
};

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

export interface CreateStepRequest {
  title: string;
  detail?: string;
  orderIndex: number;
  status?: string;
  estimateMinutes?: number;
}

export const updateTaskStep = async (
  taskId: string | number,
  stepId: string | number,
  fields: UpdateStepFields
): Promise<void> => {
  await apiClient.patch(`/tasks/${taskId}/steps/${stepId}`, fields);
};

export const addTaskStep = async (
  taskId: string | number,
  stepData: CreateStepRequest
): Promise<TaskStep> => {
  const { data } = await apiClient.post<{ step: TaskStep }>(`/tasks/${taskId}/steps`, stepData);
  return data.step;
};

export const deleteTaskStep = async (
  taskId: string | number,
  stepId: string | number
): Promise<void> => {
  await apiClient.delete(`/tasks/${taskId}/steps/${stepId}`);
};
