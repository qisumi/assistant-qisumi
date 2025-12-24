import React from 'react';
import { Empty, Button } from 'antd';
import { PlusOutlined, MessageOutlined, FileTextOutlined } from '@ant-design/icons';

interface EmptyStateProps {
  title?: string;
  description?: string;
  actionText?: string;
  onAction?: () => void;
  illustration?: React.ReactNode;
}

/**
 * Generic empty state component
 */
export const EmptyState: React.FC<EmptyStateProps> = ({
  title = '暂无数据',
  description = '还没有任何内容',
  actionText,
  onAction,
  illustration
}) => {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '300px',
        padding: 48
      }}
    >
      {illustration || <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
      <h3 style={{ marginTop: 16, marginBottom: 8, color: '#18181b' }}>{title}</h3>
      <p style={{ color: '#6b6b76', marginBottom: 24 }}>{description}</p>
      {actionText && onAction && (
        <Button type="primary" icon={<PlusOutlined />} onClick={onAction}>
          {actionText}
        </Button>
      )}
    </div>
  );
};

/**
 * Empty state for tasks list
 */
export const NoTasks: React.FC<{ onCreateTask: () => void }> = ({ onCreateTask }) => {
  return (
    <EmptyState
      title="暂无任务"
      description="创建你的第一个任务开始管理你的工作"
      actionText="创建任务"
      onAction={onCreateTask}
      illustration={<FileTextOutlined style={{ fontSize: 64, color: '#d3d3d7' }} />}
    />
  );
};

/**
 * Empty state for chat messages
 */
export const NoMessages: React.FC<{ onSendMessage: () => void }> = ({ onSendMessage }) => {
  return (
    <EmptyState
      title="开始对话"
      description="向助手提问，开始智能任务管理"
      actionText="发送消息"
      onAction={onSendMessage}
      illustration={<MessageOutlined style={{ fontSize: 64, color: '#d3d3d7' }} />}
    />
  );
};
