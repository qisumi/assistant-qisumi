import apiClient from './client';
import type { Session, SessionMessagesResponse, SendMessageResponse } from '@/types';

export const fetchSessionMessages = async (
  sessionId: number | string
): Promise<SessionMessagesResponse> => {
  const { data } = await apiClient.get(`/sessions/${sessionId}/messages`);
  return data;
};

export const sendSessionMessage = async (
  sessionId: number | string,
  content: string
): Promise<SendMessageResponse> => {
  const { data } = await apiClient.post(`/sessions/${sessionId}/messages`, { content });
  return data;
};

export const clearSessionMessages = async (
  sessionId: number | string
): Promise<{ success: boolean }> => {
  const { data } = await apiClient.delete(`/sessions/${sessionId}/messages`);
  return data;
};

export const getOrCreateGlobalSession = async (): Promise<Session> => {
  const { data } = await apiClient.get<{ session: Session }>('/sessions/global');
  return data.session;
};
