import type { EnvironmentReading, ProcessStep, Work } from "../types/domain";

const API_BASE = "/api/v1";

interface ListResponse<T> {
  items: T[];
  total: number;
}

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json() as Promise<T>;
}

/** Send a JSON request and return the parsed response. */
async function mutateJSON<T>(
  url: string,
  method: string,
  body: unknown,
): Promise<T> {
  return fetchJSON<T>(url, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
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

/** Create a new work. */
export async function createWork(
  work: Pick<Work, "title" | "description" | "technique" | "material">,
): Promise<Work> {
  return mutateJSON<Work>(`${API_BASE}/works`, "POST", work);
}

/** Update an existing work. */
export async function updateWork(
  id: string,
  fields: Partial<
    Pick<Work, "title" | "description" | "technique" | "material" | "status">
  >,
): Promise<Work> {
  return mutateJSON<Work>(`${API_BASE}/works/${id}`, "PUT", fields);
}

/** Delete a work by ID. */
export async function deleteWork(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/works/${id}`, { method: "DELETE" });
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
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
