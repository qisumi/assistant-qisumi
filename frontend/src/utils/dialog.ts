import { Modal } from 'antd';

/**
 * 确认删除对话框
 */
export const confirmDelete = (
  itemName: string,
  itemTitle: string,
  onConfirm: () => void
) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除${itemName}「${itemTitle}」吗？此操作不可恢复。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: onConfirm,
  });
};

/**
 * 通用确认对话框
 */
export const confirmAction = (
  title: string,
  content: string,
  onConfirm: () => void,
  okText = '确定',
  cancelText = '取消'
) => {
  Modal.confirm({
    title,
    content,
    okText,
    cancelText,
    onOk: onConfirm,
  });
};
