import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { WorkForm } from "./WorkForm";

describe("WorkForm", () => {
  const defaultProps = {
    onSubmit: vi.fn(),
    onCancel: vi.fn(),
    submitting: false,
    error: null,
  };

  it("renders the form with all fields", () => {
    render(<WorkForm {...defaultProps} />);
    expect(screen.getByTestId("work-form")).toBeDefined();
    expect(screen.getByLabelText(/タイトル/)).toBeDefined();
    expect(screen.getByLabelText("説明")).toBeDefined();
    expect(screen.getByLabelText(/技法/)).toBeDefined();
    expect(screen.getByLabelText("素材")).toBeDefined();
  });

  it("renders the form title", () => {
    render(<WorkForm {...defaultProps} />);
    expect(screen.getByText("作品を登録")).toBeDefined();
  });

  it("disables submit when title is empty", () => {
    render(<WorkForm {...defaultProps} />);
    const btn = screen.getByTestId("submit-button");
    expect((btn as HTMLButtonElement).disabled).toBe(true);
  });

  it("enables submit when title is filled", () => {
    render(<WorkForm {...defaultProps} />);
    fireEvent.change(screen.getByLabelText(/タイトル/), {
      target: { value: "テスト作品" },
    });
    const btn = screen.getByTestId("submit-button");
    expect((btn as HTMLButtonElement).disabled).toBe(false);
  });

  it("calls onSubmit with form data", () => {
    const onSubmit = vi.fn();
    render(<WorkForm {...defaultProps} onSubmit={onSubmit} />);

    fireEvent.change(screen.getByLabelText(/タイトル/), {
      target: { value: "蒔絵香合" },
    });
    fireEvent.change(screen.getByLabelText("説明"), {
      target: { value: "テスト説明" },
    });
    fireEvent.change(screen.getByLabelText(/技法/), {
      target: { value: "raden" },
    });
    fireEvent.change(screen.getByLabelText("素材"), {
      target: { value: "欅" },
    });

    fireEvent.submit(screen.getByTestId("work-form"));

    expect(onSubmit).toHaveBeenCalledWith({
      title: "蒔絵香合",
      description: "テスト説明",
      technique: "raden",
      material: "欅",
    });
  });

  it("calls onCancel when cancel is clicked", () => {
    const onCancel = vi.fn();
    render(<WorkForm {...defaultProps} onCancel={onCancel} />);
    fireEvent.click(screen.getByText("キャンセル"));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it("shows submitting state", () => {
    render(<WorkForm {...defaultProps} submitting={true} />);
    expect(screen.getByText("登録中...")).toBeDefined();
    expect(
      (screen.getByTestId("submit-button") as HTMLButtonElement).disabled,
    ).toBe(true);
  });

  it("displays error message", () => {
    render(<WorkForm {...defaultProps} error="タイトルは必須です" />);
    expect(screen.getByTestId("form-error")).toBeDefined();
    expect(screen.getByText("タイトルは必須です")).toBeDefined();
  });

  it("includes technique options", () => {
    render(<WorkForm {...defaultProps} />);
    expect(screen.getByText("蒔絵")).toBeDefined();
    expect(screen.getByText("螺鈿")).toBeDefined();
    expect(screen.getByText("蒔絵螺鈿")).toBeDefined();
    expect(screen.getByText("その他")).toBeDefined();
  });
});
