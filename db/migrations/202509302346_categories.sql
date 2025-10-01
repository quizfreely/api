-- migrate:up
create table featured_categories (
    id uuid primary key default gen_random_uuid(),
    title text
);
grant select on featured_categories to quizfreely_api;
grant insert on featured_categories to quizfreely_api;
grant update on featured_categories to quizfreely_api;
grant delete on featured_categories to quizfreely_api;

alter table studysets
add column featured_category_id uuid references featured_categories (id) on delete set null;
-- migrate:down
