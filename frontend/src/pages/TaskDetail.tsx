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
  message as antdMessage 
} from 'antd';
import { ArrowLeftOutlined, ClockCircleOutlined, CalendarOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';

import { fetchTaskDetail } from '@/api/tasks';
import { sendSessionMessage } from '@/api/sessions';
import { ChatWindow } from '@/components/chat/ChatWindow';
import type { TaskStep } from '@/types';

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

  const sendMutation = useMutation({
    mutationFn: (content: string) => sendSessionMessage(data!.session.id, content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
    },
    onError: (err: any) => {
      console.error(err);
      antdMessage.error('发送失败，请稍后重试');
    },
  });

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

  const { task, session, messages = [] } = data;

  const getStatusTag = (status: string) => {
    const colors: Record<string, string> = {
      todo: 'default',
      in_progress: 'processing',
      done: 'success',
      cancelled: 'error',
    };
    return <Tag color={colors[status]}>{status.toUpperCase()}</Tag>;
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
                    {task.priority.toUpperCase()} 优先级
                  </Tag>
                </Space>
              </div>
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tasks')}>返回</Button>
            </div>

            <Divider />

            <div style={{ marginBottom: 24 }}>
              <Title level={5}>任务描述</Title>
              <Paragraph type="secondary">
                {task.description || '暂无描述'}
              </Paragraph>
              <Space size="large">
                {task.dueAt && (
                  <Space>
                    <CalendarOutlined />
                    <Text>截止日期: {dayjs(task.dueAt).format('YYYY-MM-DD HH:mm')}</Text>
                  </Space>
                )}
              </Space>
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
                          <Tag color={getStepStatusColor(step.status)}>{step.status}</Tag>
                        </div>
                        {step.detail && (
                          <Paragraph type="secondary" style={{ fontSize: 12, margin: '4px 0' }}>
                            {step.detail}
                          </Paragraph>
                        )}
                        {step.estimateMinutes && (
                          <Space size={4}>
                            <ClockCircleOutlined style={{ fontSize: 12, color: '#8c8c8c' }} />
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              预计 {step.estimateMinutes} 分钟
                            </Text>
                          </Space>
                        )}
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
