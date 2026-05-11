import type { QuizStatus } from '../../../types';
import styles from './Badge.module.css';

interface BadgeProps {
  status: QuizStatus;
}

const LABELS: Record<QuizStatus, string> = {
  published: 'Опубликован',
  draft: 'Черновик',
};

export default function Badge({ status }: BadgeProps) {
  return (
    <span className={[styles.badge, styles[status]].join(' ')}>
      {LABELS[status]}
    </span>
  );
}
