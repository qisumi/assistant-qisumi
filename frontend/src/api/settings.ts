import apiClient from './client';

// Thinking类型枚举
export enum ThinkingType {
  Disabled = 'disabled',
  Enabled = 'enabled',
  Auto = 'auto',
}

// Reasoning强度枚举
export enum ReasoningEffort {
  Low = 'low',
  Medium = 'medium',
  High = 'high',
  Minimal = 'minimal',
}

export interface LLMSettings {
  base_url: string;
  api_key?: string; // 仅在更新时发送，获取时可能为空或脱敏
  model: string;
  thinking_type?: ThinkingType;
  reasoning_effort?: ReasoningEffort;
}

export async function fetchLLMSettings(): Promise<LLMSettings> {
  const { data } = await apiClient.get<LLMSettings>('/settings/llm');
  return data;
}

export async function updateLLMSettings(settings: LLMSettings): Promise<void> {
  await apiClient.post('/settings/llm', settings);
}

// Thinking类型中文映射
export const ThinkingTypeLabels: Record<ThinkingType, string> = {
  [ThinkingType.Disabled]: '不启用',
  [ThinkingType.Enabled]: '启用',
  [ThinkingType.Auto]: '自动',
};

// Reasoning强度中文映射
export const ReasoningEffortLabels: Record<ReasoningEffort, string> = {
  [ReasoningEffort.Low]: '低',
  [ReasoningEffort.Medium]: '中',
  [ReasoningEffort.High]: '高',
  [ReasoningEffort.Minimal]: '不思考',
};
