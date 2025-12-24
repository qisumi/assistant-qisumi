import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Card, Tag, Button, Space, Typography, Spin, Divider, List, Checkbox, 
  Breadcrumb, Modal, message as antdMessage, Tooltip, Descriptions, 
  Form, Input, InputNumber, DatePicker
} from 'antd';
import { 
  ArrowLeftOutlined, CalendarOutlined, DeleteOutlined, FieldTimeOutlined, 
  CheckCircleOutlined, EditOutlined, PlusOutlined, CloseOutlined, CheckOutlined 
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { fetchTaskDetail, deleteTask, addTaskStep, deleteTaskStep, updateTask } from '@/api/tasks';
import { fetchSessionMessages, sendSessionMessage } from '@/api/sessions';
import { ChatWindow } from '@/components/chat/ChatWindow';
import { TaskEditForm } from '@/components/tasks/TaskEditForm';
import { StepEditableField } from '@/components/tasks/StepEditableField';
import type { TaskStep } from '@/types';
import { formatDate, formatDateTime, formatRelativeTime, formatTimeRange, isOverdue } from '@/utils/format';
import { getStatusTag, getPriorityTag } from '@/utils/tags';
import { confirmDelete } from '@/utils/dialog';

const { Title, Text, Paragraph } = Typography;

const TaskDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [editingTask, setEditingTask] = useState(false);
  const [editingTitle, setEditingTitle] = useState(false);
  const [editingDescription, setEditingDescription] = useState(false);
  const [editingDueAt, setEditingDueAt] = useState(false);
  const [isAddStepModalVisible, setIsAddStepModalVisible] = useState(false);
  const [addStepForm] = Form.useForm();
  const [titleValue, setTitleValue] = useState('');
  const [descriptionValue, setDescriptionValue] = useState('');

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

  const updateTaskMutation = useMutation({
    mutationFn: (fields: any) => updateTask(task.id, fields),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
    onError: (err: any) => {
      console.error(err);
      antdMessage.error('更新失败，请稍后重试');
    },
  });

  const addStepMutation = useMutation({
    mutationFn: ({ taskId, stepData }: { taskId: string | number; stepData: any }) =>
      addTaskStep(taskId, stepData),
    onSuccess: () => {
      antdMessage.success('步骤添加成功');
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
      setIsAddStepModalVisible(false);
      addStepForm.resetFields();
    },
    onError: (error: any) => {
      antdMessage.error(error?.response?.data?.error || '添加失败，请稍后重试');
    },
  });

  const deleteStepMutation = useMutation({
    mutationFn: ({ taskId, stepId }: { taskId: string | number; stepId: string | number }) =>
      deleteTaskStep(taskId, stepId),
    onSuccess: () => {
      antdMessage.success('步骤已删除');
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
    },
    onError: (error: any) => {
      antdMessage.error(error?.response?.data?.error || '删除失败，请稍后重试');
    },
  });

  const handleDeleteTask = () => {
    confirmDelete('任务', task.title, () => deleteMutation.mutate(task.id));
  };

  const handleAddStep = async () => {
    try {
      const values = await addStepForm.validateFields();
      const stepData = {
        title: values.title,
        detail: values.detail || '',
        orderIndex: (task.steps?.length || 0),
        status: 'todo',
        estimateMinutes: values.estimateMinutes || null,
      };
      addStepMutation.mutate({ taskId: id!, stepData });
    } catch (error) {
      // 表单验证失败
    }
  };

  const handleDeleteStep = (stepId: number, stepTitle: string) => {
    confirmDelete('步骤', stepTitle, () => deleteStepMutation.mutate({ taskId: id!, stepId }));
  };

  const handleUpdateDueAt = async (date: dayjs.Dayjs | null) => {
    try {
      await updateTaskMutation.mutateAsync({ 
        dueAt: date ? date.toISOString() : null 
      });
      antdMessage.success('截止时间已更新');
      setEditingDueAt(false);
    } catch (error) {
      // 错误已在mutation中处理
    }
  };

  const handleUpdateTitle = async () => {
    if (titleValue.trim() === '') {
      antdMessage.error('标题不能为空');
      return;
    }
    if (titleValue !== task.title) {
      try {
        await updateTaskMutation.mutateAsync({ title: titleValue });
        antdMessage.success('标题已更新');
      } catch (error) {
        // 错误已在mutation中处理
      }
    }
    setEditingTitle(false);
  };

  const handleUpdateDescription = async () => {
    if (descriptionValue !== task.description) {
      try {
        await updateTaskMutation.mutateAsync({ description: descriptionValue });
        antdMessage.success('描述已更新');
      } catch (error) {
        // 错误已在mutation中处理
      }
    }
    setEditingDescription(false);
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
            {editingTask ? (
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
                  <Title level={4} style={{ margin: 0 }}>编辑任务</Title>
                  <Button onClick={() => setEditingTask(false)}>取消</Button>
                </div>
                <TaskEditForm
                  task={task}
                  onCancel={() => setEditingTask(false)}
                  onSuccess={() => setEditingTask(false)}
                />
              </div>
            ) : (
              <>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div style={{ flex: 1 }}>
                    {/* 任务标题 - 可点击编辑 */}
                    {editingTitle ? (
                      <Space.Compact style={{ width: '100%', maxWidth: 600 }}>
                        <Input
                          value={titleValue}
                          onChange={(e) => setTitleValue(e.target.value)}
                          onPressEnter={handleUpdateTitle}
                          autoFocus
                          size="large"
                        />
                        <Button
                          type="primary"
                          icon={<CheckOutlined />}
                          onClick={handleUpdateTitle}
                          size="large"
                        />
                        <Button
                          icon={<CloseOutlined />}
                          onClick={() => {
                            setTitleValue(task.title);
                            setEditingTitle(false);
                          }}
                          size="large"
                        />
                      </Space.Compact>
                    ) : (
                      <Title 
                        level={3} 
                        style={{ 
                          margin: 0, 
                          cursor: 'pointer',
                          padding: '4px 8px',
                          borderRadius: 4,
                          display: 'inline-block'
                        }}
                        onClick={() => {
                          setTitleValue(task.title);
                          setEditingTitle(true);
                        }}
                      >
                        {task.title}
                      </Title>
                    )}
                    <Space style={{ marginTop: 8 }}>
                      {getStatusTag(task.status)}
                      {getPriorityTag(task.priority)}
                      {task.isFocusToday && <Tag color="gold">今日聚焦</Tag>}
                    </Space>
                  </div>
                  <Space>
                    <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tasks')}>返回</Button>
                    <Button icon={<EditOutlined />} onClick={() => setEditingTask(true)}>编辑</Button>
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
                  {/* 任务描述 - 可点击编辑 */}
                  {editingDescription ? (
                    <div>
                      <Input.TextArea
                        value={descriptionValue}
                        onChange={(e) => setDescriptionValue(e.target.value)}
                        rows={4}
                        autoFocus
                      />
                      <Space style={{ marginTop: 8 }}>
                        <Button
                          type="primary"
                          size="small"
                          icon={<CheckOutlined />}
                          onClick={handleUpdateDescription}
                        >
                          保存
                        </Button>
                        <Button
                          size="small"
                          icon={<CloseOutlined />}
                          onClick={() => {
                            setDescriptionValue(task.description);
                            setEditingDescription(false);
                          }}
                        >
                          取消
                        </Button>
                      </Space>
                    </div>
                  ) : (
                    <div 
                      onClick={() => {
                        setDescriptionValue(task.description);
                        setEditingDescription(true);
                      }} 
                      style={{ 
                        cursor: 'pointer',
                        padding: '8px',
                        borderRadius: 4,
                        minHeight: 60
                      }}
                    >
                      {task.description ? (
                        <Paragraph type="secondary" style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                          {task.description}
                        </Paragraph>
                      ) : (
                        <Text type="secondary" italic>点击添加描述...</Text>
                      )}
                    </div>
                  )}
              
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
                <Descriptions.Item label={<Space size={4}><CalendarOutlined />截止日期</Space>}>
                  {editingDueAt ? (
                    <Space.Compact>
                      <DatePicker
                        showTime
                        format="YYYY-MM-DD HH:mm"
                        defaultValue={task.dueAt ? dayjs(task.dueAt) : null}
                        onChange={(date) => handleUpdateDueAt(date)}
                        onBlur={() => setEditingDueAt(false)}
                        autoFocus
                        open
                        style={{ width: 200 }}
                      />
                      <Button
                        size="small"
                        icon={<CloseOutlined />}
                        onClick={() => setEditingDueAt(false)}
                      />
                    </Space.Compact>
                  ) : (
                    <div
                      onClick={() => setEditingDueAt(true)}
                      style={{
                        cursor: 'pointer',
                        display: 'inline-block'
                      }}
                    >
                      {task.dueAt ? (
                        <Tooltip title={formatDateTime(task.dueAt)}>
                          <Text strong style={{ color: isOverdue(task.dueAt) ? '#ff4d4f' : '#1890ff' }}>
                            {formatDate(task.dueAt)}
                          </Text>
                        </Tooltip>
                      ) : (
                        <Text type="warning" italic>未设置</Text>
                      )}
                    </div>
                  )}
                </Descriptions.Item>
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
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                    <Title level={5} style={{ margin: 0 }}>执行步骤</Title>
                    <Button
                      type="dashed"
                      icon={<PlusOutlined />}
                      onClick={() => setIsAddStepModalVisible(true)}
                      size="small"
                    >
                      添加步骤
                    </Button>
                  </div>
                  <List
                    dataSource={task.steps || []}
                    renderItem={(step: TaskStep) => (
                      <List.Item
                        actions={[
                          <Button
                            key="delete"
                            type="text"
                            danger
                            size="small"
                            icon={<DeleteOutlined />}
                            onClick={() => handleDeleteStep(step.id, step.title)}
                          />
                        ]}
                      >
                        <div style={{ display: 'flex', alignItems: 'flex-start', width: '100%' }}>
                          <Checkbox checked={step.status === 'done'} disabled style={{ marginTop: 4 }} />
                          <div style={{ marginLeft: 12, flex: 1 }}>
                            <StepEditableField step={step} taskId={task.id} />
                            {/* 步骤时间信息 - 始终显示 */}
                            <Space size={12} style={{ marginTop: 8 }}>
                              {step.plannedStart && step.plannedEnd && (
                                <Tooltip title={formatTimeRange(step.plannedStart, step.plannedEnd)}>
                                  <Space size={4}>
                                    <CalendarOutlined style={{ fontSize: 12, color: '#8c8c8c' }} />
                                    <Text type="secondary" style={{ fontSize: 12 }}>
                                      {formatDate(step.plannedStart)} - {formatDate(step.plannedEnd)}
                                    </Text>
                                  </Space>
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
              </>
            )}
          </Card>
        </div>

        {/* 右侧：聊天窗口 */}
        <div style={{ width: 400, flexShrink: 0 }}>
          <Card title="小奇（任务）" styles={{ body: { padding: 0 } }}>
            <ChatWindow
              messages={messages}
              onSend={(content) => sendMutation.mutateAsync(content)}
              sending={sendMutation.isPending}
              height={600}
            />
          </Card>
        </div>
      </div>

      {/* 添加步骤对话框 */}
      <Modal
        title="添加步骤"
        open={isAddStepModalVisible}
        onOk={handleAddStep}
        onCancel={() => {
          setIsAddStepModalVisible(false);
          addStepForm.resetFields();
        }}
        okText="添加"
        cancelText="取消"
        confirmLoading={addStepMutation.isPending}
      >
        <Form
          form={addStepForm}
          layout="vertical"
          style={{ marginTop: 24 }}
        >
          <Form.Item
            name="title"
            label="步骤标题"
            rules={[{ required: true, message: '请输入步骤标题' }]}
          >
            <Input placeholder="例如：准备项目材料" />
          </Form.Item>

          <Form.Item
            name="detail"
            label="步骤详情"
          >
            <Input.TextArea
              rows={3}
              placeholder="描述这个步骤的详细内容（可选）"
            />
          </Form.Item>

          <Form.Item
            name="estimateMinutes"
            label="预计耗时（分钟）"
          >
            <InputNumber
              min={1}
              style={{ width: '100%' }}
              placeholder="例如：30"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default TaskDetail;
