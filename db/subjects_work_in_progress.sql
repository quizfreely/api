-- migrate:up
drop table featured_categories;
CREATE TYPE subject_category AS ENUM (
    'LANG',
    'SCI',
    'HISTORY',
    'LA',
    'MATH'
);

create table subjects (
    id uuid primary key default gen_random_uuid(),
    name text,
    category subject_category,
    icon text,
    keywords text,
    -- `keywords` is one string of space-seperated keywords
    search_vector tsvector generated always as (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(keywords, '')), 'B')
    ) stored;
)
grant select on subjects to quizfreely_api;
grant insert on subjects to quizfreely_api;
grant update on subjects to quizfreely_api;
grant delete on subjects to quizfreely_api;

insert into subjects (name, category, keywords) values
    ('Spanish', 'LANG', 'español espanol espanyol'),
    ('French', 'LANG', 'français'),
    ('Arabic', 'LANG', 'العربية'),
    ('Latin', 'LANG', ''),
    ('Catalan', 'LANG', 'català'),
    ('German', 'LANG', 'deutsch'),
    ('Russian', 'LANG', 'русский'),
    ('Chinese', 'LANG', '中文 mandarin cantonese'),
    ('Vietnamese', 'LANG', 'tiếng việt tieng viet'),
    ('Lojban', 'LANG', ''),
    ('Dutch', 'LANG', ''),
    ('Italian', 'LANG', 'italiano'),
    ('Biology', 'SCI', 'bio'),
    ('Chemistry', 'SCI', 'chem'),
    ('Physics', 'SCI', ''),
    ('Computer Science', NULL, 'compsci comp sci'),
    ('US History', 'HISTORY', 'apush'),
    ('World History', 'HISTORY', 'apwh'),
    ('English', 'LA', 'ela'),
    ('English Language & Composition', 'LA', 'ap lang'),
    ('English Literature & Composition', 'LA', 'ap lit'),
    ('')
-- migrate:down
