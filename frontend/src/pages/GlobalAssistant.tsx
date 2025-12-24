import React, { useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, Button, Space, Spin, Typography, App } from 'antd';
import { BulbOutlined, PlusOutlined, MessageOutlined } from '@ant-design/icons';

import { ChatWindow } from '@/components/chat/ChatWindow';
import { getOrCreateGlobalSession, fetchSessionMessages, sendSessionMessage, clearSessionMessages } from '@/api/sessions';
import type { Message } from '@/types';
import { confirmAction } from '@/utils/dialog';
import { PageHeader } from '@/components/layout/PageHeader';
import { useResponsive } from '@/hooks';

const { Title, Text, Paragraph } = Typography;

const GlobalAssistant: React.FC = () => {
  const queryClient = useQueryClient();
  const { message } = App.useApp();
  const { isMobile } = useResponsive();

  // 1. 获取/创建 global session
  const {
    data: session,
    isLoading: loadingSession,
    isError: sessionError,
  } = useQuery({
    queryKey: ['globalSession'],
    queryFn: () => getOrCreateGlobalSession(),
  });

  const sessionId = session?.id;

  // 2. 拉取消息
  const { data: messagesData } = useQuery({
    queryKey: ['sessionMessages', sessionId],
    queryFn: () => fetchSessionMessages(sessionId!),
    enabled: !!sessionId,
  });

  const messages: Message[] = messagesData?.messages ?? [];

  // 3. 发送消息
  const sendMutation = useMutation({
    mutationFn: (content: string) => sendSessionMessage(sessionId!, content),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['sessionMessages', sessionId] });
    },
    onError: (err: any) => {
      console.error(err);
      message.error('发送失败，请稍后重试');
    },
  });

  // 4. 清空消息（开启新对话）
  const clearMutation = useMutation({
    mutationFn: () => clearSessionMessages(sessionId!),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['sessionMessages', sessionId] });
      message.success('已开启新对话');
    },
    onError: (err: any) => {
      console.error(err);
      message.error('开启新对话失败，请稍后重试');
    },
  });

  const handleStartNewConversation = () => {
    confirmAction(
      '确认开启新对话',
      '这将清空当前小奇（全局）的所有历史对话记录，确定要继续吗？',
      () => clearMutation.mutate()
    );
  };

  const handleQuickAskToday = async () => {
    if (!sessionId) return;
    await sendMutation.mutateAsync('我今天要做什么？');
  };

  useEffect(() => {
    if (sessionError) {
      message.error('获取全局会话失败，请稍后重试');
    }
  }, [sessionError, message]);

  if (loadingSession) {
    return (
      <div style={{ textAlign: 'center', paddingTop: 100 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!session) {
    return (
      <div style={{ padding: 24 }}>
        <Text type="danger">无法初始化全局会话</Text>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto' }}>
      <PageHeader
        title="小奇（全局）"
        extra={
          <Space>
            <Button
              icon={<PlusOutlined />}
              onClick={handleStartNewConversation}
              disabled={clearMutation.isPending || !sessionId}
            >
              {!isMobile && '开启新对话'}
            </Button>
          </Space>
        }
      />

      {/* Welcome card */}
      <Card
        style={{
          marginBottom: isMobile ? 16 : 24,
          background: 'linear-gradient(135deg, #5E6AD2 0%, #722ed1 100%)',
          border: 'none',
          color: 'white'
        }}
      >
        <Space direction="vertical" size={isMobile ? 'small' : 'small'} style={{ width: '100%' }}>
          <Space>
            <BulbOutlined style={{ fontSize: isMobile ? 20 : 24, color: '#faad14' }} />
            <Title level={4} style={{ margin: 0, color: 'white', fontSize: isMobile ? 18 : 20 }}>
              智能全局助手
            </Title>
          </Space>
          <Paragraph style={{ color: 'rgba(255, 255, 255, 0.9)', marginBottom: 0, fontSize: isMobile ? 14 : 15 }}>
            这里可以问一些跨任务的问题，例如：
          </Paragraph>
          <Space wrap>
            <Text style={{ color: 'rgba(255, 255, 255, 0.8)', fontSize: isMobile ? 13 : 14 }}>「我今天要做什么？」</Text>
            <Text style={{ color: 'rgba(255, 255, 255, 0.8)', fontSize: isMobile ? 13 : 14 }}>「帮我看看这周安排」</Text>
            <Text style={{ color: 'rgba(255, 255, 255, 0.8)', fontSize: isMobile ? 13 : 14 }}>「有没有已经过期的任务？」</Text>
          </Space>
        </Space>
      </Card>

      {/* Quick actions */}
      <Space style={{ marginBottom: isMobile ? 16 : 24 }} wrap>
        <Button
          icon={<MessageOutlined />}
          onClick={handleQuickAskToday}
          disabled={sendMutation.isPending || !sessionId}
          size={isMobile ? 'middle' : 'large'}
        >
          我今天要做什么？
        </Button>
      </Space>

      {/* Chat area */}
      <Card
        styles={{ body: { padding: 0 } }}
        style={{ borderRadius: isMobile ? 8 : 12, overflow: 'hidden' }}
      >
        <ChatWindow
          messages={messages}
          onSend={(content) => sendMutation.mutateAsync(content)}
          sending={sendMutation.isPending}
          height={isMobile ? 400 : 500}
        />
      </Card>
    </div>
  );
};

export default GlobalAssistant;
