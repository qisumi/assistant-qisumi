import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Card,
  Tag,
  Button,
  Space,
  Typography,
  Spin,
  Divider,
  List,
  Checkbox,
  Breadcrumb,
  Modal,
  message as antdMessage,
  Tooltip,
  Descriptions
} from 'antd';
import { ArrowLeftOutlined, ClockCircleOutlined, CalendarOutlined, DeleteOutlined, FieldTimeOutlined, CheckCircleOutlined } from '@ant-design/icons';

import { fetchTaskDetail, deleteTask } from '@/api/tasks';
import { fetchSessionMessages, sendSessionMessage } from '@/api/sessions';
import { ChatWindow } from '@/components/chat/ChatWindow';
import type { TaskStep } from '@/types';
import { formatDate, formatDateTime, formatRelativeTime, formatTimeRange, isOverdue } from '@/utils/format';

const { Title, Text, Paragraph } = Typography;

const TaskDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['taskDetail', id],
    queryFn: () => fetchTaskDetail(id!),
    enabled: !!id,
  });

  const sessionId = data?.session.id;

  const { data: messagesData } = useQuery({
    queryKey: ['sessionMessages', sessionId],
    queryFn: () => fetchSessionMessages(sessionId!),
    enabled: !!sessionId,
  });

  const sendMutation = useMutation({
    mutationFn: (content: string) => sendSessionMessage(sessionId!, content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
      if (sessionId) {
        queryClient.invalidateQueries({ queryKey: ['sessionMessages', sessionId] });
      }
    },
    onError: (err: any) => {
      console.error(err);
      antdMessage.error('发送失败，请稍后重试');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteTask,
    onSuccess: () => {
      antdMessage.success('任务已删除');
      navigate('/tasks');
    },
    onError: () => {
      antdMessage.error('删除失败，请稍后重试');
    },
  });

  const handleDeleteTask = () => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除任务「${task.title}」吗？此操作不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => deleteMutation.mutate(task.id),
    });
  };

  if (isLoading) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <Text type="danger">加载任务详情失败</Text>
          <Button onClick={() => navigate('/tasks')} style={{ marginLeft: 16 }}>返回列表</Button>
        </Card>
      </div>
    );
  }

  const { task } = data;
  const messages = messagesData?.messages ?? data.messages ?? [];

  const statusLabels: Record<string, string> = {
    todo: '待办',
    in_progress: '进行中',
    done: '已完成',
    cancelled: '已取消',
  };

  const priorityLabels: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
  };

  const stepStatusLabels: Record<string, string> = {
    locked: '已锁定',
    todo: '待办',
    in_progress: '进行中',
    done: '已完成',
    blocked: '受阻',
  };

  const getStatusTag = (status: string) => {
    const colors: Record<string, string> = {
      todo: 'default',
      in_progress: 'processing',
      done: 'success',
      cancelled: 'error',
    };
    return <Tag color={colors[status]}>{statusLabels[status] || status}</Tag>;
  };

  const getStepStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      locked: 'default',
      todo: 'blue',
      in_progress: 'orange',
      done: 'green',
      blocked: 'red',
    };
    return colors[status] || 'default';
  };

  return (
    <div style={{ padding: '12px 24px' }}>
      <Breadcrumb
        style={{ marginBottom: 16 }}
        items={[
          { title: <a onClick={() => navigate('/tasks')}>任务列表</a> },
          { title: task.title },
        ]}
      />

      <div style={{ display: 'flex', gap: 24, alignItems: 'flex-start' }}>
        {/* 左侧：任务信息与步骤 */}
        <div style={{ flex: 1, minWidth: 0 }}>
          <Card>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div>
                <Title level={3} style={{ margin: 0 }}>{task.title}</Title>
                <Space style={{ marginTop: 8 }}>
                  {getStatusTag(task.status)}
                  <Tag color={task.priority === 'high' ? 'red' : task.priority === 'medium' ? 'orange' : 'blue'}>
                    {(priorityLabels[task.priority] || task.priority)}优先级
                  </Tag>
                </Space>
              </div>
              <Space>
                <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tasks')}>返回</Button>
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  onClick={handleDeleteTask}
                >
                  删除任务
                </Button>
              </Space>
            </div>

            <Divider />

            <div style={{ marginBottom: 24 }}>
              <Title level={5}>任务描述</Title>
              <Paragraph type="secondary">
                {task.description || '暂无描述'}
              </Paragraph>
              
              {/* 时间信息 */}
              <Descriptions size="small" column={2} style={{ marginTop: 16 }}>
                <Descriptions.Item label={<Space size={4}><FieldTimeOutlined />创建时间</Space>}>
                  <Tooltip title={formatDateTime(task.createdAt)}>
                    {formatRelativeTime(task.createdAt)}
                  </Tooltip>
                </Descriptions.Item>
                <Descriptions.Item label={<Space size={4}><FieldTimeOutlined />更新时间</Space>}>
                  <Tooltip title={formatDateTime(task.updatedAt)}>
                    {formatRelativeTime(task.updatedAt)}
                  </Tooltip>
                </Descriptions.Item>
                {task.dueAt && (
                  <Descriptions.Item label={<Space size={4}><CalendarOutlined />截止日期</Space>}>
                    <Tooltip title={formatDateTime(task.dueAt)}>
                      <Text style={{ color: isOverdue(task.dueAt) ? '#ff4d4f' : undefined }}>
                        {formatDate(task.dueAt)}
                      </Text>
                    </Tooltip>
                  </Descriptions.Item>
                )}
                {task.completedAt && (
                  <Descriptions.Item label={<Space size={4}><CheckCircleOutlined />完成时间</Space>}>
                    <Tooltip title={formatDateTime(task.completedAt)}>
                      {formatRelativeTime(task.completedAt)}
                    </Tooltip>
                  </Descriptions.Item>
                )}
              </Descriptions>
            </div>

            <div>
              <Title level={5}>执行步骤</Title>
              <List
                dataSource={task.steps || []}
                renderItem={(step: TaskStep) => (
                  <List.Item>
                    <div style={{ display: 'flex', alignItems: 'flex-start', width: '100%' }}>
                      <Checkbox checked={step.status === 'done'} disabled style={{ marginTop: 4 }} />
                      <div style={{ marginLeft: 12, flex: 1 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                          <Text strong delete={step.status === 'done'}>{step.title}</Text>
                          <Tag color={getStepStatusColor(step.status)}>
                            {stepStatusLabels[step.status] || step.status}
                          </Tag>
                        </div>
                        {step.detail && (
                          <Paragraph type="secondary" style={{ fontSize: 12, margin: '4px 0' }}>
                            {step.detail}
                          </Paragraph>
                        )}
                        {/* 步骤时间信息 */}
                        <Space size={12} style={{ marginTop: 4 }}>
                          {step.estimateMinutes && (
                            <Space size={4}>
                              <ClockCircleOutlined style={{ fontSize: 12, color: '#8c8c8c' }} />
                              <Text type="secondary" style={{ fontSize: 12 }}>
                                预计 {step.estimateMinutes} 分钟
                              </Text>
                            </Space>
                          )}
                          {step.plannedStart && step.plannedEnd && (
                            <Tooltip title={formatTimeRange(step.plannedStart, step.plannedEnd)}>
                              <Text type="secondary" style={{ fontSize: 12 }}>
                                计划: {formatDate(step.plannedStart)} - {formatDate(step.plannedEnd)}
                              </Text>
                            </Tooltip>
                          )}
                          {step.completedAt && (
                            <Tooltip title={`完成于 ${formatDateTime(step.completedAt)}`}>
                              <Space size={4}>
                                <CheckCircleOutlined style={{ fontSize: 12, color: '#52c41a' }} />
                                <Text type="secondary" style={{ fontSize: 12 }}>
                                  {formatRelativeTime(step.completedAt)}
                                </Text>
                              </Space>
                            </Tooltip>
                          )}
                        </Space>
                      </div>
                    </div>
                  </List.Item>
                )}
              />
            </div>
          </Card>
        </div>

        {/* 右侧：聊天窗口 */}
        <div style={{ width: 400, flexShrink: 0 }}>
          <Card title="任务助手" styles={{ body: { padding: 0 } }}>
            <ChatWindow
              messages={messages}
              onSend={(content) => sendMutation.mutateAsync(content)}
              sending={sendMutation.isPending}
              height={600}
            />
          </Card>
        </div>
      </div>
    </div>
  );
};

export default TaskDetail;
