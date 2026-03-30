import { describe, expect, it } from "vitest";
import {
  formatDate,
  formatDateTime,
  stepCategoryLabels,
  techniqueLabels,
  workStatusLabels,
} from "./labels";

describe("labels", () => {
  describe("techniqueLabels", () => {
    it("maps makie to 蒔絵", () => {
      expect(techniqueLabels.makie).toBe("蒔絵");
    });

    it("maps raden to 螺鈿", () => {
      expect(techniqueLabels.raden).toBe("螺鈿");
    });

    it("maps makie_raden to 蒔絵螺鈿", () => {
      expect(techniqueLabels.makie_raden).toBe("蒔絵螺鈿");
    });

    it("maps other to その他", () => {
      expect(techniqueLabels.other).toBe("その他");
    });
  });

  describe("workStatusLabels", () => {
    it("maps in_progress to 制作中", () => {
      expect(workStatusLabels.in_progress).toBe("制作中");
    });

    it("maps completed to 完成", () => {
      expect(workStatusLabels.completed).toBe("完成");
    });

    it("maps archived to アーカイブ", () => {
      expect(workStatusLabels.archived).toBe("アーカイブ");
    });
  });

  describe("stepCategoryLabels", () => {
    it("maps all step categories correctly", () => {
      expect(stepCategoryLabels.shitanuri).toBe("下塗り");
      expect(stepCategoryLabels.nakanuri).toBe("中塗り");
      expect(stepCategoryLabels.uwanuri).toBe("上塗り");
      expect(stepCategoryLabels.makie).toBe("蒔絵");
      expect(stepCategoryLabels.raden).toBe("螺鈿");
      expect(stepCategoryLabels.togidashi).toBe("研ぎ出し");
      expect(stepCategoryLabels.roiro).toBe("呂色仕上げ");
      expect(stepCategoryLabels.other).toBe("その他");
    });
  });

  describe("formatDate", () => {
    it("formats an ISO date string", () => {
      const result = formatDate("2026-03-29T10:00:00Z");
      expect(result).toContain("2026");
      expect(result).toContain("29");
    });
  });

  describe("formatDateTime", () => {
    it("formats an ISO date string with time", () => {
      const result = formatDateTime("2026-03-29T10:30:00Z");
      expect(result).toContain("2026");
    });
  });
});
