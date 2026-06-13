type Props = {
  label: string;
  value: string;
};

export function StatusTile({ label, value }: Props) {
  return (
    <div className="rounded-md border border-slate-200 bg-white p-4 shadow-sm">
      <p className="text-sm text-slate-500">{label}</p>
      <p className="mt-2 break-words text-xl font-semibold text-ink">{value}</p>
    </div>
  );
}

