import React, { useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, Button, Space, Spin, Typography, message as antdMessage, Modal } from 'antd';
import { BulbOutlined, PlusOutlined } from '@ant-design/icons';

import { ChatWindow } from '@/components/chat/ChatWindow';
import { getOrCreateGlobalSession, fetchSessionMessages, sendSessionMessage, clearSessionMessages } from '@/api/sessions';
import type { Message } from '@/types';

const { Title, Text, Paragraph } = Typography;

const GlobalAssistant: React.FC = () => {
  const queryClient = useQueryClient();

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
      antdMessage.error('发送失败，请稍后重试');
    },
  });

  // 4. 清空消息（开启新对话）
  const clearMutation = useMutation({
    mutationFn: () => clearSessionMessages(sessionId!),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['sessionMessages', sessionId] });
      antdMessage.success('已开启新对话');
    },
    onError: (err: any) => {
      console.error(err);
      antdMessage.error('开启新对话失败，请稍后重试');
    },
  });

  const handleStartNewConversation = () => {
    Modal.confirm({
      title: '确认开启新对话',
      content: '这将清空当前小奇（全局）的所有历史对话记录，确定要继续吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        clearMutation.mutate();
      },
    });
  };

  const handleQuickAskToday = async () => {
    if (!sessionId) return;
    await sendMutation.mutateAsync('我今天要做什么？');
  };

  useEffect(() => {
    if (sessionError) {
      antdMessage.error('获取全局会话失败，请稍后重试');
    }
  }, [sessionError]);

  if (loadingSession) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Spin />
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
    <div style={{ padding: 24, display: 'flex', justifyContent: 'center' }}>
      <Card style={{ maxWidth: 1000, width: '100%' }}>
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div>
            <Space align="center">
              <BulbOutlined style={{ fontSize: 24, color: '#faad14' }} />
              <Title level={3} style={{ margin: 0 }}>
                小奇（全局）
              </Title>
            </Space>
            <Paragraph type="secondary" style={{ marginTop: 8 }}>
              这里可以问一些跨任务的问题，例如：
              <br />
              「我今天要做什么？」、「帮我看看这周安排」、「有没有已经过期的任务？」等。
            </Paragraph>
          </div>

          <Space style={{ justifyContent: 'space-between', width: '100%' }}>
            <Space>
              <Button
                icon={<BulbOutlined />}
                onClick={handleQuickAskToday}
                disabled={sendMutation.isPending || !sessionId}
              >
                我今天要做什么？
              </Button>
            </Space>
            <Space>
              <Button
                icon={<PlusOutlined />}
                onClick={handleStartNewConversation}
                disabled={clearMutation.isPending || !sessionId}
              >
                开启新对话
              </Button>
            </Space>
          </Space>

          <ChatWindow
            messages={messages}
            onSend={(content) => sendMutation.mutateAsync(content)}
            sending={sendMutation.isPending}
            height={520}
          />
        </Space>
      </Card>
    </div>
  );
};

export default GlobalAssistant;
