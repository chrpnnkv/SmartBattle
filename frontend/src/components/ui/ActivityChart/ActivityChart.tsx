import styles from './ActivityChart.module.css';

interface ActivityChartProps {
  data: { label: string; value: number; max: number }[];
}

export default function ActivityChart({ data }: ActivityChartProps) {
  const maxVal = Math.max(...data.map((d) => d.value), 1);

  return (
    <div className={styles.chart}>
      {data.map((item) => (
        <div key={item.label} className={styles.bar}>
          <div className={styles.barTrack}>
            <div
              className={styles.barFill}
              style={{ height: `${(item.value / maxVal) * 100}%` }}
            />
          </div>
          <span className={styles.barLabel}>{item.label}</span>
        </div>
      ))}
    </div>
  );
}
