import React, { useState, useRef, useEffect } from 'react';
import { List, Input, Button, Avatar, Space, Typography, Tag } from 'antd';
import { 
  UserOutlined, RobotOutlined, CheckCircleOutlined, CalendarOutlined, 
  FileTextOutlined, BulbOutlined, SettingOutlined 
} from '@ant-design/icons';
import type { Message } from '@/types';
import { AGENT_LABELS, AGENT_COLORS } from '@/constants';

const { Text } = Typography;

interface ChatWindowProps {
  messages: Message[];
  onSend: (content: string) => void;
  sending?: boolean;
  height?: number | string;
}

const AGENT_ICONS: Record<string, React.ReactNode> = {
  executor: <CheckCircleOutlined />,
  planner: <CalendarOutlined />,
  summarizer: <FileTextOutlined />,
  global: <BulbOutlined />,
};

const ASSISTANT_NAME = '小奇';

export const ChatWindow: React.FC<ChatWindowProps> = ({
  messages,
  onSend,
  sending = false,
  height = 500,
}) => {

  const [inputValue, setInputValue] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  const handleSend = () => {
    if (!inputValue.trim() || sending) return;
    onSend(inputValue);
    setInputValue('');
  };

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height }}>
      <div
        ref={scrollRef}
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '16px',
          backgroundColor: '#f5f5f5',
          borderRadius: '8px',
          marginBottom: '16px',
        }}
      >
        <List
          dataSource={messages}
          renderItem={(item) => {
            const isUser = item.role === 'user';
            const isSystem = item.role === 'system';
            const variantLabel = item.agentName ? AGENT_LABELS[item.agentName] : undefined;
            const senderLabel = isUser ? '我' : isSystem ? '系统' : ASSISTANT_NAME;
            const trimmedContent = item.content?.trim() ?? '';
            const isEmptyContent = trimmedContent.length === 0;
            const emptyHint = isUser ? '（空消息）' : `${senderLabel}${variantLabel ? `（${variantLabel}）` : ''}暂未返回内容`;

            const avatarIcon = isUser
              ? <UserOutlined />
              : isSystem
                ? <SettingOutlined />
                : (item.agentName ? (AGENT_ICONS[item.agentName] || <RobotOutlined />) : <RobotOutlined />);

            const avatarBg = isUser
              ? '#1677ff'
              : isSystem
                ? '#8c8c8c'
                : (item.agentName ? (AGENT_COLORS[item.agentName] || '#13c2c2') : '#13c2c2');

            return (
              <List.Item style={{ border: 'none', padding: '8px 0' }}>
                <div
                  style={{
                    display: 'flex',
                    flexDirection: isUser ? 'row-reverse' : 'row',
                    width: '100%',
                    alignItems: 'flex-start',
                  }}
                >
                  <Avatar
                    icon={avatarIcon}
                    style={{
                      backgroundColor: avatarBg,
                      flexShrink: 0,
                    }}
                  />
                  <div
                    style={{
                      margin: isUser ? '0 12px 0 0' : '0 0 0 12px',
                      maxWidth: '70%',
                    }}
                  >
                    <div
                      style={{
                        textAlign: isUser ? 'right' : 'left',
                        marginBottom: '4px',
                      }}
                    >
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        {senderLabel}
                      </Text>
                      {!isUser && !isSystem && variantLabel && (
                        <Tag style={{ marginLeft: '8px' }}>{variantLabel}</Tag>
                      )}
                    </div>
                    <div
                      style={{
                        backgroundColor: isEmptyContent ? '#fffbe6' : isUser ? '#1677ff' : '#fff',
                        color: isEmptyContent ? '#8c8c8c' : isUser ? '#fff' : 'inherit',
                        border: isEmptyContent ? '1px dashed #d9d9d9' : 'none',
                        padding: '8px 12px',
                        borderRadius: '8px',
                        boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-word',
                      }}
                    >
                      {isEmptyContent ? (
                        <Text type="secondary" style={{ fontStyle: 'italic' }}>
                          {emptyHint}
                        </Text>
                      ) : (
                        item.content
                      )}
                    </div>
                  </div>
                </div>
              </List.Item>
            );
          }}
        />
      </div>
      <Space.Compact style={{ width: '100%' }}>
        <Input
          placeholder="输入消息..."
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onPressEnter={handleSend}
          disabled={sending}
        />
        <Button type="primary" onClick={handleSend} loading={sending}>
          发送
        </Button>
      </Space.Compact>
    </div>
  );
};
