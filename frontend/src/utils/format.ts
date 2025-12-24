import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/zh-cn';

dayjs.extend(relativeTime);
dayjs.locale('zh-cn');

/**
 * 格式化日期时间为完整格式
 * @param date 日期字符串或null
 * @returns 格式化后的日期字符串，如 "2024-12-24 15:30"
 */
export function formatDateTime(date: string | null | undefined): string {
  if (!date) return '-';
  return dayjs(date).format('YYYY-MM-DD HH:mm');
}

/**
 * 格式化日期为短格式
 * @param date 日期字符串或null
 * @returns 格式化后的日期字符串，如 "2024-12-24"
 */
export function formatDate(date: string | null | undefined): string {
  if (!date) return '-';
  return dayjs(date).format('YYYY-MM-DD');
}

/**
 * 格式化相对时间
 * @param date 日期字符串或null
 * @returns 相对时间字符串，如 "2小时前"、"3天前"
 */
export function formatRelativeTime(date: string | null | undefined): string {
  if (!date) return '-';
  return dayjs(date).fromNow();
}

/**
 * 格式化时间范围
 * @param start 开始时间
 * @param end 结束时间
 * @returns 格式化后的时间范围，如 "2024-12-24 09:00 - 2024-12-24 18:00"
 */
export function formatTimeRange(start: string | null | undefined, end: string | null | undefined): string {
  if (!start && !end) return '-';
  if (!start) return `至 ${formatDateTime(end)}`;
  if (!end) return `${formatDateTime(start)} 起`;
  return `${formatDateTime(start)} - ${formatDateTime(end)}`;
}

/**
 * 检查日期是否过期
 * @param date 日期字符串或null
 * @returns 是否过期
 */
export function isOverdue(date: string | null | undefined): boolean {
  if (!date) return false;
  return dayjs(date).isBefore(dayjs());
}

/**
 * 获取日期状态标签
 * @param date 日期字符串或null
 * @returns 日期状态标签文本
 */
export function getDateStatus(date: string | null | undefined): string {
  if (!date) return '未设置';
  if (isOverdue(date)) return '已过期';
  return '进行中';
}
