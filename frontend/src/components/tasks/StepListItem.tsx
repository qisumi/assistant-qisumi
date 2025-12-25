import React from 'react';
import { Typography, Space, Button, Input, Grid, App, Tooltip } from 'antd';
import {
  DeleteOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  CloseOutlined,
} from '@ant-design/icons';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { StepEditableField } from '@/components/tasks/StepEditableField';
import type { TaskStep } from '@/types';
import { updateTaskStep } from '@/api/tasks';
import { formatDate, formatDateTime, formatRelativeTime, formatTimeRange } from '@/utils/format';

const { Text } = Typography;
const screens = Grid.useBreakpoint();

interface StepListItemProps {
  step: TaskStep;
  taskId: number;
  index: number;
  totalSteps: number;
  onUpdate?: () => void;
  onDelete?: (stepId: number, stepTitle: string) => void;
}

export const StepListItem: React.FC<StepListItemProps> = ({
  step,
  taskId,
  index,
  totalSteps,
  onUpdate,
  onDelete,
}) => {
  const [editingTitle, setEditingTitle] = React.useState(false);
  const [editingDetail, setEditingDetail] = React.useState(false);
  const [titleValue, setTitleValue] = React.useState('');
  const [detailValue, setDetailValue] = React.useState('');
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const updateMutation = useMutation({
    mutationFn: ({ stepId, fields }: { stepId: number; fields: any }) =>
      updateTaskStep(taskId, stepId, fields),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', String(taskId)] });
      onUpdate?.();
    },
    onError: (err: any) => {
      console.error(err);
      message.error('更新失败，请稍后重试');
    },
  });

  const handleStartEditTitle = () => {
    setTitleValue(step.title);
    setEditingTitle(true);
  };

  const handleStartEditDetail = () => {
    setDetailValue(step.detail || '');
    setEditingDetail(true);
  };

  const handleUpdateTitle = async () => {
    if (titleValue.trim() === '') {
      message.error('标题不能为空');
      return;
    }
    if (titleValue !== step.title) {
      try {
        await updateMutation.mutateAsync({
          stepId: step.id,
          fields: { title: titleValue },
        });
        message.success('步骤标题已更新');
      } catch (error) {
        return;
      }
    }
    setEditingTitle(false);
  };

  const handleUpdateDetail = async () => {
    if (detailValue !== (step.detail || '')) {
      try {
        await updateMutation.mutateAsync({
          stepId: step.id,
          fields: { detail: detailValue },
        });
        message.success('步骤详情已更新');
      } catch (error) {
        return;
      }
    }
    setEditingDetail(false);
  };

  const handleCancelEdit = () => {
    setEditingTitle(false);
    setEditingDetail(false);
    setTitleValue('');
    setDetailValue('');
  };

  const isMobile = !screens.md;

  return (
    <div
      key={step.id}
      style={{
        borderBottom: index < totalSteps - 1 ? '1px solid #f0f0f0' : 'none',
        padding: '12px 0',
      }}
    >
      {isMobile ? (
        // Mobile layout: vertical
        <div>
          <div style={{ display: 'flex', alignItems: 'flex-start', gap: '8px' }}>
            <input type="checkbox" checked={step.status === 'done'} disabled style={{ marginTop: 2, flexShrink: 0 }} />
            <div style={{ flex: 1, minWidth: 0 }}>
              {/* Title editing or display */}
              {editingTitle ? (
                <Space.Compact style={{ width: '100%', marginBottom: 4 }}>
                  <Input
                    value={titleValue}
                    onChange={(e) => setTitleValue(e.target.value)}
                    onPressEnter={handleUpdateTitle}
                    autoFocus
                    size="small"
                  />
                  <Button type="primary" icon={<CheckOutlined />} onClick={handleUpdateTitle} size="small" />
                  <Button icon={<CloseOutlined />} onClick={handleCancelEdit} size="small" />
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
                  onClick={handleStartEditTitle}
                  onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = '#f5f5f5')}
                  onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
                >
                  <Text strong style={{ fontSize: 'clamp(0.875rem, 1.5vw, 1rem)' }}>
                    {step.title}
                  </Text>
                </div>
              )}

              {/* Detail editing or display */}
              {step.detail && (
                editingDetail ? (
                  <div style={{ marginBottom: 4 }}>
                    <Input.TextArea
                      value={detailValue}
                      onChange={(e) => setDetailValue(e.target.value)}
                      onPressEnter={handleUpdateDetail}
                      autoSize={{ minRows: 1, maxRows: 4 }}
                      autoFocus
                      size="small"
                      style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}
                    />
                    <Space style={{ marginTop: 4 }}>
                      <Button type="primary" size="small" icon={<CheckOutlined />} onClick={handleUpdateDetail}>
                        保存
                      </Button>
                      <Button size="small" icon={<CloseOutlined />} onClick={handleCancelEdit}>
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
                    onClick={handleStartEditDetail}
                    onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = '#f5f5f5')}
                    onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
                  >
                    <Text type="secondary" style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}>
                      {step.detail}
                    </Text>
                  </div>
                )
              )}
            </div>
          </div>

          {/* Tags and actions */}
          <Space size={6} wrap style={{ marginTop: 6, marginLeft: 24 }}>
            {/* Status badge */}
            <div
              style={{
                padding: '2px 6px',
                borderRadius: 4,
                backgroundColor:
                  step.status === 'done'
                    ? '#52c41a'
                    : step.status === 'in_progress'
                      ? '#faad14'
                      : step.status === 'blocked'
                        ? '#ff4d4f'
                        : '#1890ff',
                color: '#fff',
                fontSize: 'clamp(0.625rem, 1vw, 0.75rem)',
                display: 'inline-block',
              }}
            >
              {step.status === 'locked'
                ? '已锁定'
                : step.status === 'todo'
                  ? '待办'
                  : step.status === 'in_progress'
                    ? '进行中'
                    : step.status === 'done'
                      ? '已完成'
                      : '受阻'}
            </div>

            {/* Planned time */}
            {step.plannedStart && step.plannedEnd && (
              <Tooltip title={formatTimeRange(step.plannedStart, step.plannedEnd)}>
                <Text type="secondary" style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)' }}>
                  {formatDate(step.plannedStart)} - {formatDate(step.plannedEnd)}
                </Text>
              </Tooltip>
            )}

            {/* Completed time */}
            {step.completedAt && (
              <Tooltip title={`完成于 ${formatDateTime(step.completedAt)}`}>
                <Text type="secondary" style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)' }}>
                  {formatRelativeTime(step.completedAt)}
                </Text>
              </Tooltip>
            )}

            {/* Delete button */}
            {onDelete && (
              <Button
                type="text"
                danger
                size="small"
                icon={<DeleteOutlined />}
                onClick={() => onDelete(step.id, step.title)}
                style={{ padding: '4px 8px' }}
              />
            )}
          </Space>
        </div>
      ) : (
        // Desktop layout: horizontal
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: '8px' }}>
          <input type="checkbox" checked={step.status === 'done'} disabled style={{ marginTop: 8, flexShrink: 0 }} />
          <div style={{ flex: 1, minWidth: 0 }}>
            {/* Title editing or display */}
            {editingTitle ? (
              <Space.Compact style={{ width: '100%', marginBottom: 4 }}>
                <Input
                  value={titleValue}
                  onChange={(e) => setTitleValue(e.target.value)}
                  onPressEnter={handleUpdateTitle}
                  autoFocus
                />
                <Button type="primary" icon={<CheckOutlined />} onClick={handleUpdateTitle} />
                <Button icon={<CloseOutlined />} onClick={handleCancelEdit} />
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
                onClick={handleStartEditTitle}
                onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = '#f5f5f5')}
                onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
              >
                <Text strong style={{ fontSize: 'clamp(0.875rem, 1.5vw, 1rem)' }}>
                  {step.title}
                </Text>
              </div>
            )}

            {/* Detail editing or display */}
            {step.detail && (
              editingDetail ? (
                <div style={{ marginBottom: 4 }}>
                  <Input.TextArea
                    value={detailValue}
                    onChange={(e) => setDetailValue(e.target.value)}
                    autoSize={{ minRows: 1, maxRows: 4 }}
                    autoFocus
                    style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}
                  />
                  <Space style={{ marginTop: 4 }}>
                    <Button type="primary" size="small" icon={<CheckOutlined />} onClick={handleUpdateDetail}>
                      保存
                    </Button>
                    <Button size="small" icon={<CloseOutlined />} onClick={handleCancelEdit}>
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
                  onClick={handleStartEditDetail}
                  onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = '#f5f5f5')}
                  onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
                >
                  <Text type="secondary" style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}>
                    {step.detail}
                  </Text>
                </div>
              )
            )}
          </div>

          {/* Right side: status, time info, delete button */}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'flex-end',
              gap: '6px',
              flexShrink: 0,
            }}
          >
            <StepEditableField step={step} taskId={taskId} compact />
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap', justifyContent: 'flex-end' }}>
              {step.plannedStart && step.plannedEnd && (
                <Text
                  type="secondary"
                  style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)' }}
                  title={formatTimeRange(step.plannedStart, step.plannedEnd)}
                >
                  <CalendarOutlined style={{ marginRight: 4 }} />
                  {formatDate(step.plannedStart)} - {formatDate(step.plannedEnd)}
                </Text>
              )}
              {step.completedAt && (
                <Text
                  type="secondary"
                  style={{ fontSize: 'clamp(0.7rem, 1.1vw, 0.875rem)' }}
                  title={`完成于 ${formatDateTime(step.completedAt)}`}
                >
                  <CheckCircleOutlined style={{ marginRight: 4, color: '#52c41a' }} />
                  {formatRelativeTime(step.completedAt)}
                </Text>
              )}
              {onDelete && (
                <Button
                  type="text"
                  danger
                  size="small"
                  icon={<DeleteOutlined />}
                  onClick={() => onDelete(step.id, step.title)}
                  style={{ flexShrink: 0, padding: '4px 8px' }}
                />
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
