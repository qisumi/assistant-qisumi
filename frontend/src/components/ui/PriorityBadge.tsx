import React from 'react';
import { Tag } from 'antd';
import { PRIORITY_LABELS, PRIORITY_COLORS } from '@/constants';

interface PriorityBadgeProps {
  priority: string;
  showLabel?: boolean;
  style?: React.CSSProperties;
  className?: string;
}

/**
 * Unified priority badge component with consistent styling
 */
export const PriorityBadge: React.FC<PriorityBadgeProps> = ({
  priority,
  showLabel = true,
  style,
  className
}) => {
  const label = PRIORITY_LABELS[priority] || priority;
  const color = PRIORITY_COLORS[priority] || 'default';

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
      {showLabel ? `${label}优先级` : label}
    </Tag>
  );
};
