-- migrate:up
create table term_progress_history (
    id uuid primary key default gen_random_uuid(),
    timestamp timestamptz not null default now(),
    term_id uuid not null references terms(id) on delete cascade,
    user_id uuid not null references auth.users(id) on delete cascade,
    term_correct_count int,
    term_incorrect_count int,
    def_correct_count int,
    def_incorrect_count int
);

-- migrate:down
