export interface User {
  id: number;
  email: string;
  display_name?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}
