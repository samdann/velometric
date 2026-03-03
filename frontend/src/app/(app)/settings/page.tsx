'use client';

import { useEffect, useState } from 'react';
import { PageHeader } from '@/components/layout';
import { HRZoneSlider } from '@/components/settings/HRZoneSlider';
import { PowerZoneSlider } from '@/components/settings/PowerZoneSlider';
import { api } from '@/lib/api';
import type { UserProfile, HRZone, PowerZone } from '@/lib/api';

function SectionCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-background p-6 space-y-4">
      <h2 className="text-base font-semibold text-foreground">{title}</h2>
      {children}
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-sm text-foreground-muted">{label}</label>
      {children}
    </div>
  );
}

const inputClass = "rounded-md border border-border bg-background-subtle px-3 py-2 text-sm text-foreground focus:outline-none focus:border-border-hover w-full";

export default function SettingsPage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [hrData, setHRData] = useState<{ max_hr: number; zones: HRZone[] }>({ max_hr: 0, zones: [] });
  const [powerData, setPowerData] = useState<{ ftp: number; zones: PowerZone[] }>({ ftp: 0, zones: [] });
  const [loading, setLoading] = useState(true);

  // Profile form state
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [weight, setWeight] = useState('');
  const [profileSaving, setProfileSaving] = useState(false);
  const [profileSaved, setProfileSaved] = useState(false);

  useEffect(() => {
    Promise.all([api.getUserProfile(), api.getHRZones(), api.getPowerZones()])
      .then(([p, hr, power]) => {
        setProfile(p);
        setName(p.name);
        setEmail(p.email);
        setWeight(p.weight != null ? String(p.weight) : '');
        setHRData(hr);
        setPowerData(power);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const handleProfileSave = async () => {
    setProfileSaving(true);
    try {
      const w = weight !== '' ? parseFloat(weight) : undefined;
      const updated = await api.updateUserProfile({ name, email, weight: w });
      setProfile(updated);
      setProfileSaved(true);
      setTimeout(() => setProfileSaved(false), 2000);
    } catch (e) {
      console.error(e);
    } finally {
      setProfileSaving(false);
    }
  };

  const handleHRSave = async (maxHR: number, boundaries: number[]) => {
    const updated = await api.saveHRZones(maxHR, boundaries);
    setHRData(updated);
  };

  const handlePowerSave = async (ftp: number, boundaries: number[]) => {
    const updated = await api.savePowerZones(ftp, boundaries);
    setPowerData(updated);
  };

  if (loading) {
    return (
      <div>
        <PageHeader title="Settings" description="Configure your profile and training zones" />
        <div className="p-6">
          <div className="rounded-lg border border-border bg-background-subtle p-8 text-center">
            <p className="text-foreground-muted text-sm">Loading...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div>
      <PageHeader title="Settings" description="Configure your profile and training zones" />
      <div className="p-6 space-y-6 max-w-4xl">

        {/* User Profile */}
        <SectionCard title="User Profile">
          <div className="grid grid-cols-2 gap-4">
            <Field label="Full Name">
              <input className={inputClass} value={name} onChange={e => setName(e.target.value)} />
            </Field>
            <Field label="Email">
              <input className={inputClass} type="email" value={email} onChange={e => setEmail(e.target.value)} />
            </Field>
            <Field label="Weight (kg)">
              <input className={inputClass} type="number" value={weight} onChange={e => setWeight(e.target.value)} placeholder="Optional" step="0.1" min="30" max="200" />
            </Field>
          </div>
          <div className="flex justify-end">
            <button
              onClick={handleProfileSave}
              disabled={profileSaving}
              className="px-4 py-2 text-sm font-medium rounded-lg bg-primary text-white hover:bg-primary-hover disabled:opacity-50 transition-colors"
            >
              {profileSaving ? 'Saving...' : profileSaved ? 'Saved!' : 'Save Profile'}
            </button>
          </div>
        </SectionCard>

        {/* Heart Rate Zones */}
        <SectionCard title="Heart Rate Zones">
          <p className="text-xs text-foreground-muted">
            Drag the handles to set zone boundaries. The right edge is your max heart rate.
          </p>
          <HRZoneSlider
            initialMaxHR={hrData.max_hr}
            initialZones={hrData.zones}
            onSave={handleHRSave}
          />
        </SectionCard>

        {/* Power Zones */}
        <SectionCard title="Power Zones">
          <p className="text-xs text-foreground-muted">
            Drag the handles to set zone boundaries. The white marker shows your FTP. Zone 7 is open-ended.
          </p>
          <PowerZoneSlider
            initialFTP={powerData.ftp}
            initialZones={powerData.zones}
            onSave={handlePowerSave}
          />
        </SectionCard>

      </div>
    </div>
  );
}
