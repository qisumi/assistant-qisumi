import React from 'react';
import { Tag } from 'antd';
import { TASK_STATUS_LABELS, TASK_STATUS_COLORS } from '@/constants';

interface StatusBadgeProps {
  status: string;
  style?: React.CSSProperties;
  className?: string;
}

/**
 * Unified status badge component with consistent styling
 */
export const StatusBadge: React.FC<StatusBadgeProps> = ({ status, style, className }) => {
  const label = TASK_STATUS_LABELS[status] || status;
  const color = TASK_STATUS_COLORS[status] || 'default';

  return (
    <Tag
      color={color}
      style={{
        borderRadius: '4px',
        fontWeight: 500,
        fontSize: '12px',
        padding: '2px 8px',
        ...style
      }}
      className={className}
    >
      {label}
    </Tag>
  );
};
