// src/types/index.ts

// 任务状态
export type TaskStatus = 'todo' | 'in_progress' | 'done' | 'cancelled';

// 步骤状态
export type StepStatus = 'locked' | 'todo' | 'in_progress' | 'done' | 'blocked';

// 优先级
export type TaskPriority = 'low' | 'medium' | 'high';

// 会话类型
export type SessionType = 'task' | 'global';

// 消息角色
export type MessageRole = 'user' | 'assistant' | 'system';

// Agent 名称（后端 messages.agent_name）
export type AgentName =
  | 'executor'
  | 'planner'
  | 'summarizer'
  | 'global'
  | 'system'
  | null;

export interface TaskStep {
  id: number;
  taskId: number;
  orderIndex: number;
  title: string;
  detail: string;
  status: StepStatus;
  blockingReason?: string | null;
  estimateMinutes?: number | null;
  plannedStart?: string | null; // ISO 字符串
  plannedEnd?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Task {
  id: number;
  userId: number;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  dueAt?: string | null;
  createdAt: string;
  updatedAt: string;
  steps?: TaskStep[];
}

// 任务详情 API 返回值
export interface TaskDetailResponse {
  task: Task;
  session: Session;
  messages?: Message[];
}

export interface Session {
  id: number;
  userId: number;
  taskId?: number | null;
  type: SessionType;
  createdAt: string;
}

export interface Message {
  id: number;
  sessionId: number;
  role: MessageRole;
  agentName?: AgentName;
  content: string;
  createdAt: string;
}

// 会话消息接口返回
export interface SessionMessagesResponse {
  messages: Message[];
}

// 发送消息接口返回
export interface SendMessageResponse {
  assistantMessage: string;
  taskPatches?: unknown;
}
