import React, { useState } from 'react';
import { Input, Select, Typography, Space, Button, message as antdMessage, InputNumber } from 'antd';
import { EditOutlined, CheckOutlined, CloseOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { TaskStep, StepStatus } from '@/types';
import { updateTaskStep, type UpdateStepFields } from '@/api/tasks';

const { Text } = Typography;
const { TextArea } = Input;

interface StepEditableFieldProps {
  step: TaskStep;
  taskId: number;
  onUpdate?: () => void;
}

export const StepEditableField: React.FC<StepEditableFieldProps> = ({ step, taskId, onUpdate }) => {
  const [editingTitle, setEditingTitle] = useState(false);
  const [editingDetail, setEditingDetail] = useState(false);
  const [editingStatus, setEditingStatus] = useState(false);
  const [editingEstimate, setEditingEstimate] = useState(false);
  
  const [titleValue, setTitleValue] = useState(step.title);
  const [detailValue, setDetailValue] = useState(step.detail);
  const [statusValue, setStatusValue] = useState<StepStatus>(step.status);
  const [estimateValue, setEstimateValue] = useState<number | null>(step.estimateMinutes || null);
  
  const queryClient = useQueryClient();

  const updateMutation = useMutation({
    mutationFn: (fields: UpdateStepFields) => updateTaskStep(taskId, step.id, fields),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', String(taskId)] });
      onUpdate?.();
    },
    onError: (err: any) => {
      console.error(err);
      antdMessage.error('更新失败，请稍后重试');
    },
  });

  const handleUpdateTitle = async () => {
    if (titleValue.trim() === '') {
      antdMessage.error('标题不能为空');
      return;
    }
    if (titleValue !== step.title) {
      await updateMutation.mutateAsync({ title: titleValue });
      antdMessage.success('标题已更新');
    }
    setEditingTitle(false);
  };

  const handleUpdateDetail = async () => {
    if (detailValue !== step.detail) {
      await updateMutation.mutateAsync({ detail: detailValue });
      antdMessage.success('详情已更新');
    }
    setEditingDetail(false);
  };

  const handleUpdateStatus = async (newStatus: StepStatus) => {
    if (newStatus !== step.status) {
      await updateMutation.mutateAsync({ status: newStatus });
      antdMessage.success('状态已更新');
    }
    setStatusValue(newStatus);
    setEditingStatus(false);
  };

  const handleUpdateEstimate = async () => {
    if (estimateValue !== step.estimateMinutes) {
      await updateMutation.mutateAsync({ estimateMinutes: estimateValue ?? undefined });
      antdMessage.success('预计耗时已更新');
    }
    setEditingEstimate(false);
  };

  const stepStatusLabels: Record<string, string> = {
    locked: '已锁定',
    todo: '待办',
    in_progress: '进行中',
    done: '已完成',
    blocked: '受阻',
  };

  const getStepStatusColor = (status: string): string => {
    const colors: Record<string, string> = {
      locked: '#d9d9d9',
      todo: '#1890ff',
      in_progress: '#faad14',
      done: '#52c41a',
      blocked: '#ff4d4f',
    };
    return colors[status] || '#d9d9d9';
  };

  return (
    <div style={{ marginBottom: 12 }}>
      {/* 标题编辑 */}
      <div style={{ marginBottom: 8 }}>
        {editingTitle ? (
          <Space.Compact style={{ width: '100%' }}>
            <Input
              value={titleValue}
              onChange={(e) => setTitleValue(e.target.value)}
              onPressEnter={handleUpdateTitle}
              autoFocus
            />
            <Button
              type="primary"
              icon={<CheckOutlined />}
              onClick={handleUpdateTitle}
            />
            <Button
              icon={<CloseOutlined />}
              onClick={() => {
                setTitleValue(step.title);
                setEditingTitle(false);
              }}
            />
          </Space.Compact>
        ) : (
          <div
            onClick={() => setEditingTitle(true)}
            style={{
              cursor: 'pointer',
              display: 'inline-block',
              padding: '2px 4px',
              borderRadius: 4
            }}
          >
            <Text strong>{step.title}</Text>
          </div>
        )}
      </div>

      {/* 详情编辑 */}
      <div style={{ marginBottom: 8 }}>
        {editingDetail ? (
          <div>
            <TextArea
              value={detailValue}
              onChange={(e) => setDetailValue(e.target.value)}
              rows={3}
              autoFocus
            />
            <Space style={{ marginTop: 8 }}>
              <Button
                type="primary"
                size="small"
                icon={<CheckOutlined />}
                onClick={handleUpdateDetail}
              >
                保存
              </Button>
              <Button
                size="small"
                icon={<CloseOutlined />}
                onClick={() => {
                  setDetailValue(step.detail);
                  setEditingDetail(false);
                }}
              >
                取消
              </Button>
            </Space>
          </div>
        ) : (
          <div onClick={() => setEditingDetail(true)} style={{ cursor: 'pointer' }}>
            {step.detail ? (
              <Text type="secondary">{step.detail}</Text>
            ) : (
              <Text type="secondary" italic>点击添加详情...</Text>
            )}
          </div>
        )}
      </div>

      {/* 状态和预计耗时 */}
      <Space>
        {editingStatus ? (
          <Select
            value={statusValue}
            onChange={handleUpdateStatus}
            style={{ width: 120 }}
            autoFocus
            open
            onBlur={() => setEditingStatus(false)}
          >
            <Select.Option value="locked">已锁定</Select.Option>
            <Select.Option value="todo">待办</Select.Option>
            <Select.Option value="in_progress">进行中</Select.Option>
            <Select.Option value="done">已完成</Select.Option>
            <Select.Option value="blocked">受阻</Select.Option>
          </Select>
        ) : (
          <Space
            onClick={() => setEditingStatus(true)}
            style={{
              cursor: 'pointer',
              padding: '2px 8px',
              borderRadius: 4,
              backgroundColor: getStepStatusColor(step.status),
              color: '#fff',
              fontSize: '12px',
            }}
          >
            {stepStatusLabels[step.status]}
          </Space>
        )}
        
        {/* 预计耗时编辑 */}
        {editingEstimate ? (
          <Space.Compact>
            <InputNumber
              value={estimateValue}
              onChange={(value) => setEstimateValue(value)}
              min={1}
              placeholder="分钟"
              style={{ width: 100 }}
              autoFocus
              addonAfter="分钟"
            />
            <Button
              type="primary"
              size="small"
              icon={<CheckOutlined />}
              onClick={handleUpdateEstimate}
            />
            <Button
              size="small"
              icon={<CloseOutlined />}
              onClick={() => {
                setEstimateValue(step.estimateMinutes || null);
                setEditingEstimate(false);
              }}
            />
          </Space.Compact>
        ) : (
          <div
            onClick={() => setEditingEstimate(true)}
            style={{
              cursor: 'pointer',
              display: 'inline-block'
            }}
          >
            <Space size={4}>
              <ClockCircleOutlined style={{ fontSize: 12 }} />
              <Text style={{ fontSize: 12 }}>
                {step.estimateMinutes ? `预计 ${step.estimateMinutes} 分钟` : '设置预计耗时'}
              </Text>
            </Space>
          </div>
        )}
      </Space>
    </div>
  );
};
