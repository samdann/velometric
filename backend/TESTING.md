# Backend Testing Plan

## Status legend
- ✅ Done
- 🔲 Not started

---

## Layer 1 — Unit / Service tests (no I/O, no DB)

### `internal/service/metrics_test.go` ✅
All pure computation functions covered (~60 cases).

### `internal/service/activity_test.go` ✅ (partial)
Covered: `generateActivityName`, `classifyHRZone`, `classifyPowerZone`.

**Remaining 🔲** — smoothing and downsampling logic embedded in:
- `GetElevationProfile` — 7-point moving average + 400-pt downsample
- `GetSpeedProfile` — same pattern, nullable power averaged separately
- `GetHRCadenceProfile` — same, but HR and cadence are nullable ints
- `GetRoute` — 1000-pt downsample, no smoothing

These methods call `s.repo.*` first, so they can't be tested without a fake repo.
**Required first**: define `activityRepoer` interface in service package (mirrors repository.ActivityRepository methods), inject it into ActivityService, swap tests to use a fake.

Test cases to write once interface is in place:
```
GetElevationProfile
  - empty input returns empty slice
  - fewer than 400 pts returned as-is (no downsampling)
  - more than 400 pts downsampled to ≤401 (400 + last point)
  - last point always preserved
  - smoothing reduces a spike: isolated outlier altitude averaged down

GetSpeedProfile
  - power nil in all window points → SpeedPoint.Power is nil
  - power present in some points → averaged correctly
  - downsample same rules as elevation

GetHRCadenceProfile
  - all HR nil → HeartRate nil in output
  - all cadence nil → Cadence nil in output
  - mixed nil/non-nil averaged over non-nil only

GetRoute
  - fewer than 1000 pts returned as-is
  - more than 1000 pts downsampled; last point preserved
```

### `internal/service/strava_test.go` ✅
Covered: `findMatch`, `findCandidates`, `mapStravaType`.

### `internal/service/user_test.go` ✅
Covered: `buildHRZones`, `buildPowerZones`.

### `internal/service/batchimport_test.go` 🔲

`BatchImportService` runs a goroutine and tracks job state in a `sync.RWMutex` map.
The pure / lockable parts to test without touching disk or DB:

```
GetJob
  - unknown ID returns (nil, false)
  - after StartImport is called, ID is present and returns (job, true)
  - job snapshot is a copy: mutating returned job does not affect internal state

Job status transitions
  - new job starts with status "running"
  - completed job has status "completed"
  (requires a fake activityRepo and a temp directory with .fit files)

importOne classification (private — test via StartImport)
  - file with wrong sport type (filtered out) → outcome = skipped
  - duplicate FIT → outcome = duplicate
  - valid FIT → outcome = imported
  - unreadable file → outcome = errored
```

**Required**: fake `activityRepo` interface, temp dir with fixture `.fit` files.

---

## Layer 2 — Handler tests (httptest, fake services)

### `internal/handler/activity_test.go` ✅ (partial)
Covered: `GetActivity`, `DeleteActivity`, `GetLaps`, `ListActivities`, `GetElevationProfile`, `GetPowerCurve`, `CreateActivity`.

**Remaining 🔲**:

```
GetActivityRecords
  - 400 on bad UUID
  - 200 with empty array when no records
  - 200 with correct count

GetSpeedProfile / GetHRCadenceProfile / GetRoute
  - 400 on bad UUID
  - 200 with data
  - 500 on service error
  (mirror GetElevationProfile tests — table-driven to avoid repetition)

GetPowerZoneDistribution
  - 200 empty array when userService.GetPowerZones returns ftp=0
  - 200 empty array when zones slice is empty
  - 200 with distribution when ftp > 0 and zones present
  - activityService.ComputePowerZoneDistribution error → 500

GetHRZoneDistribution
  - same cases as power zone distribution but for maxHR

GetFeed
  - 200 with paged response
  - default page=1 limit=25 when not specified
  - invalid limit falls back to 25
  - no DB → 200 empty list
```

### `internal/handler/user_test.go` 🔲

```
GetProfile
  - 200 with user JSON
  - 503 when no DB
  - 500 on service error

UpdateProfile
  - 400 when name or email missing in body
  - 400 on invalid JSON
  - 200 with updated user
  - 503 when no DB

GetHRZones / GetPowerZones
  - 200 with correct shape { max_hr, zones } / { ftp, zones }
  - 503 when no DB
  - 500 on service error

SaveHRZones
  - 400 on invalid JSON
  - 400 when service returns validation error (wrong boundary count, maxHR <= 0)
  - 200 with zone list

SavePowerZones
  - same as SaveHRZones
```

