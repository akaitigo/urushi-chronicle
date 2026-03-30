import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  fetchEnvironmentReadings,
  fetchProcessSteps,
  fetchWork,
  fetchWorks,
} from "./api";

const mockFetch = vi.fn();

beforeEach(() => {
  vi.stubGlobal("fetch", mockFetch);
});

afterEach(() => {
  vi.restoreAllMocks();
});

function mockResponse(data: unknown, ok = true, status = 200) {
  return {
    ok,
    status,
    statusText: ok ? "OK" : "Error",
    json: () => Promise.resolve(data),
  };
}

describe("api", () => {
  describe("fetchWorks", () => {
    it("fetches works from /api/v1/works", async () => {
      const works = [{ id: "1", title: "test" }];
      mockFetch.mockResolvedValueOnce(mockResponse({ items: works, total: 1 }));

      const result = await fetchWorks();
      expect(mockFetch).toHaveBeenCalledWith("/api/v1/works");
      expect(result).toEqual(works);
    });

    it("throws on API error", async () => {
      mockFetch.mockResolvedValueOnce(mockResponse(null, false, 500));

      await expect(fetchWorks()).rejects.toThrow("API error");
    });
  });

  describe("fetchWork", () => {
    it("fetches a single work by ID", async () => {
      const work = { id: "abc", title: "test" };
      mockFetch.mockResolvedValueOnce(mockResponse(work));

      const result = await fetchWork("abc");
      expect(mockFetch).toHaveBeenCalledWith("/api/v1/works/abc");
      expect(result).toEqual(work);
    });
  });

  describe("fetchProcessSteps", () => {
    it("fetches steps for a work", async () => {
      const steps = [{ id: "s1", name: "step" }];
      mockFetch.mockResolvedValueOnce(mockResponse({ items: steps, total: 1 }));

      const result = await fetchProcessSteps("work-1");
      expect(mockFetch).toHaveBeenCalledWith("/api/v1/works/work-1/steps");
      expect(result).toEqual(steps);
    });
  });

  describe("fetchEnvironmentReadings", () => {
    it("fetches readings by sensor ID", async () => {
      const readings = [{ time: "2026-01-01T00:00:00Z", temperature: 22 }];
      mockFetch.mockResolvedValueOnce(
        mockResponse({ items: readings, total: 1 }),
      );

      const result = await fetchEnvironmentReadings("sensor-1");
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/environment/readings?sensor_id=sensor-1&limit=100",
      );
      expect(result).toEqual(readings);
    });

    it("respects custom limit", async () => {
      mockFetch.mockResolvedValueOnce(mockResponse({ items: [], total: 0 }));

      await fetchEnvironmentReadings("sensor-1", 50);
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/environment/readings?sensor_id=sensor-1&limit=50",
      );
    });
  });
});
