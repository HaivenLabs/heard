export function BrandBackdrop() {
  return (
    <div aria-hidden="true" className="pointer-events-none absolute inset-0 overflow-hidden">
      <div className="brand-grid absolute inset-0" />
      <div className="absolute right-[-12rem] top-20 h-[34rem] w-[34rem] rounded-full bg-teal/14 blur-3xl" />
      <div className="absolute -left-48 top-[42rem] h-[32rem] w-[32rem] rounded-full bg-sage/14 blur-3xl" />
    </div>
  );
}
