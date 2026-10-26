import { notFound } from "next/navigation";

import { RaccoonGallery } from "@/components/RaccoonGallery";

export const metadata = { title: "Galeria do guaxinim · Raccoons World" };

/**
 * Development-only art gallery (RW-003, RA8). Production builds answer 404,
 * so players never see it.
 */
export default function GalleryPage() {
  if (process.env.NODE_ENV === "production") notFound();
  return <RaccoonGallery />;
}
