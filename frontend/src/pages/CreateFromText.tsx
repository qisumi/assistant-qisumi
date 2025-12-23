import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { Button, Card, Form, Input, Space, Typography, message as antdMessage } from 'antd';

import { createTaskFromText } from '@/api/tasks';

const { TextArea } = Input;
const { Title, Text, Paragraph } = Typography;

const CreateFromText: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [length, setLength] = useState(0);

  const createMutation = useMutation({
    mutationFn: (rawText: string) => createTaskFromText(rawText),
    onSuccess: (data) => {
      antdMessage.success('已根据文本生成任务');
      // 假设后端返回 { task: Task, session: Session }
      const taskId = data.task.id;
      navigate(`/tasks/${taskId}`);
    },
    onError: (err: any) => {
      console.error(err);
      antdMessage.error('生成任务失败，请稍后重试');
    },
  });

  const handleSubmit = async (values: { rawText: string }) => {
    const rawText = values.rawText?.trim();
    if (!rawText) {
      antdMessage.warning('请输入一些内容');
      return;
    }
    await createMutation.mutateAsync(rawText);
  };

  return (
    <div style={{ padding: 24, display: 'flex', justifyContent: 'center' }}>
      <Card style={{ maxWidth: 800, width: '100%' }}>
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div>
            <Title level={3} style={{ marginBottom: 8 }}>
              从文本创建任务
            </Title>
            <Paragraph type="secondary">
              你可以直接粘贴会议记录、聊天记录或备忘录，系统会自动帮你生成一个任务，并拆解为多个步骤。
            </Paragraph>
          </div>

          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
          >
            <Form.Item
              label={
                <Space>
                  <span>原始文本</span>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    （当前字数：{length}）
                  </Text>
                </Space>
              }
              name="rawText"
              rules={[{ required: true, message: '请粘贴或输入文本' }]}
            >
              <TextArea
                rows={12}
                placeholder="例如：本周需要完成 AIGC 小论文，周一前看 10 篇文章..."
                onChange={(e) => setLength(e.target.value.length)}
                disabled={createMutation.isPending}
              />
            </Form.Item>

            <Form.Item style={{ textAlign: 'right' }}>
              <Space>
                <Button onClick={() => form.resetFields()} disabled={createMutation.isPending}>
                  清空
                </Button>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={createMutation.isPending}
                >
                  生成任务
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </Space>
      </Card>
    </div>
  );
};

export default CreateFromText;
