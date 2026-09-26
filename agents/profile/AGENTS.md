# Profile Agent Rules

## Profile hierarchy

```text
Master/Base Profile
    ↓
Domain Profile
    ↓
Application Profile
```

The profile service owns inheritance and returns a resolved `EffectiveProfile`. Agents and application code must not reconstruct lineage or apply overrides independently.

## Canonical facts

Master profile data is canonical applicant information. It is the factual source used by later workflows. Canonical records remain relational and may carry evidence from documents, imports, or URLs.

## Derived profiles

Domain and application profiles may hide inherited entries, override supported fields, change emphasis, and add profile-specific relational content. These changes must not alter the parent. Whole new entities belong directly to the derived profile; JSONB overrides are for inherited field customization or visibility.

Override order is deterministic: canonical master content, then parent overrides, then child overrides. The closest child wins. Removing an override restores the inherited value.

## AI rules

AI must never fabricate employers, dates, degrees, grades, awards, publications, certifications, skills, or research work. AI-generated wording may only rephrase supported facts. Uncertainty, confidence, and source text must remain visible during review.

## Import rules

CV extraction creates candidates. Candidates require explicit acceptance or rejection before application. Pending candidates are never applied. A possible duplicate produces a proposed merge; it must not silently overwrite conflicting canonical fields.

## Evidence

Where available, facts should retain provenance through `ProfileEvidence`. Evidence can reference a document, import session, source URL, source text, and confidence. Never expose storage filesystem paths or document contents to agents.
