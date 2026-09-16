import type { ContactRecipient } from "@/components/public/contact/contact-form";

/**
 * Recipients of the contact form of the Banque centrale d'Astoria.
 *
 * Structure only: every label is resolved from the message catalogs
 * (`pages.contact.recipients.<id>`). The `presidence` entry is a real
 * destination (the form renders it as its first, hard-coded option and still
 * needs it in the array to resolve the `mailto:` target), the others fill the
 * `optgroup` of the form.
 */
const recipients = [
  { id: "presidence", email: "presidence@bca.astoria-gouv.org" },
  { id: "communications", email: "communication@bca.astoria-gouv.org" },
  { id: "presse", email: "presse@bca.astoria-gouv.org" },
  { id: "monnaie", email: "monnaie@bca.astoria-gouv.org" },
  { id: "etablissements", email: "etablissements@bca.astoria-gouv.org" },
] as const;

export type ContactRecipientRef = (typeof recipients)[number];

/** The recipients of the contact form, as structural references. */
export function contactRecipients(): ReadonlyArray<ContactRecipientRef> {
  return recipients;
}

/** Builds the localized recipients array consumed by `ContactForm`. */
export function localizedContactRecipients(
  labelOf: (id: string) => string
): ReadonlyArray<ContactRecipient> {
  return recipients.map((recipient) => ({
    id: recipient.id,
    label: labelOf(recipient.id),
    email: recipient.email,
  }));
}