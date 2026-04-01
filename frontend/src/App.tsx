import { useCallback, useState } from "react";
import { EnvironmentChart } from "./components/EnvironmentChart";
import { ProcessTimeline } from "./components/ProcessTimeline";
import { type WorkFormData, WorkForm } from "./components/WorkForm";
import { WorkGallery } from "./components/WorkGallery";
import { useAsync } from "./hooks/useAsync";
import {
  createWork,
  deleteWork,
  fetchEnvironmentReadings,
  fetchProcessSteps,
  fetchWorks,
} from "./lib/api";
import type { Work } from "./types/domain";
import "./index.css";

export function App() {
  const [selectedWork, setSelectedWork] = useState<Work | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [sensorId, setSensorId] = useState("urushi-buro-1");
  const [activeSensorId, setActiveSensorId] = useState("urushi-buro-1");

  const works = useAsync(() => fetchWorks(), []);

  const steps = useAsync(
    () =>
      selectedWork ? fetchProcessSteps(selectedWork.id) : Promise.resolve([]),
    [selectedWork?.id],
  );

  const readings = useAsync(
    () => fetchEnvironmentReadings(activeSensorId),
    [activeSensorId],
  );

  const handleSelectWork = useCallback((work: Work) => {
    setSelectedWork(work);
  }, []);

  const handleBack = useCallback(() => {
    setSelectedWork(null);
  }, []);

  const handleLoadReadings = useCallback(() => {
    setActiveSensorId(sensorId);
  }, [sensorId]);

  const handleShowForm = useCallback(() => {
    setShowForm(true);
    setFormError(null);
  }, []);

  const handleCancelForm = useCallback(() => {
    setShowForm(false);
    setFormError(null);
  }, []);

  const handleCreateWork = useCallback(
    async (data: WorkFormData) => {
      setSubmitting(true);
      setFormError(null);
      try {
        await createWork(data);
        setShowForm(false);
        works.refetch();
      } catch (err: unknown) {
        setFormError(
          err instanceof Error ? err.message : "作品の登録に失敗しました",
        );
      } finally {
        setSubmitting(false);
      }
    },
    [works],
  );

  const handleDeleteWork = useCallback(
    async (work: Work) => {
      try {
        await deleteWork(work.id);
        works.refetch();
      } catch {
        // Silently ignore; the gallery will re-fetch and show current state
      }
    },
    [works],
  );

  return (
    <div className="app">
      <header className="app-header">
        <h1 className="app-title">urushi-chronicle</h1>
        <p className="app-subtitle">蒔絵・螺鈿制作工程デジタルアーカイブ</p>
      </header>

      <main>
        {selectedWork ? (
          <>
            {steps.loading && <div className="loading">読み込み中...</div>}
            {steps.error && (
              <div className="error">
                工程の取得に失敗しました: {steps.error.message}
              </div>
            )}
            {!steps.loading && !steps.error && steps.data && (
              <ProcessTimeline
                steps={steps.data}
                workTitle={selectedWork.title}
                onBack={handleBack}
              />
            )}
          </>
        ) : showForm ? (
          <WorkForm
            onSubmit={handleCreateWork}
            onCancel={handleCancelForm}
            submitting={submitting}
            error={formError}
          />
        ) : (
          <>
            <div className="gallery-actions" data-testid="gallery-actions">
              <button
                type="button"
                className="btn-primary"
                onClick={handleShowForm}
                data-testid="add-work-button"
              >
                作品を登録
              </button>
            </div>

            {works.loading && <div className="loading">読み込み中...</div>}
            {works.error && (
              <div className="error">
                作品の取得に失敗しました: {works.error.message}
              </div>
            )}
            {!works.loading && !works.error && works.data && (
              <WorkGallery
                works={works.data}
                onSelect={handleSelectWork}
                onDelete={handleDeleteWork}
              />
            )}
          </>
        )}

        <div className="sensor-selector" data-testid="sensor-selector">
          <label htmlFor="sensor-id-input">センサーID:</label>
          <input
            id="sensor-id-input"
            type="text"
            className="sensor-input"
            value={sensorId}
            onChange={(e) => setSensorId(e.target.value)}
          />
          <button
            type="button"
            className="sensor-button"
            onClick={handleLoadReadings}
          >
            データ取得
          </button>
        </div>

        {readings.loading && (
          <div className="loading">環境データ読み込み中...</div>
        )}
        {readings.error && (
          <div className="error">
            環境データの取得に失敗しました: {readings.error.message}
          </div>
        )}
        {!readings.loading && !readings.error && readings.data && (
          <EnvironmentChart readings={readings.data} />
        )}
      </main>
    </div>
  );
}
