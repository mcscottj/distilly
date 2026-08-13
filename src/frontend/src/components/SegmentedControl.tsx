type Option<T extends string> = { value: T; label: string }

type SegmentedControlProps<T extends string> = {
  options: Option<T>[]
  value: T
  onChange: (value: T) => void
  disabled?: boolean
}

export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
  disabled,
}: SegmentedControlProps<T>) {
  return (
    <div className="inline-flex rounded-md border border-hairline bg-fill p-0.5">
      {options.map((opt) => {
        const active = opt.value === value
        return (
          <button
            key={opt.value}
            type="button"
            disabled={disabled}
            onClick={() => onChange(opt.value)}
            className={
              active
                ? 'rounded-[5px] bg-surface px-3 py-1 text-sm font-medium text-fg shadow-sm'
                : 'rounded-[5px] px-3 py-1 text-sm text-muted hover:text-fg'
            }
          >
            {opt.label}
          </button>
        )
      })}
    </div>
  )
}
