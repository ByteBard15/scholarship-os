# Application domain

An `Application` is a concrete opportunity owned by a user and linked to a derived Application Profile. Creation validates a Master or Domain parent, then atomically creates the child profile, application, default checklist, and audit entry. Deleting an application soft-deletes its application-owned structures and linked Application Profile; it never deletes institutions, programmes, scholarships, or parent profiles.

Catalog entities (`Institution`, `Programme`, and `Scholarship`) describe reusable opportunity identity. Application-owned relational entities capture requirements, deadlines, funding, contacts, supervisors, URLs, tasks, and requirement evidence. This avoids a wide application table and supports multiple deadlines and requirements.

Application status and research status are separate state machines. Normal application status transitions are validated. A deliberate manual override is available through the service request and is audited. Readiness is an operational completeness summary, never an admission probability.
