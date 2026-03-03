'use client';

import { useEffect, useState } from 'react';
import type { PowerZone } from '@/lib/api';

const ZONE_NAMES = ['Recovery', 'Endurance', 'Tempo', 'Threshold', 'VO2 Max', 'Anaerobic', 'Neuromuscular'];
const ZONE_COLORS = ['#64748B', '#3B82F6', '#22C55E', '#EAB308', '#F97316', '#EF4444', '#DC2626'];
// Standard Coggan percentages: boundaries between zones
const BOUNDARIES_PCT = [0.55, 0.75, 0.90, 1.05, 1.20, 1.50];

interface Props {
  initialFTP: number;
  initialZones: PowerZone[];
  onSave: (ftp: number, boundaries: number[]) => Promise<void>;
}

function computeZones(ftp: number) {
  return ZONE_NAMES.map((name, i) => {
    const minW = i === 0 ? 0 : Math.round(BOUNDARIES_PCT[i - 1] * ftp);
    const maxW = i < 6 ? Math.round(BOUNDARIES_PCT[i] * ftp) : null;
    return { name, color: ZONE_COLORS[i], minW, maxW };
  });
}

export function PowerZoneSlider({ initialFTP, initialZones, onSave }: Props) {
  const effectiveFTP = initialFTP > 0 ? initialFTP : 250;
  const [ftpInput, setFtpInput] = useState(String(effectiveFTP));
  const [ftp, setFtp] = useState(effectiveFTP);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (initialFTP > 0) {
      setFtp(initialFTP);
      setFtpInput(String(initialFTP));
    }
  }, [initialFTP]);

  const handleFtpChange = (val: string) => {
    setFtpInput(val);
    const n = parseInt(val, 10);
    if (!isNaN(n) && n > 0) setFtp(n);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const boundaries = BOUNDARIES_PCT.map(p => Math.round(p * ftp));
      await onSave(ftp, boundaries);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } finally {
      setSaving(false);
    }
  };

  const zones = computeZones(ftp);

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <span className="text-sm text-foreground-muted font-mono">FTP</span>
        <input
          type="number"
          value={ftpInput}
          onChange={e => handleFtpChange(e.target.value)}
          className="w-20 rounded-md border border-border bg-background-subtle px-2 py-1 text-sm font-mono text-center focus:outline-none focus:border-border-hover"
          min={50}
          max={600}
        />
        <span className="text-sm text-foreground-muted font-mono">W</span>
      </div>

      <div className="divide-y divide-border rounded-lg border border-border overflow-hidden">
        {zones.map((z, i) => (
          <div key={i} className="flex items-center gap-3 px-3 py-2 bg-background">
            <div className="w-2 h-2 rounded-full flex-shrink-0" style={{ backgroundColor: z.color }} />
            <span className="text-xs font-mono text-foreground-muted w-4">Z{i + 1}</span>
            <span className="text-xs text-foreground flex-1">{z.name}</span>
            <span className="text-xs font-mono text-foreground">
              {z.minW}–{z.maxW !== null ? z.maxW : '∞'} W
            </span>
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
