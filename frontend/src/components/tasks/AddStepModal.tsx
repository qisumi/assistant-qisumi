import React from 'react';
import { Modal, Form, Input, InputNumber, App } from 'antd';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addTaskStep } from '@/api/tasks';

interface AddStepModalProps {
  open: boolean;
  taskId: string | number;
  stepsCount: number;
  onCancel: () => void;
  onSuccess?: () => void;
}

export const AddStepModal: React.FC<AddStepModalProps> = ({
  open,
  taskId,
  stepsCount,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const addMutation = useMutation({
    mutationFn: ({ taskId, stepData }: { taskId: string | number; stepData: any }) =>
      addTaskStep(taskId, stepData),
    onSuccess: () => {
      message.success('步骤添加成功');
      queryClient.invalidateQueries({ queryKey: ['taskDetail', String(taskId)] });
      form.resetFields();
      onSuccess?.();
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '添加失败，请稍后重试');
    },
  });

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      const stepData = {
        title: values.title,
        detail: values.detail || '',
        orderIndex: stepsCount,
        status: 'todo',
        estimateMinutes: values.estimateMinutes || null,
      };
      addMutation.mutate({ taskId, stepData });
    } catch (error) {
      // Form validation failed
    }
  };

  const handleCancel = () => {
    form.resetFields();
    onCancel();
  };

  return (
    <Modal
      title="添加步骤"
      open={open}
      onOk={handleOk}
      onCancel={handleCancel}
      okText="添加"
      cancelText="取消"
      confirmLoading={addMutation.isPending}
    >
      <Form form={form} layout="vertical" style={{ marginTop: 24 }}>
        <Form.Item name="title" label="步骤标题" rules={[{ required: true, message: '请输入步骤标题' }]}>
          <Input placeholder="例如：准备项目材料" />
        </Form.Item>

        <Form.Item name="detail" label="步骤详情">
          <Input.TextArea rows={3} placeholder="描述这个步骤的详细内容（可选）" />
        </Form.Item>

        <Form.Item name="estimateMinutes" label="预计耗时（分钟）">
          <InputNumber min={1} style={{ width: '100%' }} placeholder="例如：30" />
        </Form.Item>
      </Form>
    </Modal>
  );
};
