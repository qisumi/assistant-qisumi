import React, { useState } from 'react';
import { Input, Select, Typography, Space, Button, App, InputNumber } from 'antd';
import { CheckOutlined, CloseOutlined, ClockCircleOutlined } from '@ant-design/icons';
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
  const [estimateUnit, setEstimateUnit] = useState<'minutes' | 'hours' | 'days'>('minutes');

  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const updateMutation = useMutation({
    mutationFn: (fields: UpdateStepFields) => updateTaskStep(taskId, step.id, fields),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['taskDetail', String(taskId)] });
      onUpdate?.();
    },
    onError: (err: any) => {
      console.error(err);
      message.error('更新失败，请稍后重试');
    },
  });

  const handleUpdateTitle = async () => {
    if (titleValue.trim() === '') {
      message.error('标题不能为空');
      return;
    }
    if (titleValue !== step.title) {
      await updateMutation.mutateAsync({ title: titleValue });
      message.success('标题已更新');
    }
    setEditingTitle(false);
  };

  const handleUpdateDetail = async () => {
    if (detailValue !== step.detail) {
      await updateMutation.mutateAsync({ detail: detailValue });
      message.success('详情已更新');
    }
    setEditingDetail(false);
  };

  const handleUpdateStatus = async (newStatus: StepStatus) => {
    if (newStatus !== step.status) {
      await updateMutation.mutateAsync({ status: newStatus });
      message.success('状态已更新');
    }
    setStatusValue(newStatus);
    setEditingStatus(false);
  };

  const handleUpdateEstimate = async () => {
    if (estimateValue !== step.estimateMinutes) {
      await updateMutation.mutateAsync({ estimateMinutes: estimateValue ?? undefined });
      message.success('预计耗时已更新');
    }
    setEditingEstimate(false);
  };

  // 将分钟数转换为最合适的显示单位
  const formatEstimate = (minutes: number | null): { value: number; unit: string; label: string } => {
    if (!minutes) return { value: 0, unit: 'minutes', label: '未设置' };

    // 优先级：天 > 小时 > 分钟
    if (minutes >= 480 && minutes % 480 === 0) { // 8小时为1天，且是整数天
      const days = minutes / 480;
      return { value: days, unit: 'days', label: `${days}天` };
    } else if (minutes >= 60) { // 至少1小时
      const hours = Math.round(minutes / 60);
      // 如果是整数小时，显示小时
      if (minutes % 60 === 0) {
        return { value: hours, unit: 'hours', label: `${hours}小时` };
      } else {
        // 如果不是整数小时，保留1位小数
        const hoursDecimal = (minutes / 60).toFixed(1);
        return { value: parseFloat(hoursDecimal), unit: 'hours', label: `${hoursDecimal}小时` };
      }
    } else {
      return { value: minutes, unit: 'minutes', label: `${minutes}分钟` };
    }
  };

  // 将用户输入的值和单位转换为分钟
  const convertToMinutes = (value: number | null, unit: 'minutes' | 'hours' | 'days'): number | null => {
    if (value === null || value === undefined) return null;

    switch (unit) {
      case 'days':
        return value * 480; // 1天 = 8小时 = 480分钟
      case 'hours':
        return Math.round(value * 60);
      case 'minutes':
      default:
        return value;
    }
  };

  // 保存预计耗时时的处理
  const handleSaveEstimate = async () => {
    const minutes = convertToMinutes(estimateValue, estimateUnit);
    if (minutes !== step.estimateMinutes) {
      await updateMutation.mutateAsync({ estimateMinutes: minutes ?? undefined });
      message.success('预计耗时已更新');
    }
    setEditingEstimate(false);
  };

  // 开始编辑时初始化单位和值
  const handleStartEditEstimate = () => {
    if (step.estimateMinutes) {
      const formatted = formatEstimate(step.estimateMinutes);
      setEstimateValue(formatted.value);
      setEstimateUnit(formatted.unit as 'minutes' | 'hours' | 'days');
    } else {
      setEstimateValue(null);
      setEstimateUnit('minutes');
    }
    setEditingEstimate(true);
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
    <div>
      {/* 标题编辑 */}
      <div style={{ marginBottom: 4 }}>
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
            <Text strong style={{ fontSize: 'clamp(0.875rem, 1.5vw, 1rem)' }}>{step.title}</Text>
          </div>
        )}
      </div>

      {/* 详情编辑 */}
      {editingDetail ? (
        <div style={{ marginBottom: 4 }}>
          <TextArea
            value={detailValue}
            onChange={(e) => setDetailValue(e.target.value)}
            rows={2}
            autoFocus
            style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}
          />
          <Space style={{ marginTop: 4 }}>
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
        step.detail && (
          <div
            onClick={() => setEditingDetail(true)}
            style={{ cursor: 'pointer', marginBottom: 4 }}
          >
            <Text type="secondary" style={{ fontSize: 'clamp(0.75rem, 1.2vw, 0.875rem)' }}>{step.detail}</Text>
          </div>
        )
      )}

      {/* 状态和预计耗时 */}
      <Space size="small" wrap>
        {editingStatus ? (
          <Select
            value={statusValue}
            onChange={handleUpdateStatus}
            style={{ width: 100 }}
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
              padding: 'clamp(2px, 0.3vw, 4px) clamp(4px, 0.6vw, 8px)',
              borderRadius: 4,
              backgroundColor: getStepStatusColor(step.status),
              color: '#fff',
              fontSize: 'clamp(0.625rem, 1vw, 0.75rem)',
            }}
          >
            {stepStatusLabels[step.status]}
          </Space>
        )}

        {/* 预计耗时编辑 */}
        {editingEstimate ? (
          <Space.Compact style={{ flexWrap: 'wrap' }}>
            <InputNumber
              value={estimateValue}
              onChange={(value) => setEstimateValue(value)}
              min={0.1}
              step={estimateUnit === 'days' ? 0.5 : 1}
              placeholder="数量"
              style={{ width: 80 }}
              controls={false}
              autoFocus
              size="small"
            />
            <Select
              value={estimateUnit}
              onChange={(value) => setEstimateUnit(value)}
              size="small"
              style={{ width: 75, minHeight: 24 }}
              popupMatchSelectWidth={false}
              dropdownStyle={{ minWidth: 80 }}
            >
              <Select.Option value="minutes">分钟</Select.Option>
              <Select.Option value="hours">小时</Select.Option>
              <Select.Option value="days">天</Select.Option>
            </Select>
            <Button
              type="primary"
              size="small"
              icon={<CheckOutlined />}
              onClick={handleSaveEstimate}
              style={{ height: 24 }}
            />
            <Button
              size="small"
              icon={<CloseOutlined />}
              onClick={() => {
                setEstimateValue(step.estimateMinutes || null);
                setEditingEstimate(false);
              }}
              style={{ height: 24 }}
            />
          </Space.Compact>
        ) : (
          <div
            onClick={handleStartEditEstimate}
            style={{
              cursor: 'pointer',
              display: 'inline-block',
              padding: '2px 6px',
              borderRadius: 4,
              transition: 'background-color 0.2s',
            }}
            onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f5f5f5'}
            onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
          >
            <Space size={4}>
              <ClockCircleOutlined style={{ fontSize: 'clamp(0.625rem, 1vw, 0.75rem)', color: step.estimateMinutes ? '#8c8c8c' : '#d9d9d9' }} />
              <Text style={{ fontSize: 'clamp(0.625rem, 1vw, 0.75rem)', color: step.estimateMinutes ? '#8c8c8c' : '#d9d9d9' }}>
                {formatEstimate(step.estimateMinutes).label}
              </Text>
            </Space>
          </div>
        )}
      </Space>
    </div>
  );
};
