'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import type { HRZone } from '@/lib/api';

const ZONE_NAMES = ['Recovery', 'Endurance', 'Tempo', 'Threshold', 'Anaerobic'];
const ZONE_COLORS = ['#FCA5A5', '#F87171', '#EF4444', '#DC2626', '#B91C1C'];
const MIN_GAP = 2; // minimum bpm between adjacent handles

interface Props {
  initialMaxHR: number;
  initialZones: HRZone[];
  onSave: (maxHR: number, boundaries: number[]) => Promise<void>;
}

function defaultBoundaries(maxHR: number): number[] {
  return [0.60, 0.70, 0.80, 0.90].map(p => Math.round(p * maxHR));
}

function boundariesFromZones(zones: HRZone[], maxHR: number): number[] {
  // Each zone's max_percentage (except last) is a boundary
  return zones.slice(0, 4).map(z => Math.round((z.max_percentage ?? 90) * maxHR / 100));
}

export function HRZoneSlider({ initialMaxHR, initialZones, onSave }: Props) {
  const effectiveMaxHR = initialMaxHR > 0 ? initialMaxHR : 185;

  const [maxHR, setMaxHR] = useState(effectiveMaxHR);
  const [maxHRInput, setMaxHRInput] = useState(String(effectiveMaxHR));
  const [boundaries, setBoundaries] = useState<number[]>(() => {
    if (initialZones.length >= 5 && initialMaxHR > 0) {
      return boundariesFromZones(initialZones, initialMaxHR);
    }
    return defaultBoundaries(effectiveMaxHR);
  });
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const trackRef = useRef<HTMLDivElement>(null);
  const boundariesRef = useRef(boundaries);
  const maxHRRef = useRef(maxHR);
  boundariesRef.current = boundaries;
  maxHRRef.current = maxHR;

  // Sync when parent data loads
  useEffect(() => {
    if (initialMaxHR > 0) {
      setMaxHR(initialMaxHR);
      setMaxHRInput(String(initialMaxHR));
      if (initialZones.length >= 5) {
        setBoundaries(boundariesFromZones(initialZones, initialMaxHR));
      } else {
        setBoundaries(defaultBoundaries(initialMaxHR));
      }
    }
  }, [initialMaxHR, initialZones]);

  const startDrag = useCallback((index: number) => (e: React.MouseEvent) => {
    e.preventDefault();
    const startX = e.clientX;
    const startVal = boundariesRef.current[index];

    const onMove = (ev: MouseEvent) => {
      const track = trackRef.current;
      if (!track) return;
      const width = track.getBoundingClientRect().width;
      const delta = Math.round(((ev.clientX - startX) / width) * maxHRRef.current);
      let val = startVal + delta;
      const lo = index > 0 ? boundariesRef.current[index - 1] + MIN_GAP : MIN_GAP;
      const hi = index < 3 ? boundariesRef.current[index + 1] - MIN_GAP : maxHRRef.current - MIN_GAP;
      val = Math.max(lo, Math.min(hi, val));
      setBoundaries(prev => {
        const next = [...prev];
        next[index] = val;
        return next;
      });
    };

    const onUp = () => {
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
    };

    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
  }, []);

  const handleMaxHRCommit = (raw: string) => {
    const n = parseInt(raw, 10);
    if (isNaN(n) || n <= 0) return;
    // Scale existing boundaries proportionally
    const prev = maxHRRef.current;
    setBoundaries(b => b.map(v => Math.round(Math.min(v * (n / prev), n - MIN_GAP))));
    setMaxHR(n);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(maxHR, boundaries);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } finally {
      setSaving(false);
    }
  };

  // Build zone segments for rendering
  const segments = ZONE_NAMES.map((name, i) => {
    const minBpm = i === 0 ? 0 : boundaries[i - 1];
    const maxBpm = i < 4 ? boundaries[i] : maxHR;
    const leftPct = (minBpm / maxHR) * 100;
    const widthPct = ((maxBpm - minBpm) / maxHR) * 100;
    return { name, color: ZONE_COLORS[i], leftPct, widthPct, minBpm, maxBpm };
  });

  return (
    <div className="space-y-3">
      {/* Max HR input */}
      <div className="flex items-center gap-2">
        <span className="text-sm text-foreground-muted font-mono">Max Heart Rate</span>
        <input
          type="number"
          value={maxHRInput}
          onChange={e => setMaxHRInput(e.target.value)}
          onBlur={e => handleMaxHRCommit(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleMaxHRCommit(maxHRInput)}
          className="w-20 rounded-md border border-border bg-background-subtle px-2 py-1 text-sm font-mono text-center focus:outline-none focus:border-border-hover"
          min={100}
          max={250}
        />
        <span className="text-sm text-foreground-muted font-mono">bpm</span>
      </div>

      {/* Track */}
      <div className="select-none space-y-1">
        {/* Bar + handles in a fixed-height container so top-1/2 is the exact bar center */}
        <div ref={trackRef} className="relative h-5">
          {/* Colored bar — vertically centered */}
          <div className="absolute inset-x-0 top-1/2 -translate-y-1/2 h-2 rounded-full overflow-hidden">
            {segments.map((seg, i) => (
              <div
                key={i}
                className="absolute top-0 h-full transition-all duration-75"
                style={{ left: `${seg.leftPct}%`, width: `${seg.widthPct}%`, backgroundColor: seg.color }}
              />
            ))}
          </div>

          {/* Handles — circles exactly centered on bar */}
          {boundaries.map((bpm, i) => {
            const pct = (bpm / maxHR) * 100;
            return (
              <div
                key={i}
                className="absolute -translate-y-1/2 -translate-x-1/2"
                style={{ left: `${pct}%`, top: '10px', zIndex: 10 }}
                onMouseDown={startDrag(i)}
              >
                <div className="w-4 h-4 rounded-full bg-background border-2 border-foreground-muted shadow cursor-grab active:cursor-grabbing" />
              </div>
            );
          })}
        </div>

        {/* BPM labels */}
        <div className="relative h-4">
          <span className="absolute text-[10px] font-mono text-foreground-muted" style={{ left: 0 }}>0</span>
          {boundaries.map((bpm, i) => {
            const pct = (bpm / maxHR) * 100;
            return (
              <span
                key={i}
                className="absolute text-[10px] font-mono text-foreground-muted -translate-x-1/2"
                style={{ left: `${pct}%` }}
              >
                {bpm}
              </span>
            );
          })}
          <span className="absolute text-[10px] font-mono text-foreground-muted" style={{ right: 0 }}>{maxHR}</span>
        </div>
      </div>

      {/* Zone details grid */}
      <div className="grid grid-cols-5 gap-1.5">
        {segments.map((seg, i) => (
          <div key={i} className="flex items-center gap-1.5 px-2 py-1 rounded" style={{ backgroundColor: seg.color + '18' }}>
            <div className="w-2 h-2 rounded-full flex-shrink-0" style={{ backgroundColor: seg.color }} />
            <div className="min-w-0">
              <div className="text-[10px] text-foreground-muted leading-none truncate">{seg.name}</div>
              <div className="text-[10px] font-mono text-foreground mt-0.5">{seg.minBpm}–{seg.maxBpm}</div>
            </div>
          </div>
        ))}
      </div>

      <div className="flex justify-end">
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-4 py-2 text-sm font-medium rounded-lg bg-primary text-white hover:bg-primary-hover disabled:opacity-50 transition-colors"
        >
          {saving ? 'Saving...' : saved ? 'Saved!' : 'Save Zones'}
        </button>
      </div>
    </div>
  );
}
