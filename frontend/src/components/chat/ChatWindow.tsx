import React, { useState, useRef, useEffect } from 'react';
import { Input, Button } from 'antd';
import { SendOutlined } from '@ant-design/icons';
import { MessageBubble } from '@/components/ui';
import { designTokens } from '@/theme';
import { useResponsive } from '@/hooks';
import type { Message } from '@/types';

interface ChatWindowProps {
  messages: Message[];
  onSend: (content: string) => void;
  sending?: boolean;
  height?: number | string;
}

/**
 * Chat window component with markdown support and animations
 */
export const ChatWindow: React.FC<ChatWindowProps> = ({
  messages,
  onSend,
  sending = false,
  height = 500,
}) => {
  const [inputValue, setInputValue] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);
  const { isMobile, isTablet } = useResponsive();

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

  // Responsive sizing
  const messageAreaPadding = isMobile ? designTokens.spacing.md : designTokens.spacing.lg;
  const inputButtonWidth = isMobile ? '60px' : isTablet ? '70px' : '80px';
  const buttonSize = isMobile ? 'middle' : 'large';
  const placeholderSize = isMobile ? 14 : 15;
  const emojiSize = isMobile ? '36px' : '48px';
  const buttonText = isMobile ? '' : '发送';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height }}>
      {/* Messages area */}
      <div
        ref={scrollRef}
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: messageAreaPadding,
          backgroundColor: designTokens.colors.bg.secondary,
          borderRadius: designTokens.borderRadius.lg,
          marginBottom: isMobile ? designTokens.spacing.sm : designTokens.spacing.md,
          scrollBehavior: 'smooth',
        }}
      >
        {messages.length === 0 ? (
          <div
            style={{
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              color: designTokens.colors.text.secondary,
              gap: designTokens.spacing.sm,
            }}
          >
            <div style={{
              fontSize: emojiSize,
              opacity: 0.5,
            }}>
              💬
            </div>
            <p style={{
              margin: 0,
              fontSize: `var(--font-size-${isMobile ? 'sm' : 'base'})`,
            }}>
              开始与小奇对话...
            </p>
            <p style={{
              margin: 0,
              fontSize: `var(--font-size-xs)`,
              color: designTokens.colors.text.tertiary,
            }}>
              输入消息，按回车发送
            </p>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: designTokens.spacing.sm }}>
            {messages.map((message, index) => (
              <MessageBubble
                key={message.id}
                message={message}
                showAgent={true}
                index={index}
              />
            ))}
          </div>
        )}
      </div>

      {/* Input area */}
      <div style={{
        display: 'flex',
        gap: isMobile ? designTokens.spacing.xs : designTokens.spacing.sm,
        alignItems: 'stretch',
      }}>
        <Input
          placeholder="输入消息..."
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onPressEnter={handleSend}
          disabled={sending}
          size={buttonSize as any}
          style={{
            flex: 1,
            borderRadius: designTokens.borderRadius.md,
            border: `1px solid ${designTokens.colors.gray[200]}`,
            boxShadow: designTokens.shadow.sm,
            fontSize: `${placeholderSize}px`,
          }}
        />
        <Button
          type="primary"
          icon={<SendOutlined />}
          onClick={handleSend}
          loading={sending}
          size={buttonSize as any}
          style={{
            borderRadius: designTokens.borderRadius.md,
            boxShadow: designTokens.shadow.sm,
            minWidth: inputButtonWidth,
          }}
        >
          {buttonText}
        </Button>
      </div>
    </div>
  );
};
