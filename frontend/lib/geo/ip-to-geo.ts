const REGION_PRESETS = [
  { range: [0, 49], lat: 15, lon: -75 }, // Americas
  { range: [50, 99], lat: 50, lon: 10 }, // Europe
  { range: [100, 149], lat: 25, lon: 95 }, // South Asia
  { range: [150, 199], lat: -10, lon: 130 }, // Oceania / SE Asia
  { range: [200, 255], lat: -15, lon: 25 }, // Africa
];

const clamp = (value: number, min: number, max: number) =>
  Math.max(min, Math.min(max, value));

export function approximateGeoFromIp(ip: string, seed = 0) {
  const segments = ip.split(".").map((segment) => {
    const parsed = Number(segment);
    return Number.isFinite(parsed) ? clamp(parsed, 0, 255) : 0;
  });
  while (segments.length < 4) {
    segments.push(0);
  }

  const [s0, s1, s2, s3] = segments;
  const region =
    REGION_PRESETS.find(
      (preset) => s0 >= preset.range[0] && s0 <= preset.range[1],
    ) ?? REGION_PRESETS[REGION_PRESETS.length - 1];

  const noiseSeed = (seed + s3 * 131 + s2 * 17) % 1000;
  const latJitter = (s1 / 255) * 20 - 10 + noiseSeed * 0.01;
  const lonJitter = (s2 / 255) * 25 - 12.5 + noiseSeed * 0.015;
  const lat = clamp(region.lat + latJitter, -75, 78);
  const lon = clamp(region.lon + lonJitter, -179, 179);

  return {
    latitude: lat,
    longitude: lon,
  };
}
