import { describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { useAsync } from "./useAsync";

describe("useAsync", () => {
  it("starts in loading state", () => {
    const fetcher = vi.fn(() => new Promise<string>(() => {}));
    const { result } = renderHook(() => useAsync(fetcher));

    expect(result.current.loading).toBe(true);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it("resolves data on success", async () => {
    const fetcher = vi.fn(() => Promise.resolve("hello"));
    const { result } = renderHook(() => useAsync(fetcher));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBe("hello");
    expect(result.current.error).toBeNull();
  });

  it("sets error on failure", async () => {
    const fetcher = vi.fn(() => Promise.reject(new Error("fail")));
    const { result } = renderHook(() => useAsync(fetcher));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe("fail");
  });

  it("handles non-Error rejection", async () => {
    const fetcher = vi.fn(() => Promise.reject("string error"));
    const { result } = renderHook(() => useAsync(fetcher));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error?.message).toBe("string error");
  });

  it("refetches when refetch is called", async () => {
    let count = 0;
    const fetcher = vi.fn(() => {
      count += 1;
      return Promise.resolve(count);
    });

    const { result } = renderHook(() => useAsync(fetcher));

    await waitFor(() => {
      expect(result.current.data).toBe(1);
    });

    result.current.refetch();

    await waitFor(() => {
      expect(result.current.data).toBe(2);
    });
  });
});
