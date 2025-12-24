import { Tag } from 'antd';
import { TASK_STATUS_LABELS, TASK_STATUS_COLORS, PRIORITY_LABELS, PRIORITY_COLORS } from '@/constants';

/**
 * 获取任务状态标签组件
 */
export const getStatusTag = (status: string) => {
  return (
    <Tag color={TASK_STATUS_COLORS[status] || 'default'}>
      {TASK_STATUS_LABELS[status] || status}
    </Tag>
  );
};

/**
 * 获取优先级标签组件
 */
export const getPriorityTag = (priority: string) => {
  return (
    <Tag color={PRIORITY_COLORS[priority] || 'default'}>
      {PRIORITY_LABELS[priority] || priority}优先级
    </Tag>
  );
};
