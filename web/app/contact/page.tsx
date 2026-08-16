import ContactForm from "./contact-form";

export default async function ContactPage({ searchParams }: { searchParams: Promise<{ source?: string }> }) {
  const params = await searchParams;
  const source = params.source === "guest_demo" ? "guest_demo" : "marketing_site";

  return <ContactForm source={source} />;
}
