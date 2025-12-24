import React from 'react';
import { Card, Tag, Space, Typography } from 'antd';
import { StarFilled, ClockCircleOutlined } from '@ant-design/icons';
import { motion } from 'framer-motion';
import { StatusBadge, PriorityBadge } from './';
import type { Task } from '@/types';
import dayjs from 'dayjs';
import { hoverCard } from '@/animations/variants';

const { Text, Paragraph } = Typography;

interface TaskCardProps {
  task: Task;
  onClick?: () => void;
  style?: React.CSSProperties;
  className?: string;
  variant?: 'default' | 'compact' | 'focus';
  index?: number;
}

/**
 * Reusable task card component with hover effects and animations
 */
export const TaskCard: React.FC<TaskCardProps> = ({
  task,
  onClick,
  style,
  className,
  variant = 'default',
  index = 0
}) => {
  const isOverdue = task.dueAt && dayjs(task.dueAt).isBefore(dayjs()) && task.status !== 'done';

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        delay: index * 0.05,
        type: 'spring',
        stiffness: 300,
        damping: 24
      }}
      {...hoverCard}
    >
      <Card
        hoverable
        onClick={onClick}
        className={className}
        style={{
          borderRadius: '12px',
          cursor: 'pointer',
          border: variant === 'focus' || task.isFocusToday
            ? '2px solid #faad14'
            : '1px solid #e8e8eb',
          position: 'relative',
          ...style
        }}
        styles={{ body: { padding: variant === 'compact' ? '12px 16px' : '16px' } }}
      >
        {/* Focus indicator */}
        {(variant === 'focus' || task.isFocusToday) && (
          <div style={{
            position: 'absolute',
            top: '8px',
            right: '8px',
            color: '#faad14',
          }}>
            <StarFilled style={{ fontSize: '14px' }} />
          </div>
        )}

        {/* Title */}
        <div style={{ marginBottom: 8 }}>
          <Text
            strong
            style={{
              fontSize: '15px',
              color: '#18181b',
              display: 'block'
            }}
            ellipsis={{ tooltip: task.title }}
          >
            {task.title}
          </Text>
        </div>

        {/* Description (only for non-compact variant) */}
        {variant !== 'compact' && task.description && (
          <Paragraph
            ellipsis={{ rows: 2, tooltip: task.description }}
            style={{
              fontSize: '13px',
              color: '#6b6b76',
              marginBottom: 12
            }}
          >
            {task.description}
          </Paragraph>
        )}

        {/* Meta information */}
        <Space wrap size={[8, 8]}>
          <StatusBadge status={task.status} />
          <PriorityBadge priority={task.priority} />

          {/* Due date */}
          <Tag
            icon={<ClockCircleOutlined />}
            style={{
              borderRadius: '4px',
              fontSize: '12px',
              padding: '2px 8px',
              color: isOverdue ? '#eb5757' : '#6b6b76'
            }}
          >
            {task.dueAt ? dayjs(task.dueAt).format('MM-DD') : '未设置'}
          </Tag>

          {/* Step count */}
          {task.steps && task.steps.length > 0 && (
            <Tag style={{
              borderRadius: '4px',
              fontSize: '12px',
              padding: '2px 8px',
              color: '#6b6b76'
            }}>
              {task.steps.filter(s => s.status === 'done').length}/{task.steps.length} 步骤
            </Tag>
          )}
        </Space>
      </Card>
    </motion.div>
  );
};