### `internal/handler/health_test.go` 🔲

```
Health
  - 200 {"status":"ok"} when DB available and healthy
  - 503 when DB nil
  - 503 when DB ping fails (fake DB that returns error)
```

### `internal/handler/internal_test.go` 🔲

```
StartBatchImport
  - 400 on invalid JSON body
  - 400 on invalid RFC3339 "from" / "to"
  - 202 Accepted with job ID when valid
  - 503 when no DB

GetBatchImportStatus
  - 400 on invalid job UUID
  - 404 when job ID unknown
  - 200 with job status when found
```

---

## Layer 3 — Integration tests (real PostgreSQL)

These require a running database. Use **testcontainers-go** to spin up a disposable
Postgres instance per test run, or point at a dedicated test DB.

### Setup pattern

```go
// testmain_test.go (in each repository package)
func TestMain(m *testing.M) {
    pool, cleanup := startTestDB()   // testcontainers or env var TEST_DATABASE_URL
    defer cleanup()
    testPool = pool
    os.Exit(m.Run())
}
```

Run migrations before tests:
```go
goose.Up(db, "../../migrations")
```

### `internal/repository/activity_test.go` 🔲

```
Create
  - inserts activity and returns UUID
  - duplicate (same user+time+sport+distance+duration) returns ErrDuplicateActivity

GetByID
  - returns ErrActivityNotFound for unknown ID
  - returns full activity for known ID

ListByUserIDPaginated
  - returns correct page and total count
  - filter by sport narrows results
  - filter by date range works
  - sort by distance desc
  - empty result returns empty slice (not nil)

InsertRecords
  - bulk-inserts all records
  - NaN float values rejected at DB level (constraint) or sanitized before insert

GetRecords
  - NaN/Inf floats sanitized to nil (sanitizeFloat)
  - returns empty slice (not nil) when no records

GetElevationProfile
  - only records with non-null altitude and distance returned
  - distance converted from meters to km

GetSpeedProfile
  - only records with non-null speed returned
  - speed converted from m/s to km/h

GetRoute
  - only records with lat+lon returned
  - distance converted to km

InsertLaps / GetLaps
  - duplicate lap_number on same activity errors
  - gradient clamped: |gradient| > 200 → nil (clampGradient)

InsertPowerCurve / GetPowerCurve
  - upserts correctly (conflict on activity_id, duration_seconds)
  - all optional fields stored and retrieved

GetHRTimeSeries / GetPowerTimeSeries
  - returns only non-null values in timestamp order

DeleteActivity
  - cascades to records, laps, power_curve, events
  - returns ErrActivityNotFound for unknown ID

UpdateActivityLocation
  - sets location on existing activity

FindByTimeRange
  - returns only activities within window

GetFeedActivities
  - returns correct pagination
  - mini-route points are downsampled
```

### `internal/repository/user_test.go` 🔲

```
GetFirst
  - returns demo user after seed
  - error when table empty

UpdateProfile / UpdateMaxHR / UpdateFTP
  - values persisted and readable after update

GetHRZones / UpsertHRZones
  - empty zones returns empty slice (not nil)
  - upsert replaces existing zones for user

GetPowerZones / UpsertPowerZones
  - same as HR zones
```

### `internal/repository/strava_test.go` 🔲

```
Upsert
  - inserts new strava activity
  - second upsert with same strava_id updates existing row (no duplicate)

GetByUserID
  - returns all activities for user
  - empty result returns empty slice (not nil)
```

---

## Tooling

| Tool | Purpose |
|---|---|
| `testing` stdlib | All unit and handler tests (already in use) |
| `net/http/httptest` | Handler tests (already in use) |
| `testcontainers-go` | Spin up disposable Postgres for integration tests |
| `github.com/pressly/goose/v3` | Run migrations inside test setup |

Install when starting integration tests:
```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/pressly/goose/v3
```

---

## Priority order

1. 🔲 `handler/user_test.go` — no new infrastructure needed, just more fakes
2. 🔲 `handler/health_test.go` — trivial, one file
3. 🔲 `handler/internal_test.go` — needs fake `batchImportServicer`
4. 🔲 `service/activity_test.go` (smoothing) — needs `activityRepoer` interface in service
5. 🔲 `service/batchimport_test.go` — needs fake repo + temp dir
6. 🔲 Integration tests — largest effort, needs testcontainers setup
