import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Card, Tag, Button, Space, Typography, Spin, Divider, List, Checkbox,
  Breadcrumb, Modal, App, Tooltip, Descriptions,
  Form, Input, InputNumber, DatePicker, Row, Col, Grid
} from 'antd';
import {
  ArrowLeftOutlined, CalendarOutlined, DeleteOutlined, FieldTimeOutlined,
  CheckCircleOutlined, EditOutlined, PlusOutlined, CloseOutlined, CheckOutlined,
  ClockCircleOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { fetchTaskDetail, deleteTask, addTaskStep, deleteTaskStep, updateTask, updateTaskStep } from '@/api/tasks';
import { fetchSessionMessages, sendSessionMessage } from '@/api/sessions';
import { ChatWindow } from '@/components/chat/ChatWindow';
import { TaskEditForm } from '@/components/tasks/TaskEditForm';
import { StepEditableField } from '@/components/tasks/StepEditableField';
import { StatusBadge, PriorityBadge, Markdown } from '@/components/ui';
import type { TaskStep } from '@/types';
import { formatDate, formatDateTime, formatRelativeTime, formatTimeRange, isOverdue } from '@/utils/format';
import { confirmDelete } from '@/utils/dialog';

const { Title, Text } = Typography;

const TaskDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const screens = Grid.useBreakpoint(); // Get current breakpoint
  const [editingTask, setEditingTask] = useState(false);
  const [editingTitle, setEditingTitle] = useState(false);
  const [editingDescription, setEditingDescription] = useState(false);
  const [editingDueAt, setEditingDueAt] = useState(false);
  const [dueAtValue, setDueAtValue] = useState<dayjs.Dayjs | null>(null);
  const [dueAtPickerOpen, setDueAtPickerOpen] = useState(false);
  const [isAddStepModalVisible, setIsAddStepModalVisible] = useState(false);
  const [addStepForm] = Form.useForm();
  const [titleValue, setTitleValue] = useState('');
  const [descriptionValue, setDescriptionValue] = useState('');
  // Step editing states
  const [editingStepId, setEditingStepId] = useState<number | null>(null);
  const [editingStepField, setEditingStepField] = useState<'title' | 'detail' | null>(null);
  const [stepTitleValue, setStepTitleValue] = useState('');
  const [stepDetailValue, setStepDetailValue] = useState('');
  const { message } = App.useApp();

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
      message.error('发送失败，请稍后重试');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteTask,
    onSuccess: () => {
      message.success('任务已删除');
      navigate('/tasks');
    },
    onError: () => {
      message.error('删除失败，请稍后重试');
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
      message.error('更新失败，请稍后重试');
    },
  });

  const addStepMutation = useMutation({
    mutationFn: ({ taskId, stepData }: { taskId: string | number; stepData: any }) =>
      addTaskStep(taskId, stepData),
    onSuccess: () => {
      message.success('步骤添加成功');
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
      setIsAddStepModalVisible(false);
      addStepForm.resetFields();
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '添加失败，请稍后重试');
    },
  });

  const deleteStepMutation = useMutation({
    mutationFn: ({ taskId, stepId }: { taskId: string | number; stepId: string | number }) =>
      deleteTaskStep(taskId, stepId),
    onSuccess: () => {
      message.success('步骤已删除');
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '删除失败，请稍后重试');
    },
  });

  const updateStepMutation = useMutation({
    mutationFn: ({ stepId, fields }: { stepId: number; fields: any }) =>
      updateTaskStep(id!, stepId, fields),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', id] });
    },
    onError: (err: any) => {
      console.error(err);
      message.error('更新失败，请稍后重试');
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

  const handleStartEditStep = (step: TaskStep, field: 'title' | 'detail') => {
    setEditingStepId(step.id);
    setEditingStepField(field);
    if (field === 'title') {
      setStepTitleValue(step.title);
    } else {
      setStepDetailValue(step.detail || '');
    }
  };

  const handleUpdateStepTitle = async (step: TaskStep) => {
    if (stepTitleValue.trim() === '') {
      message.error('标题不能为空');
      return;
    }
    if (stepTitleValue !== step.title) {
      try {
        await updateStepMutation.mutateAsync({
          stepId: step.id,
          fields: { title: stepTitleValue }
        });
        message.success('步骤标题已更新');
      } catch (error) {
        // 错误已在 mutation 中处理
        return;
      }
    }
    setEditingStepId(null);
    setEditingStepField(null);
  };

  const handleUpdateStepDetail = async (step: TaskStep) => {
    if (stepDetailValue !== (step.detail || '')) {
      try {
        await updateStepMutation.mutateAsync({
          stepId: step.id,
          fields: { detail: stepDetailValue }
        });
        message.success('步骤详情已更新');
      } catch (error) {
        // 错误已在 mutation 中处理
        return;
      }
    }
    setEditingStepId(null);
    setEditingStepField(null);
  };

  const handleCancelEditStep = () => {
    setEditingStepId(null);
    setEditingStepField(null);
    setStepTitleValue('');
    setStepDetailValue('');
  };

  const handleUpdateDueAt = async () => {
    try {
      await updateTaskMutation.mutateAsync({
        dueAt: dueAtValue ? dueAtValue.toISOString() : null
      });
      message.success('截止时间已更新');
      setEditingDueAt(false);
      setDueAtPickerOpen(false);
    } catch (error) {
      // 错误已在mutation中处理
    }
  };

  const handleDueDateChange = (date: dayjs.Dayjs | null) => {
    setDueAtValue(date);
    // 选择日期后立即关闭面板
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

  const handleUpdateTitle = async () => {
    if (titleValue.trim() === '') {
      message.error('标题不能为空');
      return;
    }
    if (titleValue !== task.title) {
      try {
        await updateTaskMutation.mutateAsync({ title: titleValue });
        message.success('标题已更新');
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
        message.success('描述已更新');
      } catch (error) {
        // 错误已在mutation中处理
      }
    }
    setEditingDescription(false);
  };

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', paddingTop: 100 }}>
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
    <div>
      {/* Breadcrumb */}
      <Breadcrumb
        style={{ marginBottom: 16 }}
        items={[
          { title: <a onClick={() => navigate('/tasks')}>任务列表</a> },
          { title: task.title },
        ]}
      />

      {/* Responsive layout */}
      <Row gutter={[24, 24]}>
        {/* Left: Task info */}
        <Col xs={24} lg={14} xl={16}>
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
                {/* Title and badges */}
                <div style={{ marginBottom: 16 }}>
                  {/* Action buttons - always visible at top */}
                  <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 12 }}>
                    <Space wrap size="small">
                      <Button
                        icon={<ArrowLeftOutlined />}
                        onClick={() => navigate('/tasks')}
                        size="small"
                      >
                        返回
                      </Button>
                      <Button
                        icon={<EditOutlined />}
                        onClick={() => setEditingTask(true)}
                        size="small"
                      >
                        编辑
                      </Button>
                      <Button
                        danger
                        icon={<DeleteOutlined />}
                        onClick={handleDeleteTask}
                        size="small"
                      >
                        删除
                      </Button>
                    </Space>
                  </div>

                  {/* Title section */}
                  <div>
                    {editingTitle ? (
                      <Space.Compact style={{ width: '100%' }}>
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
                      <div
                        style={{
                          cursor: 'pointer',
                          padding: '8px',
                          borderRadius: 8,
                          transition: 'background-color 0.2s',
                        }}
                        onClick={() => {
                          setTitleValue(task.title);
                          setEditingTitle(true);
                        }}
                        onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f5f5f5'}
                        onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                      >
                        <Title
                          level={3}
                          style={{
                            margin: 0,
                            fontSize: 'clamp(18px, 4vw, 24px)',
                            wordBreak: 'break-word',
                            lineHeight: 1.3,
                          }}
                          ellipsis={{ tooltip: task.title }}
                        >
                          {task.title}
                        </Title>
                      </div>
                    )}
                    <Space wrap style={{ marginTop: 12 }}>
                      <StatusBadge status={task.status} />
                      <PriorityBadge priority={task.priority} />
                      {task.isFocusToday && (
                        <Tag
                          color="gold"
                          style={{
                            borderRadius: '4px',
                            fontWeight: 500,
                            fontSize: '12px',
                            padding: '2px 8px'
                          }}
                        >
                          今日聚焦
                        </Tag>
                      )}
                    </Space>
                  </div>
                </div>

                <Divider />

                {/* Description with markdown */}
                <div style={{ marginBottom: 24 }}>
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
                        padding: '12px',
                        borderRadius: 8,
                        minHeight: 60,
                        border: '1px solid #e8e8eb'
                      }}
                    >
                      {task.description ? (
                        <Markdown content={task.description} />
                      ) : (
                        <Text type="secondary" italic>点击添加描述...</Text>
                      )}
                    </div>
                  )}
                </div>

                {/* Time info */}
                <Descriptions size="small" column={{ xs: 1, sm: 2 }} style={{ marginBottom: 24 }}>
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
                        <Button
                          type="primary"
                          size="small"
                          icon={<CheckOutlined />}
                          onClick={handleUpdateDueAt}
                        />
                        <Button
                          size="small"
                          icon={<CloseOutlined />}
                          onClick={handleCancelEditDueAt}
                        />
                      </Space.Compact>
                    ) : (
                      <div
                        onClick={handleStartEditDueAt}
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

                {/* Steps */}
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
                    renderItem={(step: TaskStep, index: number) => {
                      // Use Ant Design's breakpoint system for responsive design
                      const isMobile = !screens.md; // md is 768px and above
                      const isEditingTitle = editingStepId === step.id && editingStepField === 'title';
                      const isEditingDetail = editingStepId === step.id && editingStepField === 'detail';

                      return (
                        <div
                          key={step.id}
                          style={{
                            borderBottom: index < (task.steps?.length || 0) - 1 ? '1px solid #f0f0f0' : 'none',
                            padding: '12px 0',
                          }}
                        >
                          {/* Mobile layout: vertical */}
                          {isMobile ? (
                            <div>
                              <div style={{ display: 'flex', alignItems: 'flex-start', gap: '8px' }}>
                                <Checkbox checked={step.status === 'done'} disabled style={{ marginTop: 2, flexShrink: 0 }} />
                                <div style={{ flex: 1, minWidth: 0 }}>
                                  {/* Title and detail with edit support */}
                                  <div>
                                    {/* Title editing or display */}
                                    {isEditingTitle ? (
                                      <Space.Compact style={{ width: '100%', marginBottom: 4 }}>
                                        <Input
                                          value={stepTitleValue}
                                          onChange={(e) => setStepTitleValue(e.target.value)}
                                          onPressEnter={() => handleUpdateStepTitle(step)}
                                          autoFocus
                                          size="small"
                                        />
                                        <Button
                                          type="primary"
                                          icon={<CheckOutlined />}
                                          onClick={() => handleUpdateStepTitle(step)}
                                          size="small"
                                        />
                                        <Button
                                          icon={<CloseOutlined />}
                                          onClick={handleCancelEditStep}
                                          size="small"
                                        />
                                      </Space.Compact>
                                    ) : (
                                      <div
                                        style={{
                                          cursor: 'pointer',
                                          display: 'inline-block',
                                          padding: '2px 4px',
                                          borderRadius: 4,
                                          marginBottom: step.detail ? 4 : 0,
                                          transition: 'background-color 0.2s',
                                        }}
                                        onClick={() => handleStartEditStep(step, 'title')}
                                        onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f5f5f5'}
                                        onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                                      >
                                        <Text strong style={{ fontSize: 'clamp(0.875rem, 1.5vw, 1rem)' }}>{step.title}</Text>
                                      </div>
                                    )}

                                    {/* Detail editing or display */}
                                    {step.detail && (
                                      isEditingDetail ? (
                                        <div style={{ marginBottom: 4 }}>
                                          <Input.TextArea
                                            value={stepDetailValue}
                                            onChange={(e) => setStepDetailValue(e.target.value)}
                                            onPressEnter={() => handleUpdateStepDetail(step)}
                                            autoSize={{ minRows: 1, maxRows: 4 }}
                                            autoFocus
                                            size="small"
                                            style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}
                                          />
                                          <Space style={{ marginTop: 4 }}>
                                            <Button
                                              type="primary"
                                              size="small"
                                              icon={<CheckOutlined />}
                                              onClick={() => handleUpdateStepDetail(step)}
                                            >
                                              保存
                                            </Button>
                                            <Button
                                              size="small"
                                              icon={<CloseOutlined />}
                                              onClick={handleCancelEditStep}
                                            >
                                              取消
                                            </Button>
                                          </Space>
                                        </div>
                                      ) : (
                                        <div
                                          style={{
                                            cursor: 'pointer',
                                            marginBottom: 4,
                                            padding: '2px 4px',
                                            borderRadius: 4,
                                            transition: 'background-color 0.2s',
                                          }}
                                          onClick={() => handleStartEditStep(step, 'detail')}
                                          onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f5f5f5'}
                                          onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                                        >
                                          <Text type="secondary" style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}>{step.detail}</Text>
                                        </div>
                                      )
                                    )}
                                  </div>

                                  {/* All tags in one row */}
                                  <Space size={6} wrap style={{ marginTop: 6 }}>
                                    {/* Status */}
                                    <div
                                      style={{
                                        padding: '2px 6px',
                                        borderRadius: 4,
                                        backgroundColor: step.status === 'done' ? '#52c41a' : step.status === 'in_progress' ? '#faad14' : step.status === 'blocked' ? '#ff4d4f' : '#1890ff',
                                        color: '#fff',
                                        fontSize: 'clamp(0.625rem, 1vw, 0.75rem)',
                                        display: 'inline-block',
                                      }}
                                    >
                                      {step.status === 'locked' ? '已锁定' :
                                       step.status === 'todo' ? '待办' :
                                       step.status === 'in_progress' ? '进行中' :
                                       step.status === 'done' ? '已完成' : '受阻'}
                                    </div>

                                    {/* Estimate */}
                                    {step.estimateMinutes && (
                                      <div
                                        style={{
                                          padding: '2px 6px',
                                          borderRadius: 4,
                                          backgroundColor: '#f5f5f5',
                                          fontSize: 'clamp(0.625rem, 1vw, 0.75rem)',
                                          display: 'inline-block',
                                        }}
                                      >
                                        <Space size={4}>
                                          <Text type="secondary" style={{ fontSize: 'clamp(0.625rem, 1vw, 0.75rem)' }}>
                                            预计耗时：
                                          </Text>
                                          <ClockCircleOutlined style={{ fontSize: 'clamp(0.625rem, 1vw, 0.75rem)', color: '#8c8c8c' }} />
                                          <Text type="secondary" style={{ fontSize: 'clamp(0.625rem, 1vw, 0.75rem)' }}>
                                            {step.estimateMinutes >= 480 && step.estimateMinutes % 480 === 0
                                              ? `${step.estimateMinutes / 480}天`
                                              : step.estimateMinutes >= 60
                                              ? step.estimateMinutes % 60 === 0
                                                ? `${step.estimateMinutes / 60}小时`
                                                : `${(step.estimateMinutes / 60).toFixed(1)}小时`
                                              : `${step.estimateMinutes}分钟`}
                                          </Text>
                                        </Space>
                                      </div>
                                    )}

                                    {/* Planned time */}
                                    {step.plannedStart && step.plannedEnd && (
                                      <Tooltip title={formatTimeRange(step.plannedStart, step.plannedEnd)}>
                                        <Space size={4}>
                                          <CalendarOutlined style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)', color: '#8c8c8c' }} />
                                          <Text type="secondary" style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)' }}>
                                            {formatDate(step.plannedStart)} - {formatDate(step.plannedEnd)}
                                          </Text>
                                        </Space>
                                      </Tooltip>
                                    )}

                                    {/* Completed time */}
                                    {step.completedAt && (
                                      <Tooltip title={`完成于 ${formatDateTime(step.completedAt)}`}>
                                        <Space size={4}>
                                          <CheckCircleOutlined style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)', color: '#52c41a' }} />
                                          <Text type="secondary" style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)' }}>
                                            {formatRelativeTime(step.completedAt)}
                                          </Text>
                                        </Space>
                                      </Tooltip>
                                    )}

                                    {/* Delete button */}
                                    <Button
                                      type="text"
                                      danger
                                      size="small"
                                      icon={<DeleteOutlined />}
                                      onClick={() => handleDeleteStep(step.id, step.title)}
                                      style={{ padding: '4px 8px' }}
                                    />
                                  </Space>
                                </div>
                              </div>
                            </div>
                          ) : (
                            /* Desktop layout: horizontal */
                            <div style={{ display: 'flex', alignItems: 'flex-start', gap: '8px' }}>
                              <Checkbox checked={step.status === 'done'} disabled style={{ marginTop: 8, flexShrink: 0 }} />
                              <div style={{ flex: 1, minWidth: 0 }}>
                                {/* Step title and detail with edit support */}
                                <div>
                                  {/* Title editing or display */}
                                  {isEditingTitle ? (
                                    <Space.Compact style={{ width: '100%', marginBottom: 4 }}>
                                      <Input
                                        value={stepTitleValue}
                                        onChange={(e) => setStepTitleValue(e.target.value)}
                                        onPressEnter={() => handleUpdateStepTitle(step)}
                                        autoFocus
                                      />
                                      <Button
                                        type="primary"
                                        icon={<CheckOutlined />}
                                        onClick={() => handleUpdateStepTitle(step)}
                                      />
                                      <Button
                                        icon={<CloseOutlined />}
                                        onClick={handleCancelEditStep}
                                      />
                                    </Space.Compact>
                                  ) : (
                                    <div
                                      style={{
                                        cursor: 'pointer',
                                        display: 'inline-block',
                                        padding: '2px 4px',
                                        borderRadius: 4,
                                        marginBottom: step.detail ? 4 : 0,
                                        transition: 'background-color 0.2s',
                                      }}
                                      onClick={() => handleStartEditStep(step, 'title')}
                                      onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f5f5f5'}
                                      onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                                    >
                                      <Text strong style={{ fontSize: 'clamp(0.875rem, 1.5vw, 1rem)' }}>{step.title}</Text>
                                    </div>
                                  )}

                                  {/* Detail editing or display */}
                                  {step.detail && (
                                    isEditingDetail ? (
                                      <div style={{ marginBottom: 4 }}>
                                        <Input.TextArea
                                          value={stepDetailValue}
                                          onChange={(e) => setStepDetailValue(e.target.value)}
                                          autoSize={{ minRows: 1, maxRows: 4 }}
                                          autoFocus
                                          style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}
                                        />
                                        <Space style={{ marginTop: 4 }}>
                                          <Button
                                            type="primary"
                                            size="small"
                                            icon={<CheckOutlined />}
                                            onClick={() => handleUpdateStepDetail(step)}
                                          >
                                            保存
                                          </Button>
                                          <Button
                                            size="small"
                                            icon={<CloseOutlined />}
                                            onClick={handleCancelEditStep}
                                          >
                                            取消
                                          </Button>
                                        </Space>
                                      </div>
                                    ) : (
                                      <div
                                        style={{
                                          cursor: 'pointer',
                                          marginBottom: 4,
                                          padding: '2px 4px',
                                          borderRadius: 4,
                                          transition: 'background-color 0.2s',
                                        }}
                                        onClick={() => handleStartEditStep(step, 'detail')}
                                        onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f5f5f5'}
                                        onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                                      >
                                        <Text type="secondary" style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}>{step.detail}</Text>
                                      </div>
                                    )
                                  )}
                                </div>
                              </div>
                              {/* Right side: status, estimate, time info, delete button */}
                              <div
                                style={{
                                  display: 'flex',
                                  flexDirection: 'column',
                                  alignItems: 'flex-end',
                                  gap: '6px',
                                  flexShrink: 0,
                                }}
                              >
                                {/* Status and estimate row */}
                                <StepEditableField step={step} taskId={task.id} compact />
                                {/* Time info and delete button row */}
                                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap', justifyContent: 'flex-end' }}>
                                  {step.plannedStart && step.plannedEnd && (
                                    <Tooltip title={formatTimeRange(step.plannedStart, step.plannedEnd)}>
                                      <Space size={4}>
                                        <CalendarOutlined style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)', color: '#8c8c8c' }} />
                                        <Text type="secondary" style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)' }}>
                                          {formatDate(step.plannedStart)} - {formatDate(step.plannedEnd)}
                                        </Text>
                                      </Space>
                                    </Tooltip>
                                  )}
                                  {step.completedAt && (
                                    <Tooltip title={`完成于 ${formatDateTime(step.completedAt)}`}>
                                      <Space size={4}>
                                        <CheckCircleOutlined style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)', color: '#52c41a' }} />
                                        <Text type="secondary" style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)' }}>
                                          {formatRelativeTime(step.completedAt)}
                                        </Text>
                                      </Space>
                                    </Tooltip>
                                  )}
                                  <Button
                                    type="text"
                                    danger
                                    size="small"
                                    icon={<DeleteOutlined />}
                                    onClick={() => handleDeleteStep(step.id, step.title)}
                                    style={{ flexShrink: 0, padding: '4px 8px' }}
                                  />
                                </div>
                              </div>
                            </div>
                          )}
                        </div>
                      );
                    }}
                  />
                </div>
              </>
            )}
          </Card>
        </Col>

        {/* Right: Chat */}
        <Col xs={24} lg={10} xl={8}>
          <Card
            title="小奇（任务）"
            styles={{ body: { padding: 0 } }}
            style={{ position: 'sticky', top: 24 }}
          >
            <ChatWindow
              messages={messages}
              onSend={(content) => sendMutation.mutateAsync(content)}
              sending={sendMutation.isPending}
              height={600}
            />
          </Card>
        </Col>
      </Row>

      {/* Add step modal */}
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
