import styles from './ActivityChart.module.css';

export type ActivityDayState = 'past' | 'today' | 'future';

interface ActivityChartItem {
  label: string;
  value: number;
  max: number;
  state?: ActivityDayState;
  /** Заголовок при наведении (например, дата). */
  title?: string;
}

interface ActivityChartProps {
  data: ActivityChartItem[];
}

export default function ActivityChart({ data }: ActivityChartProps) {
  const maxVal = Math.max(...data.map((d) => d.value), 1);

  return (
    <div className={styles.chart}>
      {data.map((item, idx) => {
        const state: ActivityDayState = item.state ?? 'past';
        const trackClass = [
          styles.barTrack,
          state === 'today' ? styles.barTrackToday : '',
          state === 'future' ? styles.barTrackFuture : '',
        ].join(' ').trim();
        const fillClass = [
          styles.barFill,
          state === 'today' ? styles.barFillToday : '',
          state === 'future' ? styles.barFillFuture : '',
        ].join(' ').trim();
        const labelClass = [
          styles.barLabel,
          state === 'today' ? styles.barLabelToday : '',
          state === 'future' ? styles.barLabelFuture : '',
        ].join(' ').trim();

        const heightPct = state === 'future'
          ? 0
          : (item.value / maxVal) * 100;

        return (
          <div key={`${item.label}-${idx}`} className={styles.bar} title={item.title ?? item.label}>
            <span className={styles.barValue}>{item.value > 0 ? item.value : ''}</span>
            <div className={trackClass}>
              <div className={fillClass} style={{ height: `${heightPct}%` }} />
            </div>
            <span className={labelClass}>{item.label}</span>
          </div>
        );
      })}
    </div>
  );
}
