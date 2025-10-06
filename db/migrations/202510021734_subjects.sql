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
    id text primary key,
    name text,
    category subject_category
);
grant select on subjects to quizfreely_api;
grant insert on subjects to quizfreely_api;
grant update on subjects to quizfreely_api;
grant delete on subjects to quizfreely_api;

create table subject_keywords (
    id uuid primary key default gen_random_uuid(),
    keyword text,
    subject_id text references subjects (id) on delete cascade
);
grant select on subject_keywords to quizfreely_api;
grant insert on subject_keywords to quizfreely_api;
grant update on subject_keywords to quizfreely_api;
grant delete on subject_keywords to quizfreely_api;

create index subject_keywords_trgm_idx on subject_keywords using gin (keyword gin_trgm_ops);
-- migrate:down
