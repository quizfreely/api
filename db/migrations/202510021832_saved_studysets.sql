-- migrate:up
create table folders (
    id uuid primary key default gen_random_uuid(),
    user_id uuid references auth.users (id) on delete cascade,
    name text not null
);

create table saved_studysets (
    id uuid primary key default gen_random_uuid(),
    studyset_id uuid references studysets (id) on delete cascade,
    user_id uuid references auth.users (id) on delete cascade,
    folder_id uuid references folders (id) on delete set null,
    timestamp timestamptz not null default now()
);

-- migrate:down
