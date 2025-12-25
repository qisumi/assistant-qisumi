import React from 'react';
import { Typography, Space, Tag, Input, Button, App } from 'antd';
import { CheckOutlined, CloseOutlined } from '@ant-design/icons';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { StatusBadge, PriorityBadge } from '@/components/ui';
import type { Task } from '@/types';
import { updateTask } from '@/api/tasks';

const { Title } = Typography;

interface TaskHeaderProps {
  task: Task;
  onUpdate?: () => void;
}

export const TaskHeader: React.FC<TaskHeaderProps> = ({ task, onUpdate }) => {
  const [editing, setEditing] = React.useState(false);
  const [titleValue, setTitleValue] = React.useState(task.title);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const updateMutation = useMutation({
    mutationFn: (fields: { title: string }) => updateTask(task.id, fields),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', String(task.id)] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      message.success('标题已更新');
      onUpdate?.();
    },
    onError: (err: any) => {
      console.error(err);
      message.error('更新失败，请稍后重试');
    },
  });

  const handleUpdate = async () => {
    if (titleValue.trim() === '') {
      message.error('标题不能为空');
      return;
    }
    if (titleValue !== task.title) {
      await updateMutation.mutateAsync({ title: titleValue });
    }
    setEditing(false);
  };

  const handleCancel = () => {
    setTitleValue(task.title);
    setEditing(false);
  };

  return (
    <div style={{ marginBottom: 16 }}>
      {editing ? (
        <Space.Compact style={{ width: '100%' }}>
          <Input
            value={titleValue}
            onChange={(e) => setTitleValue(e.target.value)}
            onPressEnter={handleUpdate}
            autoFocus
            size="large"
          />
          <Button type="primary" icon={<CheckOutlined />} onClick={handleUpdate} size="large" />
          <Button icon={<CloseOutlined />} onClick={handleCancel} size="large" />
        </Space.Compact>
      ) : (
        <div
          style={{
            cursor: 'pointer',
            padding: '8px',
            borderRadius: 8,
            transition: 'background-color 0.2s',
          }}
          onClick={() => setEditing(true)}
          onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = '#f5f5f5')}
          onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
        >
          <Title
            level={3}
            style={{
              margin: 0,
              fontSize: 'clamp(18px, 4vw, 24px)',
              wordBreak: 'break-word',
              lineHeight: 1.3,
            }}
            ellipsis={{ tooltip: task.title }}
          >
            {task.title}
          </Title>
        </div>
      )}
      <Space wrap style={{ marginTop: 12 }}>
        <StatusBadge status={task.status} />
        <PriorityBadge priority={task.priority} />
        {task.isFocusToday && (
          <Tag
            color="gold"
            style={{
              borderRadius: '4px',
              fontWeight: 500,
              fontSize: '12px',
              padding: '2px 8px',
            }}
          >
            今日聚焦
          </Tag>
        )}
      </Space>
    </div>
  );
};
