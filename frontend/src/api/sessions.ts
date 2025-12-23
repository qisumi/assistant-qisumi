import apiClient from './client';
import type {
  Session,
  SessionMessagesResponse,
  SendMessageResponse,
} from '@/types';

// 获取消息列表
export async function fetchSessionMessages(
  sessionId: number | string,
): Promise<SessionMessagesResponse> {
  const { data } = await apiClient.get<SessionMessagesResponse>(`/sessions/${sessionId}/messages`);
  return data;
}

// 发送消息
export async function sendSessionMessage(
  sessionId: number | string,
  content: string,
): Promise<SendMessageResponse> {
  const { data } = await apiClient.post<SendMessageResponse>(
    `/sessions/${sessionId}/messages`,
    { content },
  );
  return data;
}

// 获取或创建当前用户的全局会话
export async function getOrCreateGlobalSession(): Promise<Session> {
  const { data } = await apiClient.get<{ session: Session }>('/sessions/global');
  return data.session;
}
