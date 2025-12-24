import React from 'react';
import { Typography } from 'antd';
import { motion } from 'framer-motion';
import { Markdown } from './';
import { AGENT_LABELS, AGENT_COLORS } from '@/constants';
import { designTokens } from '@/theme';
import { useResponsive } from '@/hooks';
import type { Message } from '@/types';
import dayjs from 'dayjs';

const { Text } = Typography;

interface MessageBubbleProps {
  message: Message;
  showAgent?: boolean;
  className?: string;
  style?: React.CSSProperties;
  index?: number;
}

/**
 * Chat message bubble with markdown support and animations
 */
export const MessageBubble: React.FC<MessageBubbleProps> = ({
  message,
  showAgent = true,
  className,
  style,
  index = 0
}) => {
  const isUser = message.role === 'user';
  const agentLabel = message.agentName ? AGENT_LABELS[message.agentName] : null;
  const { isMobile } = useResponsive();

  // Responsive sizing
  const bubbleMaxWidth = isMobile ? '85%' : '80%';
  const bubblePadding = isMobile
    ? `${designTokens.spacing.sm}px ${designTokens.spacing.md}px`
    : `${designTokens.spacing.md}px ${designTokens.spacing.lg}px`;
  const labelFontSize = isMobile ? '11px' : '12px';
  const contentFontSize = isMobile ? '14px' : '15px';

  return (
    <motion.div
      initial={{ opacity: 0, y: 10, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{
        delay: index * 0.05,
        type: 'spring',
        stiffness: 500,
        damping: 30,
      }}
      className={className}
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: isUser ? 'flex-end' : 'flex-start',
        marginBottom: isMobile ? designTokens.spacing.sm : designTokens.spacing.md,
        ...style
      }}
    >
      {/* Agent label (for assistant messages) */}
      {!isUser && showAgent && agentLabel && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: index * 0.05 + 0.1 }}
        >
          <Text
            type="secondary"
            style={{
              fontSize: labelFontSize,
              marginBottom: designTokens.spacing.xs,
              marginLeft: designTokens.spacing.md,
              color: message.agentName ? AGENT_COLORS[message.agentName] : designTokens.colors.text.secondary,
              fontWeight: designTokens.typography.fontWeight.medium,
            }}
          >
            {agentLabel}
          </Text>
        </motion.div>
      )}

      {/* Message bubble */}
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{
          delay: index * 0.05 + 0.05,
          type: 'spring',
          stiffness: 400,
          damping: 25,
        }}
        style={{
          maxWidth: bubbleMaxWidth,
          padding: bubblePadding,
          borderRadius: designTokens.borderRadius.lg,
          backgroundColor: isUser ? designTokens.colors.primary : designTokens.colors.bg.base,
          color: isUser ? designTokens.colors.text.inverse : designTokens.colors.text.primary,
          boxShadow: designTokens.shadow.md,
          border: isUser ? 'none' : `1px solid ${designTokens.colors.gray[100]}`,
        }}
      >
        {/* Markdown content for assistant, plain text for user */}
        {isUser ? (
          <Text style={{
            color: designTokens.colors.text.inverse,
            whiteSpace: 'pre-wrap',
            fontSize: contentFontSize,
            lineHeight: designTokens.typography.lineHeight.normal,
          }}>
            {message.content}
          </Text>
        ) : (
          <div style={{
            color: designTokens.colors.text.primary,
            fontSize: contentFontSize,
            lineHeight: designTokens.typography.lineHeight.normal,
          }}>
            <Markdown content={message.content} />
          </div>
        )}
      </motion.div>

      {/* Timestamp */}
      <Text
        type="secondary"
        style={{
          fontSize: labelFontSize,
          marginTop: designTokens.spacing.xs,
          marginLeft: isUser ? '0' : designTokens.spacing.md,
          marginRight: isUser ? designTokens.spacing.md : '0',
          color: designTokens.colors.text.tertiary,
        }}
      >
        {dayjs(message.createdAt).format('HH:mm')}
      </Text>
    </motion.div>
  );
};
