# Editing Writing samples

The Writing Library has an Edit action on each sample. It opens a dialog prefilled with the sample's title, document type, description, comma-separated tags and full content. Save changes uses the existing owned `PATCH /api/v1/writing-samples/{sampleID}` endpoint; it updates the list without resetting the current page. Cancel and Escape close without saving; closing and repeated submission are disabled while saving. Request errors stay in the dialog.

An empty description clears it; an empty tag list removes all tags. Saving does not change application links, origin, source sample relationships, profile snapshot provenance or review status. Editing a suggested draft does not approve it. Existing backend agent restrictions still apply: an agent can update only its permitted suggested agent drafts, while the owned user interface can edit the user's Writing samples.

No migration or backend route change is needed for the editor. Deploy the frontend.

Calgary portal answer extraction on 6 November 2026 uses the verified 5 November Print Summary rather than claiming a fresh browser read. Each reusable text answer is a separate `suggested` sample. Native GPA and checkbox selections are not writing samples. Review placeholders are preserved and explicitly described; they must not be treated as completed applicant facts or approved reusable source writing. Extraction descriptions retain portal reference, capture date and local evidence hash.
