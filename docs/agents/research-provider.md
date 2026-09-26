# Research provider contract

Future providers may combine web search, webpage fetching, deterministic parsing, and LLM extraction. Every implementation must satisfy `ApplicationResearcher` and return the typed `ResearchResult`: sources plus requirement, deadline, funding, contact, supervisor, URL, and general finding candidates.

The orchestration service owns run state, conflict marking, persistence, review, apply transactions, audit, and application research status. Provider implementations must not import repositories, access SQL, or mutate application/profile models. Live providers must record retrieval dates, distinguish source quality, retain supporting text, and return unknown rather than guess.
