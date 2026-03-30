/** Technique represents the type of lacquerware technique used. */
export type Technique = "makie" | "raden" | "makie_raden" | "other";

/** WorkStatus represents the current status of a work. */
export type WorkStatus = "in_progress" | "completed" | "archived";

/** StepCategory represents the type of production process step. */
export type StepCategory =
  | "shitanuri"
  | "nakanuri"
  | "uwanuri"
  | "makie"
  | "raden"
  | "togidashi"
  | "roiro"
  | "other";

/** ImageType represents the purpose of an image. */
export type ImageType = "process" | "macro" | "aging" | "overview";

/** Work represents a lacquerware piece being created or documented. */
export interface Work {
  id: string;
  title: string;
  description?: string;
  technique: Technique;
  material?: string;
  status: WorkStatus;
  started_at: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

/** ProcessStep represents a single step in the lacquerware production process. */
export interface ProcessStep {
  id: string;
  work_id: string;
  name: string;
  description?: string;
  step_order: number;
  category: StepCategory;
  materials_used?: string[];
  notes?: string;
  started_at: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

/** Image represents a photograph attached to a work or process step. */
export interface WorkImage {
  id: string;
  work_id: string;
  process_step_id?: string;
  file_path: string;
  file_size_bytes: number;
  content_type: "image/jpeg" | "image/png";
  image_type: ImageType;
  caption?: string;
  taken_at?: string;
  created_at: string;
}

/** EnvironmentReading represents a sensor measurement from the urushi-buro. */
export interface EnvironmentReading {
  time: string;
  sensor_id: string;
  location: string;
  temperature: number;
  humidity: number;
  work_id?: string;
  process_step_id?: string;
}
