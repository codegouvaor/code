import {
  PlaceholderPage,
  placeholderMetadata,
} from "@/components/public/content/placeholder-page";

export const metadata = placeholderMetadata("RFC");

export default function Page() {
  return <PlaceholderPage title="RFC" />;
}