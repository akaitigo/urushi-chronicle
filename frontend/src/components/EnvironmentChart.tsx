import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { EnvironmentReading } from "../types/domain";

interface EnvironmentChartProps {
  readings: EnvironmentReading[];
  title?: string;
}

interface ChartDataPoint {
  time: string;
  displayTime: string;
  temperature: number;
  humidity: number;
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString("ja-JP", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function EnvironmentChart({
  readings,
  title = "温湿度グラフ",
}: EnvironmentChartProps) {
  if (readings.length === 0) {
    return (
      <div className="chart-empty" data-testid="chart-empty">
        <p>環境データがありません。</p>
      </div>
    );
  }

  const data: ChartDataPoint[] = [...readings]
    .sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime())
    .map((r) => ({
      time: r.time,
      displayTime: formatTime(r.time),
      temperature: r.temperature,
      humidity: r.humidity,
    }));

  return (
    <div className="chart-container" data-testid="environment-chart">
      <h3 className="chart-title">{title}</h3>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart
          data={data}
          margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="displayTime" fontSize={12} />
          <YAxis
            yAxisId="temp"
            domain={[0, 50]}
            label={{
              value: "温度 (°C)",
              angle: -90,
              position: "insideLeft",
              style: { fontSize: 12 },
            }}
          />
          <YAxis
            yAxisId="humid"
            orientation="right"
            domain={[0, 100]}
            label={{
              value: "湿度 (%)",
              angle: 90,
              position: "insideRight",
              style: { fontSize: 12 },
            }}
          />
          <Tooltip
            labelFormatter={(label) => `時刻: ${String(label)}`}
            formatter={(value, name) => [
              `${Number(value).toFixed(1)}${name === "temperature" ? "°C" : "%"}`,
              name === "temperature" ? "温度" : "湿度",
            ]}
          />
          <Legend
            formatter={(value: string) =>
              value === "temperature" ? "温度 (°C)" : "湿度 (%)"
            }
          />
          <Line
            yAxisId="temp"
            type="monotone"
            dataKey="temperature"
            stroke="#e74c3c"
            strokeWidth={2}
            dot={false}
            name="temperature"
          />
          <Line
            yAxisId="humid"
            type="monotone"
            dataKey="humidity"
            stroke="#3498db"
            strokeWidth={2}
            dot={false}
            name="humidity"
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
