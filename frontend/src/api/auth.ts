import client from './client';
import { AuthResponse } from '../types/user';

export const login = async (email: string, password: string): Promise<AuthResponse> => {
  const { data } = await client.post<AuthResponse>('/auth/login', { email, password });
  return data;
};

export const register = async (email: string, password: string): Promise<AuthResponse> => {
  const { data } = await client.post<AuthResponse>('/auth/register', { email, password });
  return data;
};
