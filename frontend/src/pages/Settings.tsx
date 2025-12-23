import React, { useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, Form, Input, Button, Space, Typography, message as antdMessage, Divider, Alert } from 'antd';
import { SettingOutlined, LockOutlined, GlobalOutlined } from '@ant-design/icons';

import { fetchLLMSettings, updateLLMSettings, type LLMSettings } from '@/api/settings';

const { Title, Paragraph, Text } = Typography;

const Settings: React.FC = () => {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  const { data: settings, isLoading } = useQuery({
    queryKey: ['llmSettings'],
    queryFn: fetchLLMSettings,
  });

  const updateMutation = useMutation({
    mutationFn: (values: LLMSettings) => updateLLMSettings(values),
    onSuccess: () => {
      antdMessage.success('设置已保存');
      queryClient.invalidateQueries({ queryKey: ['llmSettings'] });
    },
    onError: (err: any) => {
      console.error(err);
      antdMessage.error('保存失败，请检查输入或稍后重试');
    },
  });

  useEffect(() => {
    if (settings) {
      form.setFieldsValue({
        base_url: settings.base_url,
        model: settings.model,
        // api_key 通常不回显
      });
    }
  }, [settings, form]);

  const onFinish = (values: LLMSettings) => {
    updateMutation.mutate(values);
  };

  return (
    <div style={{ padding: 24, display: 'flex', justifyContent: 'center' }}>
      <Card style={{ maxWidth: 800, width: '100%' }}>
        <Space direction="vertical" style={{ width: '100%', marginBottom: 24 }}>
          <Title level={2}>
            <SettingOutlined style={{ marginRight: 12 }} />
            系统设置
          </Title>
          <Paragraph type="secondary">
            配置您的个人偏好和 LLM 服务接入信息。
          </Paragraph>
        </Space>

        <Divider orientation="left">LLM 配置</Divider>
        <Alert
          message="安全提示"
          description="您的 API 密钥将被加密存储在服务器上。为了安全起见，获取设置时不会返回已保存的密钥。"
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            base_url: 'https://api.openai.com/v1',
            model: 'gpt-3.5-turbo',
          }}
        >
          <Form.Item
            label="API 基础地址"
            name="base_url"
            rules={[{ required: true, message: '请输入 API 基础地址' }]}
            tooltip="OpenAI 兼容接口的基础地址，例如 https://api.openai.com/v1"
          >
            <Input prefix={<GlobalOutlined />} placeholder="https://api.openai.com/v1" />
          </Form.Item>

          <Form.Item
            label="API 密钥"
            name="api_key"
            rules={[{ required: !settings, message: '请输入 API 密钥' }]}
            tooltip={settings ? "留空表示不修改已保存的密钥" : "请输入您的 API 密钥"}
          >
            <Input.Password prefix={<LockOutlined />} placeholder={settings ? "••••••••••••••••" : "sk-..."} />
          </Form.Item>

          <Form.Item
            label="默认模型"
            name="model"
            rules={[{ required: true, message: '请输入模型名称' }]}
          >
            <Input placeholder="gpt-3.5-turbo, gpt-4, qwen-max 等" />
          </Form.Item>

          <Form.Item>
            <Button 
              type="primary" 
              htmlType="submit" 
              loading={updateMutation.isPending}
              block
            >
              保存配置
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Settings;
