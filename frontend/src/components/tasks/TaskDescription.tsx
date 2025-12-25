import React from 'react';
import { Typography, Input, Button, Space, App } from 'antd';
import { CheckOutlined, CloseOutlined } from '@ant-design/icons';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Markdown } from '@/components/ui';
import type { Task } from '@/types';
import { updateTask } from '@/api/tasks';

const { Text } = Typography;

interface TaskDescriptionProps {
  task: Task;
  onUpdate?: () => void;
}

export const TaskDescription: React.FC<TaskDescriptionProps> = ({ task, onUpdate }) => {
  const [editing, setEditing] = React.useState(false);
  const [descriptionValue, setDescriptionValue] = React.useState(task.description);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const updateMutation = useMutation({
    mutationFn: (fields: { description: string }) => updateTask(task.id, fields),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', String(task.id)] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      message.success('描述已更新');
      onUpdate?.();
    },
    onError: (err: any) => {
      console.error(err);
      message.error('更新失败，请稍后重试');
    },
  });

  const handleUpdate = async () => {
    if (descriptionValue !== task.description) {
      await updateMutation.mutateAsync({ description: descriptionValue });
    }
    setEditing(false);
  };

  const handleCancel = () => {
    setDescriptionValue(task.description);
    setEditing(false);
  };

  return (
    <div style={{ marginBottom: 24 }}>
      {editing ? (
        <div>
          <Input.TextArea
            value={descriptionValue}
            onChange={(e) => setDescriptionValue(e.target.value)}
            rows={4}
            autoFocus
          />
          <Space style={{ marginTop: 8 }}>
            <Button type="primary" size="small" icon={<CheckOutlined />} onClick={handleUpdate}>
              保存
            </Button>
            <Button size="small" icon={<CloseOutlined />} onClick={handleCancel}>
              取消
            </Button>
          </Space>
        </div>
      ) : (
        <div
          onClick={() => setEditing(true)}
          style={{
            cursor: 'pointer',
            padding: '12px',
            borderRadius: 8,
            minHeight: 60,
            border: '1px solid #e8e8eb',
          }}
        >
          {task.description ? (
            <Markdown content={task.description} />
          ) : (
            <Text type="secondary" italic>
              点击添加描述...
            </Text>
          )}
        </div>
      )}
    </div>
  );
};
