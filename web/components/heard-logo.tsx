type HeardLogoProps = {
  className?: string;
};

export function HeardLogo({ className = "" }: HeardLogoProps) {
  return (
    <span aria-label="heard." className={`inline-flex items-center gap-2 text-ink ${className}`}>
      <span
        aria-hidden="true"
        className="block h-10 w-[1.2rem] shrink-0 bg-[url('/brand/heard_logo_notext.svg')] bg-contain bg-center bg-no-repeat"
      />
      <span aria-hidden="true" className="font-display text-2xl font-normal leading-none tracking-[-0.05em]">heard.</span>
    </span>
  );
}
