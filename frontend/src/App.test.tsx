import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { App } from "./App";

describe("App", () => {
  it("renders the application title", () => {
    render(<App />);
    expect(screen.getByText("urushi-chronicle")).toBeDefined();
  });

  it("renders the description", () => {
    render(<App />);
    expect(
      screen.getByText("蒔絵・螺鈿制作工程デジタルアーカイブ"),
    ).toBeDefined();
  });
});
