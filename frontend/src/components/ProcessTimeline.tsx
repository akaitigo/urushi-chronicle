import type { ProcessStep } from "../types/domain";
import { formatDate, stepCategoryLabels } from "../lib/labels";

interface ProcessTimelineProps {
  steps: ProcessStep[];
  workTitle: string;
  onBack: () => void;
}

export function ProcessTimeline({
  steps,
  workTitle,
  onBack,
}: ProcessTimelineProps) {
  const sorted = [...steps].sort((a, b) => a.step_order - b.step_order);

  return (
    <div className="timeline" data-testid="process-timeline">
      <div className="timeline-header">
        <button
          type="button"
          className="back-button"
          onClick={onBack}
          data-testid="back-button"
        >
          &larr; 作品一覧に戻る
        </button>
        <h2 className="timeline-title">{workTitle} - 制作工程</h2>
      </div>

      {sorted.length === 0 ? (
        <p className="timeline-empty" data-testid="timeline-empty">
          工程がまだ記録されていません。
        </p>
      ) : (
        <ol className="timeline-list">
          {sorted.map((step) => (
            <li
              key={step.id}
              className="timeline-item"
              data-testid={`timeline-step-${step.id}`}
            >
              <div className="timeline-marker" />
              <div className="timeline-content">
                <div className="step-header">
                  <span className="step-order">{step.step_order}</span>
                  <h3 className="step-name">{step.name}</h3>
                  <span className="step-category">
                    {stepCategoryLabels[step.category]}
                  </span>
                </div>

                {step.description && (
                  <p className="step-description">{step.description}</p>
                )}

                {step.notes && <p className="step-notes">{step.notes}</p>}

                {step.materials_used && step.materials_used.length > 0 && (
                  <div className="step-materials">
                    <span className="materials-label">使用材料:</span>
                    {step.materials_used.map((material) => (
                      <span key={material} className="material-tag">
                        {material}
                      </span>
                    ))}
                  </div>
                )}

                <div className="step-dates">
                  <time dateTime={step.started_at}>
                    開始: {formatDate(step.started_at)}
                  </time>
                  {step.completed_at && (
                    <time dateTime={step.completed_at}>
                      完了: {formatDate(step.completed_at)}
                    </time>
                  )}
                </div>
              </div>
            </li>
          ))}
        </ol>
      )}
    </div>
  );
}
