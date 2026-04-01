import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createWork,
  deleteWork,
  fetchEnvironmentReadings,
  fetchProcessSteps,
  fetchWork,
  fetchWorks,
  updateWork,
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
      mockFetch.mockResolvedValueOnce(
        mockResponse({ items: works, total: 1 }),
      );

      const result = await fetchWorks();
      expect(mockFetch).toHaveBeenCalledWith("/api/v1/works", undefined);
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
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/works/abc",
        undefined,
      );
      expect(result).toEqual(work);
    });
  });

  describe("createWork", () => {
    it("sends POST with work data", async () => {
      const created = { id: "new-1", title: "test" };
      mockFetch.mockResolvedValueOnce(mockResponse(created));

      const result = await createWork({
        title: "test",
        description: "desc",
        technique: "makie",
        material: "wood",
      });

      expect(mockFetch).toHaveBeenCalledWith("/api/v1/works", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: "test",
          description: "desc",
          technique: "makie",
          material: "wood",
        }),
      });
      expect(result).toEqual(created);
    });

    it("throws on API error", async () => {
      mockFetch.mockResolvedValueOnce(mockResponse(null, false, 400));
      await expect(
        createWork({
          title: "",
          technique: "makie",
        }),
      ).rejects.toThrow("API error");
    });
  });

  describe("updateWork", () => {
    it("sends PUT with partial fields", async () => {
      const updated = { id: "abc", title: "new" };
      mockFetch.mockResolvedValueOnce(mockResponse(updated));

      const result = await updateWork("abc", { title: "new" });

      expect(mockFetch).toHaveBeenCalledWith("/api/v1/works/abc", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title: "new" }),
      });
      expect(result).toEqual(updated);
    });
  });

  describe("deleteWork", () => {
    it("sends DELETE request", async () => {
      mockFetch.mockResolvedValueOnce({ ok: true, status: 204 });

      await deleteWork("abc");

      expect(mockFetch).toHaveBeenCalledWith("/api/v1/works/abc", {
        method: "DELETE",
      });
    });

    it("throws on API error", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: "Not Found",
      });
      await expect(deleteWork("missing")).rejects.toThrow("API error");
    });
  });

  describe("fetchProcessSteps", () => {
    it("fetches steps for a work", async () => {
      const steps = [{ id: "s1", name: "step" }];
      mockFetch.mockResolvedValueOnce(
        mockResponse({ items: steps, total: 1 }),
      );

      const result = await fetchProcessSteps("work-1");
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/works/work-1/steps",
        undefined,
      );
      expect(result).toEqual(steps);
    });
  });

  describe("fetchEnvironmentReadings", () => {
    it("fetches readings by sensor ID", async () => {
      const readings = [
        { time: "2026-01-01T00:00:00Z", temperature: 22 },
      ];
      mockFetch.mockResolvedValueOnce(
        mockResponse({ items: readings, total: 1 }),
      );

      const result = await fetchEnvironmentReadings("sensor-1");
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/environment/readings?sensor_id=sensor-1&limit=100",
        undefined,
      );
      expect(result).toEqual(readings);
    });

    it("respects custom limit", async () => {
      mockFetch.mockResolvedValueOnce(
        mockResponse({ items: [], total: 0 }),
      );

      await fetchEnvironmentReadings("sensor-1", 50);
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/environment/readings?sensor_id=sensor-1&limit=50",
        undefined,
      );
    });
  });
});
