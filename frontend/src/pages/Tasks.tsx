import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  List, Card, Tag, Button, Space, Typography, Spin, Empty, Modal, 
  message as antdMessage, Tooltip, Select, Form, Input, DatePicker, Switch 
} from 'antd';
import { 
  PlusOutlined, FileTextOutlined, DeleteOutlined, CheckCircleOutlined, 
  CalendarOutlined, FieldTimeOutlined, StarFilled 
} from '@ant-design/icons';
import { fetchTasks, deleteTask, createTask } from '@/api/tasks';
import type { Task } from '@/types';
import { formatDate, formatDateTime, formatRelativeTime, isOverdue } from '@/utils/format';
import { getStatusTag, getPriorityTag } from '@/utils/tags';
import { confirmDelete } from '@/utils/dialog';

const { Title, Text } = Typography;
const { TextArea } = Input;

type SortOption = 'createdAt' | 'updatedAt' | 'dueAt' | 'focusToday';

const SORT_OPTIONS = [
  { value: 'focusToday' as const, label: '今日重点任务' },
  { value: 'updatedAt' as const, label: '最近更新' },
  { value: 'createdAt' as const, label: '创建时间' },
  { value: 'dueAt' as const, label: '预期完成时间' },
];

const Tasks: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [sortBy, setSortBy] = useState<SortOption>('updatedAt');
  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false);
  const [createForm] = Form.useForm();

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
        // 今日重点任务排在前面，然后按更新时间倒序
        return tasksCopy.sort((a, b) => {
          if (a.isFocusToday && !b.isFocusToday) return -1;
          if (!a.isFocusToday && b.isFocusToday) return 1;
          return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
        });
      case 'createdAt':
        // 创建时间倒序（最新的在前）
        return tasksCopy.sort((a, b) => 
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
      case 'updatedAt':
        // 更新时间倒序（最近更新的在前）
        return tasksCopy.sort((a, b) => 
          new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
        );
      case 'dueAt':
        // 截止时间升序（最早截止的在前），没有截止时间的排在后面
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

  const deleteMutation = useMutation({
    mutationFn: deleteTask,
    onSuccess: () => {
      antdMessage.success('任务已删除');
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
    onError: () => {
      antdMessage.error('删除失败，请稍后重试');
    },
  });

  const createMutation = useMutation({
    mutationFn: createTask,
    onSuccess: (newTask) => {
      antdMessage.success('任务创建成功');
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      setIsCreateModalVisible(false);
      createForm.resetFields();
      // 导航到新任务详情页
      navigate(`/tasks/${newTask.id}`);
    },
    onError: (error: any) => {
      antdMessage.error(error?.response?.data?.error || '创建失败，请稍后重试');
    },
  });

  const handleDeleteTask = (taskId: number, taskTitle: string) => {
    confirmDelete('任务', taskTitle, () => deleteMutation.mutate(taskId));
  };

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
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>任务列表</Title>
        <Space>
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
      </div>

      {/* 排序选择器 */}
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Text>排序方式：</Text>
          <Select value={sortBy} onChange={setSortBy} style={{ width: 180 }} options={SORT_OPTIONS} />
        </Space>
      </div>

      {isError ? (
        <Card>
          <Text type="danger">加载任务失败，请稍后重试</Text>
        </Card>
      ) : sortedTasks && sortedTasks.length > 0 ? (
        <List
          grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 3 }}
          dataSource={sortedTasks}
          renderItem={(task: Task) => (
            <List.Item>
              <Card
                hoverable
                title={
                  <Space>
                    {task.isFocusToday && <StarFilled style={{ color: '#faad14' }} />}
                    <span>{task.title}</span>
                  </Space>
                }
                extra={
                  <Space>
                    {getStatusTag(task.status)}
                    <Button
                      type="text"
                      danger
                      icon={<DeleteOutlined />}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteTask(task.id, task.title);
                      }}
                    />
                  </Space>
                }
                onClick={() => navigate(`/tasks/${task.id}`)}
                style={{
                  borderColor: task.isFocusToday ? '#faad14' : undefined,
                  borderWidth: task.isFocusToday ? 2 : undefined,
                }}
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Text type="secondary" ellipsis={{ tooltip: task.description }}>
                    {task.description || '无描述'}
                  </Text>
                  
                  {/* 时间信息 */}
                  <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                    <Space size={8} wrap>
                      {task.createdAt && (
                        <Tooltip title={`创建于 ${formatDateTime(task.createdAt)}`}>
                          <Space size={4}>
                            <FieldTimeOutlined style={{ fontSize: 12 }} />
                            <Text type="secondary">创建: {formatRelativeTime(task.createdAt)}</Text>
                          </Space>
                        </Tooltip>
                      )}
                      {task.updatedAt && task.updatedAt !== task.createdAt && (
                        <Tooltip title={`更新于 ${formatDateTime(task.updatedAt)}`}>
                          <Text type="secondary">更新: {formatRelativeTime(task.updatedAt)}</Text>
                        </Tooltip>
                      )}
                      {task.dueAt && (
                        <Tooltip title={`截止于 ${formatDateTime(task.dueAt)}`}>
                          <Space size={4}>
                            <CalendarOutlined style={{ fontSize: 12, color: isOverdue(task.dueAt) ? '#ff4d4f' : '#8c8c8c' }} />
                            <Text type="secondary" style={{ color: isOverdue(task.dueAt) ? '#ff4d4f' : undefined }}>
                              截止: {formatDate(task.dueAt)}
                            </Text>
                          </Space>
                        </Tooltip>
                      )}
                    </Space>
                  </div>

                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 8 }}>
                    <Space>
                      {getPriorityTag(task.priority)}
                      {task.isFocusToday && (
                        <Tag icon={<StarFilled />} color="gold">
                          今日重点
                        </Tag>
                      )}
                    </Space>
                  </div>
                </Space>
              </Card>
            </List.Item>
          )}
        />
      ) : (
        <Empty description="暂无任务" style={{ marginTop: 64 }}>
          <Button type="primary" onClick={() => navigate('/create-from-text')}>
            立即创建一个
          </Button>
        </Empty>
      )}

      {/* 创建任务对话框 */}
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
