import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { List, Card, Tag, Button, Space, Typography, Spin, Empty, Modal, message as antdMessage, Tooltip } from 'antd';
import { PlusOutlined, FileTextOutlined, ClockCircleOutlined, DeleteOutlined, CheckCircleOutlined, CalendarOutlined, FieldTimeOutlined } from '@ant-design/icons';

import { fetchTasks, deleteTask } from '@/api/tasks';
import type { Task } from '@/types';
import { formatDate, formatDateTime, formatRelativeTime, isOverdue } from '@/utils/format';

const { Title, Text } = Typography;

const Tasks: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: tasks, isLoading, isError } = useQuery({
    queryKey: ['tasks'],
    queryFn: fetchTasks,
  });

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

  const handleDeleteTask = (taskId: number, taskTitle: string) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除任务「${taskTitle}」吗？此操作不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => deleteMutation.mutate(taskId),
    });
  };

  const getStatusTag = (status: string) => {
    const colors: Record<string, string> = {
      todo: 'default',
      in_progress: 'processing',
      done: 'success',
      cancelled: 'error',
    };
    const labels: Record<string, string> = {
      todo: '待办',
      in_progress: '进行中',
      done: '已完成',
      cancelled: '已取消',
    };
    return <Tag color={colors[status]}>{labels[status] || status}</Tag>;
  };

  const getPriorityTag = (priority: string) => {
    const colors: Record<string, string> = {
      low: 'blue',
      medium: 'orange',
      high: 'red',
    };
    const labels: Record<string, string> = {
      low: '低',
      medium: '中',
      high: '高',
    };
    return <Tag color={colors[priority]}>{labels[priority] || priority}优先级</Tag>;
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
            onClick={() => { /* TODO: Implement simple create */ }}
          >
            新建任务
          </Button>
        </Space>
      </div>

      {isError ? (
        <Card>
          <Text type="danger">加载任务失败，请稍后重试</Text>
        </Card>
      ) : tasks && tasks.length > 0 ? (
        <List
          grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 3 }}
          dataSource={tasks}
          renderItem={(task: Task) => (
            <List.Item>
              <Card
                hoverable
                title={task.title}
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
    </div>
  );
};

export default Tasks;
