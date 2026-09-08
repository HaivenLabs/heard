type HeardLogoProps = {
  className?: string;
};

export function HeardLogo({ className = "" }: HeardLogoProps) {
  return (
    <span className={`inline-flex items-center gap-2 text-ink ${className}`}>
      <span
        aria-hidden="true"
        className="block h-10 w-[1.35rem] shrink-0 bg-[url('/brand/heard-logo.svg')] bg-[length:100%_auto] bg-top bg-no-repeat"
      />
      <span className="font-display text-2xl font-semibold tracking-[-0.05em]">heard</span>
    </span>
  );
}
