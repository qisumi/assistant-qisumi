import React, { useState } from 'react';
import { Form, Input, Select, DatePicker, Switch, Button, Space, message as antdMessage } from 'antd';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import dayjs from 'dayjs';
import type { Task } from '@/types';
import { updateTask, type UpdateTaskFields } from '@/api/tasks';

const { TextArea } = Input;
const { Option } = Select;

interface TaskEditFormProps {
  task: Task;
  onCancel?: () => void;
  onSuccess?: () => void;
}

export const TaskEditForm: React.FC<TaskEditFormProps> = ({ task, onCancel, onSuccess }) => {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();
  const [loading, setLoading] = useState(false);

  const updateMutation = useMutation({
    mutationFn: (fields: UpdateTaskFields) => updateTask(task.id, fields),
    onSuccess: () => {
      antdMessage.success('任务更新成功');
      queryClient.invalidateQueries({ queryKey: ['taskDetail', String(task.id)] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      onSuccess?.();
    },
    onError: (err: any) => {
      console.error(err);
      antdMessage.error('更新失败，请稍后重试');
    },
  });

  const handleSubmit = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      
      const fields: UpdateTaskFields = {};
      
      if (values.title !== task.title) {
        fields.title = values.title;
      }
      if (values.description !== task.description) {
        fields.description = values.description;
      }
      if (values.status !== task.status) {
        fields.status = values.status;
      }
      if (values.priority !== task.priority) {
        fields.priority = values.priority;
      }
      if (values.isFocusToday !== task.isFocusToday) {
        fields.isFocusToday = values.isFocusToday;
      }
      
      // 处理日期
      if (values.dueAt) {
        fields.dueAt = values.dueAt.toISOString();
      } else if (task.dueAt) {
        fields.dueAt = null; // 清除日期
      }

      // 只有在有更改时才发送请求
      if (Object.keys(fields).length > 0) {
        await updateMutation.mutateAsync(fields);
      } else {
        antdMessage.info('没有任何更改');
        onSuccess?.();
      }
    } catch (err) {
      console.error('表单验证失败:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Form
      form={form}
      layout="vertical"
      initialValues={{
        title: task.title,
        description: task.description,
        status: task.status,
        priority: task.priority,
        isFocusToday: task.isFocusToday ?? false,
        dueAt: task.dueAt ? dayjs(task.dueAt) : null,
      }}
    >
      <Form.Item
        label="任务标题"
        name="title"
        rules={[{ required: true, message: '请输入任务标题' }]}
      >
        <Input placeholder="输入任务标题" />
      </Form.Item>

      <Form.Item label="任务描述" name="description">
        <TextArea rows={4} placeholder="输入任务描述" />
      </Form.Item>

      <Form.Item label="状态" name="status" rules={[{ required: true }]}>
        <Select>
          <Option value="todo">待办</Option>
          <Option value="in_progress">进行中</Option>
          <Option value="done">已完成</Option>
          <Option value="cancelled">已取消</Option>
        </Select>
      </Form.Item>

      <Form.Item label="优先级" name="priority" rules={[{ required: true }]}>
        <Select>
          <Option value="low">低</Option>
          <Option value="medium">中</Option>
          <Option value="high">高</Option>
        </Select>
      </Form.Item>

      <Form.Item label="截止时间" name="dueAt">
        <DatePicker showTime style={{ width: '100%' }} format="YYYY-MM-DD HH:mm" />
      </Form.Item>

      <Form.Item label="今日聚焦" name="isFocusToday" valuePropName="checked">
        <Switch />
      </Form.Item>

      <Form.Item>
        <Space>
          <Button type="primary" onClick={handleSubmit} loading={loading}>
            保存
          </Button>
          {onCancel && (
            <Button onClick={onCancel}>取消</Button>
          )}
        </Space>
      </Form.Item>
    </Form>
  );
};
