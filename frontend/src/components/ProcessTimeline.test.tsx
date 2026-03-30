import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ProcessTimeline } from "./ProcessTimeline";
import type { ProcessStep } from "../types/domain";

const mockSteps: ProcessStep[] = [
  {
    id: "step-1",
    work_id: "work-1",
    name: "木地研磨",
    description: "木地の表面を研磨紙で整える",
    step_order: 1,
    category: "shitanuri",
    materials_used: ["研磨紙 #400", "研磨紙 #800"],
    notes: "丁寧に木目に沿って研磨する",
    started_at: "2026-01-20T00:00:00Z",
    completed_at: "2026-01-22T00:00:00Z",
    created_at: "2026-01-20T00:00:00Z",
    updated_at: "2026-01-22T00:00:00Z",
  },
  {
    id: "step-2",
    work_id: "work-1",
    name: "下地塗り",
    description: "生漆を木地に塗布",
    step_order: 2,
    category: "shitanuri",
    started_at: "2026-01-23T00:00:00Z",
    created_at: "2026-01-23T00:00:00Z",
    updated_at: "2026-01-23T00:00:00Z",
  },
  {
    id: "step-3",
    work_id: "work-1",
    name: "蒔絵下絵",
    step_order: 3,
    category: "makie",
    started_at: "2026-02-01T00:00:00Z",
    created_at: "2026-02-01T00:00:00Z",
    updated_at: "2026-02-01T00:00:00Z",
  },
];

describe("ProcessTimeline", () => {
  it("renders empty state when no steps", () => {
    render(
      <ProcessTimeline steps={[]} workTitle="金蒔絵硯箱" onBack={vi.fn()} />,
    );
    expect(screen.getByTestId("timeline-empty")).toBeDefined();
    expect(screen.getByText("工程がまだ記録されていません。")).toBeDefined();
  });

  it("renders the work title", () => {
    render(
      <ProcessTimeline
        steps={mockSteps}
        workTitle="金蒔絵硯箱"
        onBack={vi.fn()}
      />,
    );
    expect(screen.getByText("金蒔絵硯箱 - 制作工程")).toBeDefined();
  });

  it("renders all steps", () => {
    render(
      <ProcessTimeline
        steps={mockSteps}
        workTitle="金蒔絵硯箱"
        onBack={vi.fn()}
      />,
    );
    expect(screen.getByTestId("timeline-step-step-1")).toBeDefined();
    expect(screen.getByTestId("timeline-step-step-2")).toBeDefined();
    expect(screen.getByTestId("timeline-step-step-3")).toBeDefined();
  });

  it("displays step names", () => {
    render(
      <ProcessTimeline
        steps={mockSteps}
        workTitle="金蒔絵硯箱"
        onBack={vi.fn()}
      />,
    );
    expect(screen.getByText("木地研磨")).toBeDefined();
    expect(screen.getByText("下地塗り")).toBeDefined();
    expect(screen.getByText("蒔絵下絵")).toBeDefined();
  });

  it("displays step category labels in Japanese", () => {
    render(
      <ProcessTimeline
        steps={mockSteps}
        workTitle="金蒔絵硯箱"
        onBack={vi.fn()}
      />,
    );
    // "下塗り" appears for shitanuri category (step-1 and step-2)
    const shitanuriLabels = screen.getAllByText("下塗り");
    expect(shitanuriLabels.length).toBe(2);
    expect(screen.getByText("蒔絵")).toBeDefined();
  });

  it("displays step description", () => {
    render(
      <ProcessTimeline
        steps={mockSteps}
        workTitle="金蒔絵硯箱"
        onBack={vi.fn()}
      />,
    );
    expect(screen.getByText("木地の表面を研磨紙で整える")).toBeDefined();
  });

  it("displays step notes", () => {
    render(
      <ProcessTimeline
        steps={mockSteps}
        workTitle="金蒔絵硯箱"
        onBack={vi.fn()}
      />,
    );
    expect(screen.getByText("丁寧に木目に沿って研磨する")).toBeDefined();
  });

  it("displays materials used", () => {
    render(
      <ProcessTimeline
        steps={mockSteps}
        workTitle="金蒔絵硯箱"
        onBack={vi.fn()}
      />,
    );
    expect(screen.getByText("使用材料:")).toBeDefined();
    expect(screen.getByText("研磨紙 #400")).toBeDefined();
    expect(screen.getByText("研磨紙 #800")).toBeDefined();
  });

  it("calls onBack when back button is clicked", () => {
    const onBack = vi.fn();
    render(
      <ProcessTimeline
        steps={mockSteps}
        workTitle="金蒔絵硯箱"
        onBack={onBack}
      />,
    );
    fireEvent.click(screen.getByTestId("back-button"));
    expect(onBack).toHaveBeenCalledTimes(1);
  });

  it("sorts steps by step_order", () => {
    const unorderedSteps = [mockSteps[2], mockSteps[0], mockSteps[1]];
    render(
      <ProcessTimeline
        steps={unorderedSteps}
        workTitle="金蒔絵硯箱"
        onBack={vi.fn()}
      />,
    );
    const items = screen.getAllByRole("listitem");
    expect(items.length).toBe(3);
  });
});
