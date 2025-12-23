import apiClient from './client';

export interface LLMSettings {
  base_url: string;
  api_key?: string; // 仅在更新时发送，获取时可能为空或脱敏
  model: string;
}

export async function fetchLLMSettings(): Promise<LLMSettings> {
  const { data } = await apiClient.get<LLMSettings>('/settings/llm');
  return data;
}

export async function updateLLMSettings(settings: LLMSettings): Promise<void> {
  await apiClient.post('/settings/llm', settings);
}
