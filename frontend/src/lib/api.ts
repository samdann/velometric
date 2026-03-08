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

export interface FeedActivity {
  id: string;
  userName: string;
  startTime: string;
  deviceName?: string;
  location?: string;
  name: string;
  distanceKm: number;
  durationSeconds: number;
  elevationGainM: number;
  route: { lat: number; lon: number }[];
}

export interface PaginatedFeed {
  activities: FeedActivity[];
  total: number;
  page: number;
  limit: number;
}

export interface PaginatedActivities {
  activities: Activity[];
  total: number;
  page: number;
  limit: number;
}

export interface ActivityFilters {
  q?: string;
  sport?: string;
  dateFrom?: string;
  dateTo?: string;
  distMin?: number;
  distMax?: number;
  sortBy?: string;
  sortOrder?: "asc" | "desc";
}

export interface UploadResponse {
  id: string;
  message: string;
}

export interface ApiError {
  error: string;
}

export interface UserProfile {
  id: string;
  email: string;
  name: string;
  ftp?: number;
  max_hr?: number;
  weight?: number;
}

export interface HRZone {
  zone_number: number;
  name: string;
  min_percentage: number;
  max_percentage: number | null;
  color: string;
}

export interface HRZonesResponse {
  max_hr: number;
  zones: HRZone[];
}

export interface HRZoneDistributionPoint {
  zone_number: number;
  name: string;
  color: string;
  min_bpm: number;
  max_bpm: number | null;
  seconds: number;
  percentage: number;
}

export interface PowerZone {
  zone_number: number;
  name: string;
  min_percentage: number;
  max_percentage: number | null;
  color: string;
}

export interface PowerZoneDistributionPoint {
  zone_number: number;
  name: string;
  color: string;
  min_watts: number;
  max_watts: number | null;
  seconds: number;
  percentage: number;
}

export interface PowerZonesResponse {
  ftp: number;
  zones: PowerZone[];
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

  async getFeed(page = 1, limit = 25): Promise<PaginatedFeed> {
    const response = await fetch(`${this.baseUrl}/api/feed?page=${page}&limit=${limit}`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch feed");
    }
    return response.json();
  }

  async getActivities(page = 1, limit = 25, filters: ActivityFilters = {}): Promise<PaginatedActivities> {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (filters.q) params.set("q", filters.q);
    if (filters.sport) params.set("sport", filters.sport);
    if (filters.dateFrom) params.set("date_from", filters.dateFrom);
    if (filters.dateTo) params.set("date_to", filters.dateTo);
    if (filters.distMin != null) params.set("dist_min", String(filters.distMin));
    if (filters.distMax != null) params.set("dist_max", String(filters.distMax));
    if (filters.sortBy) params.set("sort_by", filters.sortBy);
    if (filters.sortOrder) params.set("sort_order", filters.sortOrder);
    const response = await fetch(`${this.baseUrl}/api/activities?${params.toString()}`);

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

  async getPowerCurve(id: string): Promise<{
    durationSeconds: number;
    bestPower: number;
    avgHeartRate?: number;
    avgSpeed?: number;       // m/s
    avgGradient?: number;    // %
    avgCadence?: number;     // rpm
    avgLRBalance?: number;   // %
    avgTorqueEffectiveness?: number; // %
    wattsPerKg?: number;
  }[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}/power-curve`);

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch power curve");
    }

    return response.json();
  }

  async getElevationProfile(id: string): Promise<{ distance: number; altitude: number; temperature?: number }[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}/elevation`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch elevation profile");
    }
    return response.json();
  }

  async getSpeedProfile(id: string): Promise<{ distance: number; speed: number; power?: number }[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}/speed`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch speed profile");
    }
    return response.json();
  }

  async getHRCadenceProfile(id: string): Promise<{ distance: number; heartRate?: number; cadence?: number }[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}/hr-cadence`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch HR/cadence profile");
    }
    return response.json();
  }

  async getLaps(id: string): Promise<{
    lapNumber: number;
    startTime: string;
    duration: number;
    distance: number;
    avgPower?: number;
    maxPower?: number;
    avgHeartRate?: number;
    maxHeartRate?: number;
    avgCadence?: number;
    avgSpeed?: number;
    maxSpeed?: number;
    ascent?: number;
    descent?: number;
    trigger?: string;
  }[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}/laps`);

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch laps");
    }

    return response.json();
  }

  async getActivityRoute(id: string): Promise<{ lat: number; lon: number; distance?: number }[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}/route`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch route");
    }
    return response.json();
  }

  async checkHealth(): Promise<{ status: string; database: string }> {
    const response = await fetch(`${this.baseUrl}/health`);
    return response.json();
  }

  async getUserProfile(): Promise<UserProfile> {
    const response = await fetch(`${this.baseUrl}/api/user/profile`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch profile");
    }
    return response.json();
  }

  async updateUserProfile(data: { name: string; email: string; weight?: number }): Promise<UserProfile> {
    const response = await fetch(`${this.baseUrl}/api/user/profile`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to update profile");
    }
    return response.json();
  }

  async deleteActivity(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/api/activities/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to delete activity");
    }
  }

  async getHRZoneDistribution(activityId: string): Promise<HRZoneDistributionPoint[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${activityId}/hr-zone-distribution`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch HR zone distribution");
    }
    return response.json();
  }

  async getPowerZoneDistribution(activityId: string): Promise<PowerZoneDistributionPoint[]> {
    const response = await fetch(`${this.baseUrl}/api/activities/${activityId}/power-zone-distribution`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch power zone distribution");
    }
    return response.json();
  }

  async getHRZones(): Promise<HRZonesResponse> {
    const response = await fetch(`${this.baseUrl}/api/user/hr-zones`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch HR zones");
    }
    return response.json();
  }

  async saveHRZones(maxHR: number, boundaries: number[]): Promise<HRZonesResponse> {
    const response = await fetch(`${this.baseUrl}/api/user/hr-zones`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ max_hr: maxHR, boundaries }),
    });
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to save HR zones");
    }
    return response.json();
  }

  async getPowerZones(): Promise<PowerZonesResponse> {
    const response = await fetch(`${this.baseUrl}/api/user/power-zones`);
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to fetch power zones");
    }
    return response.json();
  }

  async savePowerZones(ftp: number, boundaries: number[]): Promise<PowerZonesResponse> {
    const response = await fetch(`${this.baseUrl}/api/user/power-zones`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ftp, boundaries }),
    });
    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || "Failed to save power zones");
    }
    return response.json();
  }
}

export const api = new ApiClient(API_URL);
