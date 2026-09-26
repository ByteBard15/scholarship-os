# Profile import workflow

```text
Uploaded document
    ↓
Deterministic text extraction
    ↓
Typed structured extraction
    ↓
Import candidates
    ↓
Explicit review
    ↓
Transactional application + evidence
```

Local development stores files under `UPLOAD_DIR` using generated UUID keys. Public DTOs never expose storage keys or filesystem paths. Uploads are size-limited and restricted to PDF, DOCX, and TXT with matching extensions and basic magic-byte validation.

TXT and DOCX text extraction are supported synchronously without external services. PDF files may be stored, but PDF text extraction intentionally returns a failed import until a reliable parser is introduced; OCR is not performed. The deterministic development extractor recognizes explicitly labeled lines such as `SKILLS: Go, PostgreSQL`. The `ProfileExtractor` interface is ready for a future controlled AI provider, but the application requires no AI credentials.

Extraction never writes canonical facts. It creates candidates with confidence and source text. Every candidate must be accepted, rejected, or marked for merge review before apply. Merge review reports conflicting fields and does not overwrite an existing entity. Applying accepted candidates and their evidence occurs in one database transaction.
