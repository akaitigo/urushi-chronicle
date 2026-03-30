import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { EnvironmentChart } from "./EnvironmentChart";
import type { EnvironmentReading } from "../types/domain";

// recharts uses ResizeObserver internally; provide a stub in jsdom.
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
globalThis.ResizeObserver =
  ResizeObserverStub as unknown as typeof ResizeObserver;

const mockReadings: EnvironmentReading[] = [
  {
    time: "2026-03-29T10:00:00Z",
    sensor_id: "urushi-buro-1",
    location: "漆風呂A",
    temperature: 22.5,
    humidity: 75.0,
  },
  {
    time: "2026-03-29T10:15:00Z",
    sensor_id: "urushi-buro-1",
    location: "漆風呂A",
    temperature: 23.0,
    humidity: 74.5,
  },
  {
    time: "2026-03-29T10:30:00Z",
    sensor_id: "urushi-buro-1",
    location: "漆風呂A",
    temperature: 22.8,
    humidity: 76.0,
  },
];

describe("EnvironmentChart", () => {
  it("renders empty state when no readings", () => {
    render(<EnvironmentChart readings={[]} />);
    expect(screen.getByTestId("chart-empty")).toBeDefined();
    expect(screen.getByText("環境データがありません。")).toBeDefined();
  });

  it("renders chart container with data", () => {
    render(<EnvironmentChart readings={mockReadings} />);
    expect(screen.getByTestId("environment-chart")).toBeDefined();
  });

  it("renders the default title", () => {
    render(<EnvironmentChart readings={mockReadings} />);
    expect(screen.getByText("温湿度グラフ")).toBeDefined();
  });

  it("renders a custom title", () => {
    render(<EnvironmentChart readings={mockReadings} title="漆風呂A 温湿度" />);
    expect(screen.getByText("漆風呂A 温湿度")).toBeDefined();
  });
});
