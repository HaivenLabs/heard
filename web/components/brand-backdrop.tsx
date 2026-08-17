export function BrandBackdrop() {
  return (
    <div aria-hidden="true" className="pointer-events-none absolute inset-0 overflow-hidden">
      <div className="absolute inset-0 [background-image:linear-gradient(rgba(23,37,29,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(23,37,29,0.045)_1px,transparent_1px)] [background-size:52px_52px]" />
      <div className="absolute right-[-12rem] top-20 h-[34rem] w-[34rem] rounded-full bg-clay/16 blur-3xl" />
      <div className="absolute -left-48 top-[42rem] h-[32rem] w-[32rem] rounded-full bg-olive/12 blur-3xl" />
    </div>
  );
}
