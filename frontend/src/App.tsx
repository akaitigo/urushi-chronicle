import { useCallback, useState } from "react";
import { EnvironmentChart } from "./components/EnvironmentChart";
import { ProcessTimeline } from "./components/ProcessTimeline";
import { WorkGallery } from "./components/WorkGallery";
import { useAsync } from "./hooks/useAsync";
import {
  fetchEnvironmentReadings,
  fetchProcessSteps,
  fetchWorks,
} from "./lib/api";
import type { Work } from "./types/domain";
import "./index.css";

export function App() {
  const [selectedWork, setSelectedWork] = useState<Work | null>(null);
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
        ) : (
          <>
            {works.loading && <div className="loading">読み込み中...</div>}
            {works.error && (
              <div className="error">
                作品の取得に失敗しました: {works.error.message}
              </div>
            )}
            {!works.loading && !works.error && works.data && (
              <WorkGallery works={works.data} onSelect={handleSelectWork} />
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
