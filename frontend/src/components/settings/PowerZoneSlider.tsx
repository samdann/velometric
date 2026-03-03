'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import type { PowerZone } from '@/lib/api';

const ZONE_NAMES = ['Recovery', 'Endurance', 'Tempo', 'Threshold', 'VO2 Max', 'Anaerobic', 'Neuromuscular'];
const ZONE_COLORS = ['#64748B', '#3B82F6', '#22C55E', '#EAB308', '#F97316', '#EF4444', '#DC2626'];
const MIN_GAP = 2; // minimum watts between adjacent handles

// Slider display max: 160% of FTP (Z7 is open-ended but we cap the visual)
const SLIDER_MAX_RATIO = 1.6;

interface Props {
  initialFTP: number;
  initialZones: PowerZone[];
  onSave: (ftp: number, boundaries: number[]) => Promise<void>;
}

function defaultBoundaries(ftp: number): number[] {
  return [0.55, 0.75, 0.90, 1.05, 1.20, 1.50].map(p => Math.round(p * ftp));
}

function boundariesFromZones(zones: PowerZone[], ftp: number): number[] {
  return zones.slice(0, 6).map(z => Math.round((z.max_percentage ?? 150) * ftp / 100));
}

export function PowerZoneSlider({ initialFTP, initialZones, onSave }: Props) {
  const effectiveFTP = initialFTP > 0 ? initialFTP : 250;

  const [ftp, setFtp] = useState(effectiveFTP);
  const [ftpInput, setFtpInput] = useState(String(effectiveFTP));
  const [boundaries, setBoundaries] = useState<number[]>(() => {
    if (initialZones.length >= 7 && initialFTP > 0) {
      return boundariesFromZones(initialZones, initialFTP);
    }
    return defaultBoundaries(effectiveFTP);
  });
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const trackRef = useRef<HTMLDivElement>(null);
  const boundariesRef = useRef(boundaries);
  const ftpRef = useRef(ftp);
  boundariesRef.current = boundaries;
  ftpRef.current = ftp;

  const sliderMax = Math.round(ftp * SLIDER_MAX_RATIO);
  const sliderMaxRef = useRef(sliderMax);
  sliderMaxRef.current = sliderMax;

  useEffect(() => {
    if (initialFTP > 0) {
      setFtp(initialFTP);
      setFtpInput(String(initialFTP));
      if (initialZones.length >= 7) {
        setBoundaries(boundariesFromZones(initialZones, initialFTP));
      } else {
        setBoundaries(defaultBoundaries(initialFTP));
      }
    }
  }, [initialFTP, initialZones]);

  const startDrag = useCallback((index: number) => (e: React.MouseEvent) => {
    e.preventDefault();
    const startX = e.clientX;
    const startVal = boundariesRef.current[index];

    const onMove = (ev: MouseEvent) => {
      const track = trackRef.current;
      if (!track) return;
      const width = track.getBoundingClientRect().width;
      const delta = Math.round(((ev.clientX - startX) / width) * sliderMaxRef.current);
      let val = startVal + delta;
      const lo = index > 0 ? boundariesRef.current[index - 1] + MIN_GAP : MIN_GAP;
      const hi = index < 5 ? boundariesRef.current[index + 1] - MIN_GAP : sliderMaxRef.current - MIN_GAP;
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

  const handleFTPCommit = (raw: string) => {
    const n = parseInt(raw, 10);
    if (isNaN(n) || n <= 0) return;
    const prev = ftpRef.current;
    const newSliderMax = Math.round(n * SLIDER_MAX_RATIO);
    setBoundaries(b => b.map(v => Math.round(Math.min(v * (n / prev), newSliderMax - MIN_GAP))));
    setFtp(n);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(ftp, boundaries);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } finally {
      setSaving(false);
    }
  };

  // Build zone segments
  const segments = ZONE_NAMES.map((name, i) => {
    const minWatts = i === 0 ? 0 : boundaries[i - 1];
    const maxWatts = i < 6 ? boundaries[i] : null; // Z7 is open-ended
    const leftPct = (minWatts / sliderMax) * 100;
    const widthPct = maxWatts !== null
      ? ((maxWatts - minWatts) / sliderMax) * 100
      : 100 - leftPct;
    return { name, color: ZONE_COLORS[i], leftPct, widthPct, minWatts, maxWatts };
  });

  // FTP marker position
  const ftpPct = Math.min((ftp / sliderMax) * 100, 100);

  return (
    <div className="space-y-5">
      {/* FTP input */}
      <div className="flex items-center gap-2">
        <span className="text-sm text-foreground-muted font-mono">FTP</span>
        <input
          type="number"
          value={ftpInput}
          onChange={e => setFtpInput(e.target.value)}
          onBlur={e => handleFTPCommit(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleFTPCommit(ftpInput)}
          className="w-20 rounded-md border border-border bg-background-subtle px-2 py-1 text-sm font-mono text-center focus:outline-none focus:border-border-hover"
          min={50}
          max={600}
        />
        <span className="text-sm text-foreground-muted font-mono">W</span>
      </div>

      {/* Track */}
      <div className="relative select-none" style={{ paddingBottom: '2rem' }}>
        <div ref={trackRef} className="relative h-10 rounded-xl overflow-visible">
          {/* Segments */}
          <div className="absolute inset-0 rounded-xl overflow-hidden">
            {segments.map((seg, i) => (
              <div
                key={i}
                className="absolute top-0 h-full transition-all duration-75"
                style={{
                  left: `${seg.leftPct}%`,
                  width: `${seg.widthPct}%`,
                  backgroundColor: seg.color,
                  // Z7 (last) gets a slight diagonal stripe pattern to indicate "open-ended"
                  ...(i === 6 ? {
                    backgroundImage: `repeating-linear-gradient(135deg, transparent, transparent 6px, rgba(0,0,0,0.15) 6px, rgba(0,0,0,0.15) 12px)`,
                    backgroundBlendMode: 'multiply',
                  } : {}),
                }}
              />
            ))}
          </div>

          {/* FTP marker */}
          <div
            className="absolute top-0 h-full pointer-events-none"
            style={{ left: `${ftpPct}%`, zIndex: 5 }}
          >
            <div className="absolute top-0 h-full w-px bg-white/80" />
            <div className="absolute -top-5 text-[10px] font-mono text-foreground-muted -translate-x-1/2">FTP</div>
          </div>

          {/* Boundary handles */}
          {boundaries.map((watts, i) => {
            const pct = Math.min((watts / sliderMax) * 100, 99);
            return (
              <div
                key={i}
                className="absolute top-0 h-full flex items-center justify-center"
                style={{ left: `${pct}%`, transform: 'translateX(-50%)', zIndex: 10 }}
                onMouseDown={startDrag(i)}
              >
                <div className="absolute w-px h-full bg-background/60" />
                <div className="relative w-3 h-8 rounded bg-background/90 border border-border shadow-md cursor-grab active:cursor-grabbing flex flex-col items-center justify-center gap-0.5">
                  <div className="w-1 h-px bg-foreground-muted/50" />
                  <div className="w-1 h-px bg-foreground-muted/50" />
                  <div className="w-1 h-px bg-foreground-muted/50" />
                </div>
              </div>
            );
          })}
        </div>

        {/* Watt labels */}
        <div className="relative h-8 mt-1">
          <div className="absolute text-xs font-mono text-foreground-muted" style={{ left: 0, top: 4 }}>0</div>
          {boundaries.map((watts, i) => {
            const pct = Math.min((watts / sliderMax) * 100, 99);
            return (
              <div
                key={i}
                className="absolute text-xs font-mono text-foreground-muted"
                style={{ left: `${pct}%`, transform: 'translateX(-50%)', top: 4 }}
              >
                {watts}
              </div>
            );
          })}
          <div className="absolute text-xs font-mono text-foreground-muted" style={{ right: 0, top: 4 }}>
            {sliderMax}+
          </div>
        </div>
      </div>

      {/* Zone details grid */}
      <div className="grid grid-cols-7 gap-1.5">
        {segments.map((seg, i) => (
          <div key={i} className="text-center">
            <div
              className="rounded py-1 text-xs font-semibold text-white mb-1"
              style={{ backgroundColor: seg.color }}
            >
              Z{i + 1}
            </div>
            <div className="text-[10px] text-foreground-muted leading-tight">{seg.name}</div>
            <div className="text-[10px] font-mono text-foreground mt-0.5">
              {seg.minWatts}{seg.maxWatts !== null ? `–${seg.maxWatts}` : '+'}
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
