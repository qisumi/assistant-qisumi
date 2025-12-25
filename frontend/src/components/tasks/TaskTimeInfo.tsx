import React from 'react';
import { Typography, DatePicker, Space, Button, App, Descriptions } from 'antd';
import {
  FieldTimeOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  CloseOutlined,
} from '@ant-design/icons';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import dayjs from 'dayjs';
import type { Task } from '@/types';
import { updateTask } from '@/api/tasks';
import { formatDate, formatDateTime, formatRelativeTime, isOverdue } from '@/utils/format';

const { Text } = Typography;

interface TaskTimeInfoProps {
  task: Task;
  onUpdate?: () => void;
}

export const TaskTimeInfo: React.FC<TaskTimeInfoProps> = ({ task, onUpdate }) => {
  const [editingDueAt, setEditingDueAt] = React.useState(false);
  const [dueAtValue, setDueAtValue] = React.useState<dayjs.Dayjs | null>(null);
  const [dueAtPickerOpen, setDueAtPickerOpen] = React.useState(false);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const updateMutation = useMutation({
    mutationFn: (fields: { dueAt: string | null }) => updateTask(task.id, fields),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', String(task.id)] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      message.success('截止时间已更新');
      onUpdate?.();
    },
    onError: (err: any) => {
      console.error(err);
      message.error('更新失败，请稍后重试');
    },
  });

  const handleUpdateDueAt = async () => {
    try {
      await updateMutation.mutateAsync({
        dueAt: dueAtValue ? dueAtValue.toISOString() : null,
      });
      setEditingDueAt(false);
      setDueAtPickerOpen(false);
    } catch (error) {
      // Error handled in mutation
    }
  };

  const handleDueDateChange = (date: dayjs.Dayjs | null) => {
    setDueAtValue(date);
    setDueAtPickerOpen(false);
  };

  const handleCancelEditDueAt = () => {
    setDueAtValue(null);
    setEditingDueAt(false);
    setDueAtPickerOpen(false);
  };

  const handleStartEditDueAt = () => {
    setDueAtValue(task.dueAt ? dayjs(task.dueAt) : null);
    setEditingDueAt(true);
    setDueAtPickerOpen(true);
  };

  return (
    <Descriptions size="small" column={{ xs: 1, sm: 2 }} style={{ marginBottom: 24 }}>
      <Descriptions.Item label={<Space size={4}>{<FieldTimeOutlined />}创建时间</Space>}>
        <Text title={formatDateTime(task.createdAt)}>{formatRelativeTime(task.createdAt)}</Text>
      </Descriptions.Item>
      <Descriptions.Item label={<Space size={4}>{<FieldTimeOutlined />}更新时间</Space>}>
        <Text title={formatDateTime(task.updatedAt)}>{formatRelativeTime(task.updatedAt)}</Text>
      </Descriptions.Item>
      <Descriptions.Item label={<Space size={4}>{<CalendarOutlined />}截止日期</Space>}>
        {editingDueAt ? (
          <Space.Compact style={{ display: 'flex', alignItems: 'center' }}>
            <DatePicker
              showTime
              format="YYYY-MM-DD HH:mm"
              value={dueAtValue}
              onChange={handleDueDateChange}
              onOpenChange={setDueAtPickerOpen}
              open={dueAtPickerOpen}
              autoFocus
              size="small"
              style={{ width: 200 }}
            />
            <Button type="primary" size="small" icon={<CheckOutlined />} onClick={handleUpdateDueAt} />
            <Button size="small" icon={<CloseOutlined />} onClick={handleCancelEditDueAt} />
          </Space.Compact>
        ) : (
          <div
            onClick={handleStartEditDueAt}
            style={{
              cursor: 'pointer',
              display: 'inline-block',
            }}
          >
            {task.dueAt ? (
              <Text strong style={{ color: isOverdue(task.dueAt) ? '#ff4d4f' : '#1890ff' }}>
                {formatDate(task.dueAt)}
              </Text>
            ) : (
              <Text type="secondary" italic>
                未设置
              </Text>
            )}
          </div>
        )}
      </Descriptions.Item>
      {task.completedAt && (
        <Descriptions.Item label={<Space size={4}>{<CheckCircleOutlined />}完成时间</Space>}>
          <Text title={formatDateTime(task.completedAt)}>{formatRelativeTime(task.completedAt)}</Text>
        </Descriptions.Item>
      )}
    </Descriptions>
  );
};
