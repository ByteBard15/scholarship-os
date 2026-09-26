-- Fictional development catalog data. Safe to run repeatedly.
INSERT INTO institutions (id, name, short_name, institution_type, country, city)
VALUES ('30000000-0000-4000-8000-000000000001', 'Example University', 'EU', 'university', 'Exampleland', 'Example City')
ON CONFLICT (id) DO NOTHING;

INSERT INTO programmes (id, institution_id, name, degree_level, field_of_study, duration_months, mode, language)
VALUES ('30000000-0000-4000-8000-000000000002', '30000000-0000-4000-8000-000000000001', 'MSc Biomedical Engineering', 'msc', 'Biomedical Engineering', 24, 'on_campus', 'English')
ON CONFLICT (id) DO NOTHING;

INSERT INTO scholarships (id, institution_id, name, provider_name, country, degree_level, scholarship_type, official_url, is_recurring)
VALUES ('30000000-0000-4000-8000-000000000003', '30000000-0000-4000-8000-000000000001', 'Example Graduate Fellowship', 'Example University', 'Exampleland', 'masters', 'full', 'https://example.test/fellowships/graduate', TRUE)
ON CONFLICT (id) DO NOTHING;
