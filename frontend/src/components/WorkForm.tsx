import { type FormEvent, useState } from "react";
import { techniqueLabels } from "../lib/labels";
import type { Technique } from "../types/domain";

export interface WorkFormData {
  title: string;
  description: string;
  technique: Technique;
  material: string;
}

interface WorkFormProps {
  onSubmit: (data: WorkFormData) => void;
  onCancel: () => void;
  submitting: boolean;
  error: string | null;
}

const techniques: Technique[] = ["makie", "raden", "makie_raden", "other"];

export function WorkForm({
  onSubmit,
  onCancel,
  submitting,
  error,
}: WorkFormProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [technique, setTechnique] = useState<Technique>("makie");
  const [material, setMaterial] = useState("");

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    onSubmit({ title, description, technique, material });
  };

  return (
    <form className="work-form" data-testid="work-form" onSubmit={handleSubmit}>
      <h2 className="form-title">作品を登録</h2>

      {error && (
        <div className="form-error" data-testid="form-error">
          {error}
        </div>
      )}

      <div className="form-field">
        <label htmlFor="work-title">
          タイトル <span className="required">*</span>
        </label>
        <input
          id="work-title"
          type="text"
          required
          maxLength={200}
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="例: 蒔絵香合 — 秋草"
        />
      </div>

      <div className="form-field">
        <label htmlFor="work-description">説明</label>
        <textarea
          id="work-description"
          rows={3}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="作品の概要や制作意図"
        />
      </div>

      <div className="form-field">
        <label htmlFor="work-technique">
          技法 <span className="required">*</span>
        </label>
        <select
          id="work-technique"
          value={technique}
          onChange={(e) => setTechnique(e.target.value as Technique)}
        >
          {techniques.map((t) => (
            <option key={t} value={t}>
              {techniqueLabels[t]}
            </option>
          ))}
        </select>
      </div>

      <div className="form-field">
        <label htmlFor="work-material">素材</label>
        <input
          id="work-material"
          type="text"
          value={material}
          onChange={(e) => setMaterial(e.target.value)}
          placeholder="例: 欅、檜"
        />
      </div>

      <div className="form-actions">
        <button
          type="submit"
          className="btn-primary"
          disabled={submitting || title.trim() === ""}
          data-testid="submit-button"
        >
          {submitting ? "登録中..." : "登録"}
        </button>
        <button
          type="button"
          className="btn-secondary"
          onClick={onCancel}
          disabled={submitting}
        >
          キャンセル
        </button>
      </div>
    </form>
  );
}
