const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081";

export interface Activity {
  id: string;
  userId: string;
  name: string;
  sport: string;
  startTime: string;
  duration: number;
  distance: number;
  elevationGain: number;
  avgPower?: number;
  maxPower?: number;
  normalizedPower?: number;
  tss?: number;
  intensityFactor?: number;
  variabilityIndex?: number;
  avgHeartRate?: number;
  maxHeartRate?: number;
  avgCadence?: number;
  maxCadence?: number;
  avgSpeed?: number;
  maxSpeed?: number;
  createdAt: string;
}

export interface UploadResponse {
  id: string;
  message: string;
}

export interface ApiError {
  error: string;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async uploadActivity(file: File): Promise<UploadResponse> {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch(`${this.baseUrl}/api/activities`, {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Upload failed");
    }

    return response.json();
  }

  async getActivities(): Promise<Activity[]> {
    const response = await fetch(`${this.baseUrl}/api/activities`);

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch activities");
    }

    return response.json();
  }

  async getActivity(id: string): Promise<Activity> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}`);

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch activity");
    }

    return response.json();
  }

  async getActivityRecords(id: string): Promise<unknown[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}/records`);

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch records");
    }

    return response.json();
  }

  async getPowerCurve(id: string): Promise<{ durationSeconds: number; bestPower: number; avgHeartRate?: number }[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}/power-curve`);

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch power curve");
    }

    return response.json();
  }

  async checkHealth(): Promise<{ status: string; database: string }> {
    const response = await fetch(`${this.baseUrl}/health`);
    return response.json();
  }
}

export const api = new ApiClient(API_URL);
