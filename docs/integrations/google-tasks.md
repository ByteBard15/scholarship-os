# Future Google Tasks integration

Phase 3 stores only `ExternalTaskReference`. A later reviewed import/sync flow may map:

```text
Google Task → application candidate → Application → ApplicationTask → ExternalTaskReference
```

OAuth, Google API calls, import, synchronization, conflict resolution, and background refresh are intentionally absent. Future synchronization must use typed services, preserve provider IDs, and never treat an external title as verified scholarship data.
