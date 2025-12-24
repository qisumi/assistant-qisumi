import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Button, Space, Typography, Modal, App,
  Select, Form, Input, DatePicker, Switch, Row, Col
} from 'antd';
import {
  PlusOutlined, FileTextOutlined, CheckCircleOutlined
} from '@ant-design/icons';
import { fetchTasks, createTask } from '@/api/tasks';
import type { Task } from '@/types';
import { TaskCard } from '@/components/ui';
import { PageHeader } from '@/components/layout/PageHeader';

const { TextArea } = Input;

type SortOption = 'createdAt' | 'updatedAt' | 'dueAt' | 'focusToday';

const SORT_OPTIONS = [
  { value: 'focusToday' as const, label: '今日重点任务' },
  { value: 'updatedAt' as const, label: '最近更新' },
  { value: 'createdAt' as const, label: '创建时间' },
  { value: 'dueAt' as const, label: '截止时间' },
];

const Tasks: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [sortBy, setSortBy] = useState<SortOption>('updatedAt');
  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false);
  const [createForm] = Form.useForm();
  const { message } = App.useApp();

  const { data: tasks, isLoading, isError } = useQuery({
    queryKey: ['tasks'],
    queryFn: fetchTasks,
  });

  // 排序任务
  const sortedTasks = useMemo(() => {
    if (!tasks) return [];

    const tasksCopy = [...tasks];

    switch (sortBy) {
      case 'focusToday':
        return tasksCopy.sort((a, b) => {
          if (a.isFocusToday && !b.isFocusToday) return -1;
          if (!a.isFocusToday && b.isFocusToday) return 1;
          return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
        });
      case 'createdAt':
        return tasksCopy.sort((a, b) =>
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
      case 'updatedAt':
        return tasksCopy.sort((a, b) =>
          new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
        );
      case 'dueAt':
        return tasksCopy.sort((a, b) => {
          if (!a.dueAt && !b.dueAt) return 0;
          if (!a.dueAt) return 1;
          if (!b.dueAt) return -1;
          return new Date(a.dueAt).getTime() - new Date(b.dueAt).getTime();
        });
      default:
        return tasksCopy;
    }
  }, [tasks, sortBy]);

  const createMutation = useMutation({
    mutationFn: createTask,
    onSuccess: (newTask) => {
      message.success('任务创建成功');
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      setIsCreateModalVisible(false);
      createForm.resetFields();
      navigate(`/tasks/${newTask.id}`);
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '创建失败，请稍后重试');
    },
  });

  const handleCreateTask = async () => {
    try {
      const values = await createForm.validateFields();
      const taskData = {
        title: values.title,
        description: values.description || '',
        priority: values.priority || 'medium',
        isFocusToday: values.isFocusToday || false,
        dueAt: values.dueAt ? values.dueAt.toISOString() : null,
      };
      createMutation.mutate(taskData);
    } catch (error) {
      // 表单验证失败
    }
  };

  const handleCancelCreate = () => {
    setIsCreateModalVisible(false);
    createForm.resetFields();
  };

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', paddingTop: 100 }}>
        {/* Loading handled by PageSpinner component if needed */}
      </div>
    );
  }

  return (
    <div>
      {/* Page Header */}
      <PageHeader
        title="任务列表"
        extra={
          <Space wrap>
            <Button
              icon={<CheckCircleOutlined />}
              onClick={() => navigate('/completed-tasks')}
            >
              已完成任务
            </Button>
            <Button
              icon={<FileTextOutlined />}
              onClick={() => navigate('/create-from-text')}
            >
              从文本创建
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setIsCreateModalVisible(true)}
            >
              新建任务
            </Button>
          </Space>
        }
      />

      {/* Sort selector */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <span style={{ color: '#6b6b76' }}>排序方式：</span>
          <Select
            value={sortBy}
            onChange={setSortBy}
            style={{ width: 180 }}
            options={SORT_OPTIONS}
          />
        </Space>
      </div>

      {/* Task list */}
      {isError ? (
        <div style={{
          padding: 48,
          textAlign: 'center',
          background: '#ffffff',
          borderRadius: '12px'
        }}>
          <Typography.Text type="danger">加载任务失败，请稍后重试</Typography.Text>
        </div>
      ) : sortedTasks && sortedTasks.length > 0 ? (
        <Row gutter={[16, 16]}>
          {sortedTasks.map((task: Task, index: number) => (
            <Col key={task.id} xs={24} sm={12} md={12} lg={8} xl={8} xxl={6}>
              <TaskCard
                task={task}
                onClick={() => navigate(`/tasks/${task.id}`)}
                index={index}
              />
            </Col>
          ))}
        </Row>
      ) : (
        <div style={{
          padding: 64,
          textAlign: 'center',
          background: '#ffffff',
          borderRadius: '12px'
        }}>
          <Typography.Text type="secondary" style={{ fontSize: '16px' }}>
            暂无任务
          </Typography.Text>
          <div style={{ marginTop: 24 }}>
            <Button type="primary" onClick={() => setIsCreateModalVisible(true)}>
              创建第一个任务
            </Button>
          </div>
        </div>
      )}

      {/* Create task modal */}
      <Modal
        title="新建任务"
        open={isCreateModalVisible}
        onOk={handleCreateTask}
        onCancel={handleCancelCreate}
        okText="创建"
        cancelText="取消"
        confirmLoading={createMutation.isPending}
        width={600}
      >
        <Form
          form={createForm}
          layout="vertical"
          style={{ marginTop: 24 }}
        >
          <Form.Item
            name="title"
            label="任务标题"
            rules={[{ required: true, message: '请输入任务标题' }]}
          >
            <Input placeholder="例如：完成项目文档" />
          </Form.Item>

          <Form.Item
            name="description"
            label="任务描述"
          >
            <TextArea
              rows={4}
              placeholder="描述任务的详细信息（可选）"
            />
          </Form.Item>

          <Form.Item
            name="priority"
            label="优先级"
            initialValue="medium"
          >
            <Select>
              <Select.Option value="low">低</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="high">高</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="dueAt"
            label="截止时间"
          >
            <DatePicker
              showTime
              style={{ width: '100%' }}
              placeholder="选择截止时间（可选）"
              format="YYYY-MM-DD HH:mm"
            />
          </Form.Item>

          <Form.Item
            name="isFocusToday"
            label="今日重点"
            valuePropName="checked"
            initialValue={false}
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Tasks;
