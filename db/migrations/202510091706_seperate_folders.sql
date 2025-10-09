-- migrate:up
create table folder_studysets (
    id uuid primary key default gen_random_uuid(),
    studyset_id uuid not null references studysets (id) on delete cascade,
    user_id uuid not null references auth.users (id) on delete cascade,
    folder_id uuid not null references folders (id) on delete cascade,
    timestamp timestamptz not null default now()
    unique (studyset_id, user_id)
);

grant select on folder_studysets to quizfreely_api;
grant insert on folder_studysets to quizfreely_api;
grant update on folder_studysets to quizfreely_api;
grant delete on folder_studysets to quizfreely_api;

insert into folder_studysets (studyset_id, user_id, folder_id, timestamp)
select studyset_id, user_id, folder_id, timestamp
from saved_studysets
where studyset_id is not null and user_id is not null and folder_id is not null;

delete from saved_studysets where studyset_id is not null and
user_id is not null and folder_id is not null;

alter table saved_studysets
drop column folder_id;

-- migrate:down
