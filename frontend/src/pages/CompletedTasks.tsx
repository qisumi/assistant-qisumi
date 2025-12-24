import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  List, Card, Button, Space, Typography, Spin, Empty, Modal, 
  message as antdMessage, Tooltip 
} from 'antd';
import { 
  ArrowLeftOutlined, CheckCircleOutlined, DeleteOutlined, 
  CalendarOutlined, FieldTimeOutlined 
} from '@ant-design/icons';
import { fetchCompletedTasks, deleteTask } from '@/api/tasks';
import type { Task } from '@/types';
import { formatDate, formatDateTime, formatRelativeTime } from '@/utils/format';
import { getStatusTag, getPriorityTag } from '@/utils/tags';
import { confirmDelete } from '@/utils/dialog';

const { Title, Text } = Typography;

const CompletedTasks: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: tasks, isLoading, isError } = useQuery({
    queryKey: ['completedTasks'],
    queryFn: fetchCompletedTasks,
  });

  const deleteMutation = useMutation({
    mutationFn: deleteTask,
    onSuccess: () => {
      antdMessage.success('任务已删除');
      queryClient.invalidateQueries({ queryKey: ['completedTasks'] });
    },
    onError: () => {
      antdMessage.error('删除失败，请稍后重试');
    },
  });

  const handleDeleteTask = (taskId: number, taskTitle: string) => {
    confirmDelete('任务', taskTitle, () => deleteMutation.mutate(taskId));
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
        <Space>
          <Button 
            icon={<ArrowLeftOutlined />} 
            onClick={() => navigate('/tasks')}
          >
            返回任务列表
          </Button>
          <Title level={2} style={{ margin: 0 }}>已完成任务</Title>
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
                      {task.completedAt && (
                        <Tooltip title={`完成于 ${formatDateTime(task.completedAt)}`}>
                          <Space size={4}>
                            <CheckCircleOutlined style={{ fontSize: 12, color: '#52c41a' }} />
                            <Text type="secondary">完成: {formatRelativeTime(task.completedAt)}</Text>
                          </Space>
                        </Tooltip>
                      )}
                      {task.dueAt && (
                        <Tooltip title={`截止于 ${formatDateTime(task.dueAt)}`}>
                          <Space size={4}>
                            <CalendarOutlined style={{ fontSize: 12 }} />
                            <Text type="secondary">截止: {formatDate(task.dueAt)}</Text>
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
        <Empty 
          description="暂无已完成任务" 
          style={{ marginTop: 64 }}
        >
          <Button type="primary" onClick={() => navigate('/tasks')}>
            查看待办任务
          </Button>
        </Empty>
      )}
    </div>
  );
};

export default CompletedTasks;
