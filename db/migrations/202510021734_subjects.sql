-- migrate:up
alter table studysets drop column featured_category_id;
drop table featured_categories;

CREATE TYPE subject_category AS ENUM (
    'LANG',
    'STEM',
    'SOCIAL_STUDIES',
    'LA',
    'MATH'
);

create table subjects (
    id uuid primary key default gen_random_uuid(),
    name text,
    category subject_category,
    keywords text,
    -- `keywords` is one string of space-seperated keywords
    search_vector tsvector generated always as (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(keywords, '')), 'B')
    ) stored
);
grant select on subjects to quizfreely_api;
grant insert on subjects to quizfreely_api;
grant update on subjects to quizfreely_api;
grant delete on subjects to quizfreely_api;

-- migrate:down
