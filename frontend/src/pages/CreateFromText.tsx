import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { Button, Card, Form, Input, Space, Typography, App, Alert } from 'antd';
import { FileTextOutlined, BulbOutlined, ArrowLeftOutlined, ClearOutlined, ThunderboltOutlined } from '@ant-design/icons';

import { createTaskFromText } from '@/api/tasks';
import { PageHeader } from '@/components/layout/PageHeader';

const { TextArea } = Input;
const { Title, Paragraph, Text } = Typography;

const CreateFromText: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [length, setLength] = useState(0);
  const { message } = App.useApp();

  const createMutation = useMutation({
    mutationFn: (rawText: string) => createTaskFromText(rawText),
    onSuccess: (data) => {
      message.success('已根据文本生成任务');
      // 只提取需要的字段，避免可能的循环引用
      const taskId = data.task.id;
      navigate(`/tasks/${taskId}`);
    },
    onError: (err: any) => {
      console.error(err);
      message.error('生成任务失败，请稍后重试');
    },
  });

  const handleSubmit = async (values: { rawText: string }) => {
    const rawText = values.rawText?.trim();
    if (!rawText) {
      message.warning('请输入一些内容');
      return;
    }
    await createMutation.mutateAsync(rawText);
  };

  const exampleText = `本周需要完成 AIGC 小论文，周一前看 10 篇相关文献；
周二整理论文大纲，和导师讨论；
周三开始撰写初稿，重点写相关工作部分；
周四完善实验部分，补充图表；
周五修改摘要和结论部分；
周末整理论文格式，准备提交材料。`;

  const handleUseExample = () => {
    form.setFieldValue('rawText', exampleText);
    setLength(exampleText.length);
  };

  const handleClear = () => {
    form.resetFields();
    setLength(0);
  };

  return (
    <div style={{ maxWidth: 900, margin: '0 auto' }}>
      <PageHeader
        title="从文本创建任务"
        extra={
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/tasks')}
          >
            返回
          </Button>
        }
      />

      {/* Info card */}
      <Card
        style={{
          marginBottom: 24,
          background: 'linear-gradient(135deg, #5E6AD2 0%, #722ed1 100%)',
          border: 'none',
          color: 'white'
        }}
      >
        <Space direction="vertical" size="small">
          <Space>
            <BulbOutlined style={{ fontSize: 24, color: '#faad14' }} />
            <Title level={4} style={{ margin: 0, color: 'white' }}>
              AI 智能任务生成
            </Title>
          </Space>
          <Paragraph style={{ color: 'rgba(255, 255, 255, 0.9)', marginBottom: 0 }}>
            粘贴会议记录、聊天记录或备忘录，系统会自动帮你生成任务并拆解为多个步骤。
          </Paragraph>
        </Space>
      </Card>

      {/* Input form */}
      <Card title="输入文本">
        <Alert
          message="提示"
          description="输入的文本越长、越具体，生成的任务质量越高。建议包含任务目标、关键步骤、时间等信息。"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />

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
                  （{length} 字符）
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
              style={{ borderRadius: '8px' }}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Space>
              <Button
                icon={<ClearOutlined />}
                onClick={handleClear}
                disabled={createMutation.isPending}
              >
                清空
              </Button>
              <Button
                type="default"
                icon={<FileTextOutlined />}
                onClick={handleUseExample}
                disabled={createMutation.isPending}
              >
                使用示例
              </Button>
              <div style={{ flex: 1 }} />
              <Button
                type="primary"
                htmlType="submit"
                loading={createMutation.isPending}
                icon={<ThunderboltOutlined />}
              >
                生成任务
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default CreateFromText;
