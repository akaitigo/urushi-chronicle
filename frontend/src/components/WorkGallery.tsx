import type { Work } from "../types/domain";
import { formatDate, techniqueLabels, workStatusLabels } from "../lib/labels";

interface WorkGalleryProps {
  works: Work[];
  onSelect: (work: Work) => void;
  onDelete?: (work: Work) => void;
}

const statusColors: Record<string, string> = {
  in_progress: "#e67e22",
  completed: "#27ae60",
  archived: "#95a5a6",
};

export function WorkGallery({ works, onSelect, onDelete }: WorkGalleryProps) {
  if (works.length === 0) {
    return (
      <div className="gallery-empty" data-testid="gallery-empty">
        <p>作品がまだ登録されていません。</p>
      </div>
    );
  }

  return (
    <div className="gallery" data-testid="work-gallery">
      <h2 className="gallery-title">作品一覧</h2>
      <div className="gallery-grid">
        {works.map((work) => (
          <div key={work.id} className="gallery-card-wrapper">
            <button
              type="button"
              className="gallery-card"
              data-testid={`work-card-${work.id}`}
              onClick={() => onSelect(work)}
            >
              <h3 className="card-title">{work.title}</h3>
              <div className="card-meta">
                <span className="card-technique">
                  {techniqueLabels[work.technique]}
                </span>
                <span
                  className="card-status"
                  style={{
                    color: statusColors[work.status] ?? "#333",
                  }}
                >
                  {workStatusLabels[work.status]}
                </span>
              </div>
              {work.description && (
                <p className="card-description">{work.description}</p>
              )}
              <time className="card-date" dateTime={work.started_at}>
                {formatDate(work.started_at)}
              </time>
            </button>
            {onDelete && (
              <button
                type="button"
                className="btn-delete"
                data-testid={`delete-work-${work.id}`}
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete(work);
                }}
                aria-label={`${work.title}を削除`}
              >
                削除
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
