import React, { useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, Form, Input, Button, Typography, App, Alert, Space, Select } from 'antd';
import { SettingOutlined, LockOutlined, GlobalOutlined, SaveOutlined } from '@ant-design/icons';

import {
  fetchLLMSettings,
  updateLLMSettings,
  type LLMSettings,
  ThinkingType,
  ReasoningEffort,
  ThinkingTypeLabels,
  ReasoningEffortLabels,
} from '@/api/settings';
import { PageHeader } from '@/components/layout/PageHeader';

const { Title, Paragraph } = Typography;

const Settings: React.FC = () => {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const { data: settings } = useQuery({
    queryKey: ['llmSettings'],
    queryFn: fetchLLMSettings,
  });

  const updateMutation = useMutation({
    mutationFn: (values: LLMSettings) => updateLLMSettings(values),
    onSuccess: () => {
      message.success('设置已保存');
      queryClient.invalidateQueries({ queryKey: ['llmSettings'] });
    },
    onError: (err: any) => {
      console.error(err);
      message.error('保存失败，请检查输入或稍后重试');
    },
  });

  useEffect(() => {
    if (settings) {
      form.setFieldsValue({
        base_url: settings.base_url,
        model: settings.model,
        thinking_type: settings.thinking_type || ThinkingType.Auto,
        reasoning_effort: settings.reasoning_effort || ReasoningEffort.Medium,
        assistant_name: settings.assistant_name || '小奇',
        // api_key 通常不回显
      });
    }
  }, [settings, form]);

  const onFinish = (values: LLMSettings) => {
    updateMutation.mutate(values);
  };

  return (
    <div style={{ maxWidth: 800, margin: '0 auto' }}>
      <PageHeader
        title="系统设置"
        extra={
          <Button
            type="primary"
            icon={<SaveOutlined />}
            onClick={() => form.submit()}
            loading={updateMutation.isPending}
          >
            保存配置
          </Button>
        }
      />

      {/* Info card */}
      <Card style={{ marginBottom: 24 }}>
        <Space direction="vertical" size="small">
          <Title level={4} style={{ margin: 0 }}>
            <SettingOutlined style={{ marginRight: 8 }} />
            LLM 服务配置
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            配置您的 LLM 服务接入信息以启用 AI 助手功能。
          </Paragraph>
        </Space>
      </Card>

      {/* Security alert */}
      <Alert
        message="安全提示"
        description="您的 API 密钥将被加密存储在服务器上。为了安全起见，获取设置时不会返回已保存的密钥。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      {/* Helper info alert */}
      {settings && (
        <Alert
          message="功能提示"
          description="您可以在不填写 API 密钥的情况下，单独修改助手名称和其他设置。只有需要更换 API 密钥时才需要填写。"
          type="success"
          showIcon
          closable
          style={{ marginBottom: 24 }}
        />
      )}

      {/* Settings form */}
      <Card title="API 配置">
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            base_url: 'https://api.openai.com/v1',
            model: 'gpt-3.5-turbo',
            thinking_type: ThinkingType.Auto,
            reasoning_effort: ReasoningEffort.Medium,
            assistant_name: '小奇',
          }}
        >
          <Form.Item
            label={
              <Space>
                <GlobalOutlined />
                <span>API 基础地址</span>
              </Space>
            }
            name="base_url"
            rules={[{ required: true, message: '请输入 API 基础地址' }]}
            tooltip="OpenAI 兼容接口的基础地址"
          >
            <Input placeholder="https://api.openai.com/v1" style={{ borderRadius: '6px' }} />
          </Form.Item>

          <Form.Item
            label={
              <Space>
                <LockOutlined />
                <span>API 密钥</span>
              </Space>
            }
            name="api_key"
            rules={[{ required: !settings, message: '请输入 API 密钥' }]}
            tooltip={
              <div>
                <div>{settings ? "留空表示不修改已保存的密钥" : "请输入您的 API 密钥"}</div>
                {!settings && <div style={{ marginTop: 4, color: '#ff4d4f' }}>首次配置必须提供 API 密钥</div>}
                {settings && <div style={{ marginTop: 4, color: '#52c41a' }}>您可以只修改助手名称而不用填写API密钥</div>}
              </div>
            }
          >
            <Input.Password
              placeholder={settings ? "留空表示不修改" : "sk-..."}
              style={{ borderRadius: '6px' }}
            />
          </Form.Item>

          <Form.Item
            label="默认模型"
            name="model"
            rules={[{ required: true, message: '请输入模型名称' }]}
            tooltip="例如: gpt-3.5-turbo, gpt-4, qwen-plus, qwen-max 等"
          >
            <Input placeholder="gpt-3.5-turbo" style={{ borderRadius: '6px' }} />
          </Form.Item>

          <Form.Item
            label="助手名称"
            name="assistant_name"
            rules={[{ required: true, message: '请输入助手名称' }]}
            tooltip={
              <div>
                <div>设置您的AI助手的名称</div>
                {settings && <div style={{ marginTop: 4, color: '#52c41a' }}>您可以随时修改助手名称，无需填写API密钥</div>}
              </div>
            }
          >
            <Input placeholder="小奇" style={{ borderRadius: '6px' }} />
          </Form.Item>

          <Form.Item
            label="深度思考"
            name="thinking_type"
            rules={[{ required: true, message: '请选择深度思考模式' }]}
            tooltip="控制模型是否使用深度思考能力"
          >
            <Select
              placeholder="请选择深度思考模式"
              style={{ borderRadius: '6px' }}
              options={[
                { label: ThinkingTypeLabels[ThinkingType.Disabled], value: ThinkingType.Disabled },
                { label: ThinkingTypeLabels[ThinkingType.Enabled], value: ThinkingType.Enabled },
                { label: ThinkingTypeLabels[ThinkingType.Auto], value: ThinkingType.Auto },
              ]}
            />
          </Form.Item>

          <Form.Item
            label="思考强度"
            name="reasoning_effort"
            rules={[{ required: true, message: '请选择思考强度' }]}
            tooltip="控制模型深度思考的强度级别"
          >
            <Select
              placeholder="请选择思考强度"
              style={{ borderRadius: '6px' }}
              options={[
                { label: ReasoningEffortLabels[ReasoningEffort.Minimal], value: ReasoningEffort.Minimal },
                { label: ReasoningEffortLabels[ReasoningEffort.Low], value: ReasoningEffort.Low },
                { label: ReasoningEffortLabels[ReasoningEffort.Medium], value: ReasoningEffort.Medium },
                { label: ReasoningEffortLabels[ReasoningEffort.High], value: ReasoningEffort.High },
              ]}
            />
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Settings;
