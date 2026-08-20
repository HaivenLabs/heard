import WalkthroughForm from "./walkthrough-form";

export default async function WalkthroughPage({ searchParams }: { searchParams: Promise<{ source?: string }> }) {
  const params = await searchParams;
  const source = params.source === "guest_demo" ? "guest_demo" : "marketing_site";

  return <WalkthroughForm source={source} />;
}
