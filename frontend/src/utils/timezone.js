export const TIMEZONE_OPTIONS = [
  { value: 'Asia/Taipei', label: '台北 (Asia/Taipei)' },
  { value: 'Asia/Tokyo', label: '東京 (Asia/Tokyo)' },
  { value: 'Asia/Shanghai', label: '上海 (Asia/Shanghai)' },
  { value: 'Asia/Hong_Kong', label: '香港 (Asia/Hong_Kong)' },
  { value: 'Asia/Singapore', label: '新加坡 (Asia/Singapore)' },
  { value: 'Asia/Seoul', label: '首爾 (Asia/Seoul)' },
  { value: 'Europe/London', label: '倫敦 (Europe/London)' },
  { value: 'Europe/Paris', label: '巴黎 (Europe/Paris)' },
  { value: 'America/New_York', label: '紐約 (America/New_York)' },
  { value: 'America/Los_Angeles', label: '洛杉磯 (America/Los_Angeles)' },
  { value: 'UTC', label: '世界協調時間 (UTC)' },
]

export function timezoneLabel(tz) {
  const found = TIMEZONE_OPTIONS.find((t) => t.value === tz)
  return found ? found.label : tz
}
