import type { EnvironmentReading, ProcessStep, Work } from "../types/domain";

const API_BASE = "/api/v1";

interface ListResponse<T> {
  items: T[];
  total: number;
}

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json() as Promise<T>;
}

/** Fetch all works. */
export async function fetchWorks(): Promise<Work[]> {
  const data = await fetchJSON<ListResponse<Work>>(`${API_BASE}/works`);
  return data.items;
}

/** Fetch a single work by ID. */
export async function fetchWork(id: string): Promise<Work> {
  return fetchJSON<Work>(`${API_BASE}/works/${id}`);
}

/** Fetch process steps for a work. */
export async function fetchProcessSteps(
  workId: string,
): Promise<ProcessStep[]> {
  const data = await fetchJSON<ListResponse<ProcessStep>>(
    `${API_BASE}/works/${workId}/steps`,
  );
  return data.items;
}

/** Fetch environment readings by sensor ID. */
export async function fetchEnvironmentReadings(
  sensorId: string,
  limit = 100,
): Promise<EnvironmentReading[]> {
  const data = await fetchJSON<ListResponse<EnvironmentReading>>(
    `${API_BASE}/environment/readings?sensor_id=${encodeURIComponent(sensorId)}&limit=${limit}`,
  );
  return data.items;
}
