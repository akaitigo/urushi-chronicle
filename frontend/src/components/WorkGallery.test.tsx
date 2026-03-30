import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { WorkGallery } from "./WorkGallery";
import type { Work } from "../types/domain";

const mockWorks: Work[] = [
  {
    id: "work-1",
    title: "金蒔絵硯箱",
    description: "蒔絵技法を用いた硯箱の制作",
    technique: "makie",
    material: "檜木",
    status: "in_progress",
    started_at: "2026-01-15T00:00:00Z",
    created_at: "2026-01-15T00:00:00Z",
    updated_at: "2026-03-01T00:00:00Z",
  },
  {
    id: "work-2",
    title: "螺鈿花器",
    technique: "raden",
    status: "completed",
    started_at: "2025-06-01T00:00:00Z",
    completed_at: "2025-12-15T00:00:00Z",
    created_at: "2025-06-01T00:00:00Z",
    updated_at: "2025-12-15T00:00:00Z",
  },
];

describe("WorkGallery", () => {
  it("renders empty state when no works", () => {
    render(<WorkGallery works={[]} onSelect={vi.fn()} />);
    expect(screen.getByTestId("gallery-empty")).toBeDefined();
    expect(screen.getByText("作品がまだ登録されていません。")).toBeDefined();
  });

  it("renders work cards", () => {
    render(<WorkGallery works={mockWorks} onSelect={vi.fn()} />);
    expect(screen.getByTestId("work-gallery")).toBeDefined();
    expect(screen.getByText("金蒔絵硯箱")).toBeDefined();
    expect(screen.getByText("螺鈿花器")).toBeDefined();
  });

  it("displays technique labels in Japanese", () => {
    render(<WorkGallery works={mockWorks} onSelect={vi.fn()} />);
    expect(screen.getByText("蒔絵")).toBeDefined();
    expect(screen.getByText("螺鈿")).toBeDefined();
  });

  it("displays work status in Japanese", () => {
    render(<WorkGallery works={mockWorks} onSelect={vi.fn()} />);
    expect(screen.getByText("制作中")).toBeDefined();
    expect(screen.getByText("完成")).toBeDefined();
  });

  it("displays work description", () => {
    render(<WorkGallery works={mockWorks} onSelect={vi.fn()} />);
    expect(screen.getByText("蒔絵技法を用いた硯箱の制作")).toBeDefined();
  });

  it("calls onSelect when a card is clicked", () => {
    const onSelect = vi.fn();
    render(<WorkGallery works={mockWorks} onSelect={onSelect} />);
    fireEvent.click(screen.getByTestId("work-card-work-1"));
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect).toHaveBeenCalledWith(mockWorks[0]);
  });

  it("renders the gallery title", () => {
    render(<WorkGallery works={mockWorks} onSelect={vi.fn()} />);
    expect(screen.getByText("作品一覧")).toBeDefined();
  });
});
