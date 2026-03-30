import type { StepCategory, Technique, WorkStatus } from "../types/domain";

/** Japanese labels for Technique values. */
export const techniqueLabels: Record<Technique, string> = {
  makie: "蒔絵",
  raden: "螺鈿",
  makie_raden: "蒔絵螺鈿",
  other: "その他",
};

/** Japanese labels for WorkStatus values. */
export const workStatusLabels: Record<WorkStatus, string> = {
  in_progress: "制作中",
  completed: "完成",
  archived: "アーカイブ",
};

/** Japanese labels for StepCategory values. */
export const stepCategoryLabels: Record<StepCategory, string> = {
  shitanuri: "下塗り",
  nakanuri: "中塗り",
  uwanuri: "上塗り",
  makie: "蒔絵",
  raden: "螺鈿",
  togidashi: "研ぎ出し",
  roiro: "呂色仕上げ",
  other: "その他",
};

/** Format an ISO date string into a localized Japanese date. */
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("ja-JP", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

/** Format an ISO date string into a localized Japanese date with time. */
export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("ja-JP", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
