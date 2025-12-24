// 任务状态标签映射
export const TASK_STATUS_LABELS: Record<string, string> = {
  todo: '待办',
  in_progress: '进行中',
  done: '已完成',
  cancelled: '已取消',
};

// 任务状态颜色映射
export const TASK_STATUS_COLORS: Record<string, string> = {
  todo: 'default',
  in_progress: 'processing',
  done: 'success',
  cancelled: 'error',
};

// 优先级标签映射
export const PRIORITY_LABELS: Record<string, string> = {
  low: '低',
  medium: '中',
  high: '高',
};

// 优先级颜色映射
export const PRIORITY_COLORS: Record<string, string> = {
  low: 'blue',
  medium: 'orange',
  high: 'red',
};

// Agent标签映射
export const AGENT_LABELS: Record<string, string> = {
  executor: '小奇（执行）',
  planner: '小奇（规划）',
  summarizer: '小奇（总结）',
  global: '小奇（全局）',
};

// Agent图标颜色映射
export const AGENT_COLORS: Record<string, string> = {
  executor: '#52c41a',
  planner: '#722ed1',
  summarizer: '#fa8c16',
  global: '#faad14',
};
